package peering

import (
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/memberlist"
	"gopkg.in/resty.v1"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-sockaddr"
)

// WebTransportConfig is used to configure a net transport.
type WebTransportConfig struct {
	// BindAddrs is a list of addresses to bind to for both TCP and UDP
	// communications.
	BindAddrs []string

	// BindPort is the port to listen on, for each address above.
	BindPort int
}

// WebTransport is a Transport implementation that uses connectionless UDP for
// packet operations, and ad-hoc TCP connections for stream operations.
type WebTransport struct {
	config       *WebTransportConfig
	packetCh     chan *memberlist.Packet
	streamCh     chan net.Conn
	wg           sync.WaitGroup
	tcpListeners []*net.TCPListener
	shutdown     int32
	wsrv         *http.Server
}

// NewWebTransport returns a web transport with the given configuration. On
// success all the network listeners will be created and listening.
func NewWebTransport(config *WebTransportConfig) (*WebTransport, error) {
	// If we reject the empty list outright we can assume that there's at
	// least one listener of each type later during operation.
	if len(config.BindAddrs) == 0 {
		return nil, fmt.Errorf("at least one bind address is required")
	}

	// Build out the new transport.
	var ok bool
	t := WebTransport{
		config:   config,
		packetCh: make(chan *memberlist.Packet),
		streamCh: make(chan net.Conn),
	}

	// Clean up listeners if there's an error.
	defer func() {
		if !ok {
			t.Shutdown()
		}
	}()

	// Build all the TCP and UDP listeners.
	port := config.BindPort
	for _, addr := range config.BindAddrs {
		ip := net.ParseIP(addr)

		tcpAddr := &net.TCPAddr{IP: ip, Port: port}
		tcpLn, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to start TCP listener on %q port %d: %v", addr, port, err)
		}
		t.tcpListeners = append(t.tcpListeners, tcpLn)

		// If the config port given was zero, use the first TCP listener
		// to pick an available port and then apply that to everything
		// else.
		if port == 0 {
			port = tcpLn.Addr().(*net.TCPAddr).Port
		}
	}

	// Fire them up now that we've been able to create them all.
	for i := 0; i < len(config.BindAddrs); i++ {
		t.wg.Add(1)
		go t.tcpListen(t.tcpListeners[i])
		go t.webListen()
	}

	ok = true
	return &t, nil
}

func (t *WebTransport) resetShutdownFlag() {
	if s := atomic.LoadInt32(&t.shutdown); s == 1 {
		fmt.Println("shutdown off")
		atomic.StoreInt32(&t.shutdown, 0)
	}
}

// GetAutoBindPort returns the bind port that was automatically given by the
// kernel, if a bind port of 0 was given.
func (t *WebTransport) GetAutoBindPort() int {
	// We made sure there's at least one TCP listener, and that one's
	// port was applied to all the others for the dynamic bind case.
	return t.tcpListeners[0].Addr().(*net.TCPAddr).Port
}

// See Transport.
func (t *WebTransport) FinalAdvertiseAddr(ip string, port int) (net.IP, int, error) {
	var advertiseAddr net.IP
	var advertisePort int
	if ip != "" {
		// If they've supplied an address, use that.
		advertiseAddr = net.ParseIP(ip)
		if advertiseAddr == nil {
			return nil, 0, fmt.Errorf("failed to parse advertise address %q", ip)
		}

		// Ensure IPv4 conversion if necessary.
		if ip4 := advertiseAddr.To4(); ip4 != nil {
			advertiseAddr = ip4
		}
		advertisePort = port
	} else {
		if t.config.BindAddrs[0] == "0.0.0.0" {
			// Otherwise, if we're not bound to a specific IP, let's
			// use a suitable private IP address.
			var err error
			ip, err = sockaddr.GetPrivateIP()
			if err != nil {
				return nil, 0, fmt.Errorf("failed to get interface addresses: %v", err)
			}
			if ip == "" {
				return nil, 0, fmt.Errorf("no private IP address found, and explicit IP not provided")
			}

			advertiseAddr = net.ParseIP(ip)
			if advertiseAddr == nil {
				return nil, 0, fmt.Errorf("failed to parse advertise address: %q", ip)
			}
		} else {
			// Use the IP that we're bound to, based on the first
			// TCP listener, which we already ensure is there.
			advertiseAddr = t.tcpListeners[0].Addr().(*net.TCPAddr).IP
		}

		// Use the port we are bound to.
		advertisePort = t.GetAutoBindPort()
	}

	return advertiseAddr, advertisePort, nil
}

// See Transport.
func (t *WebTransport) WriteTo(b []byte, addr string) (time.Time, error) {
	parts := strings.Split(addr, ":")
	host := parts[0]
	pStr := parts[1]

	prt, err := strconv.Atoi(pStr)
	if err != nil {
		log.WithError(err).Error("port is not an int")
		return time.Time{}, err
	}

	sEnc := base64.StdEncoding.EncodeToString(b)

	prt = prt + 1
	url := fmt.Sprintf("http://%s:%v/fed", host, prt)
	log.Debug("pinging ", url)
	rcl := resty.New()
	resp, err := rcl.R().
		SetHeader("X-Reply", strconv.Itoa(t.config.BindPort)).
		SetBody(sEnc).
		Post(url)

	if err != nil {
		log.WithError(err).Error("ping failed (api error)")
		return time.Time{}, err
	}

	if resp.StatusCode() != 200 {
		log.WithError(err).Error("ping failed")
		return time.Time{}, err
	}

	return time.Now(), err
}

// See Transport.
func (t *WebTransport) PacketCh() <-chan *memberlist.Packet {
	return t.packetCh
}

// See Transport.
func (t *WebTransport) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	dialer := net.Dialer{Timeout: timeout}
	return dialer.Dial("tcp", addr)
}

// See Transport.
func (t *WebTransport) StreamCh() <-chan net.Conn {
	return t.streamCh
}

// See Transport.
func (t *WebTransport) Shutdown() error {
	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&t.shutdown, 1)

	// Rip through all the connections and shut them down.
	for _, conn := range t.tcpListeners {
		conn.Close()
	}

	// Block until all the listener threads have died.
	//t.wsrv.Shutdown(context.Background())
	t.wg.Wait()
	return nil
}

// tcpListen is a long running goroutine that accepts incoming TCP connections
// and hands them off to the stream channel.
func (t *WebTransport) tcpListen(tcpLn *net.TCPListener) {
	defer t.wg.Done()
	for {
		conn, err := tcpLn.AcceptTCP()
		if err != nil {
			if s := atomic.LoadInt32(&t.shutdown); s == 1 {
				break
			}

			log.Printf("[ERR] memberlist: Error accepting TCP connection: %v", err)
			continue
		}

		t.streamCh <- conn
	}
}

func (t *WebTransport) pingHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.WithError(err).Error("failed read body")
		w.WriteHeader(500)
		return
	}

	sDec, err := base64.StdEncoding.DecodeString(string(body))

	if err != nil {
		log.WithError(err).Error("failed decode body")
		w.WriteHeader(500)
		return
	}

	addr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	if err != nil {
		log.WithError(err).Error("failed resolve remote addr")
		w.WriteHeader(500)
		return
	}

	replyPrtStr := r.Header.Get("X-Reply")
	replyPrt, _ := strconv.Atoi(replyPrtStr)
	addr.Port = replyPrt

	log.Debug("ping received from ", addr)
	ts := time.Now()

	t.packetCh <- &memberlist.Packet{
		Buf:       sDec,
		From:      addr,
		Timestamp: ts,
	}

	w.WriteHeader(http.StatusOK)
}

// udpListen is a long running goroutine that accepts incoming UDP packets and
// hands them off to the packet channel.
func (t *WebTransport) webListen() {
	r := mux.NewRouter()
	h := fmt.Sprintf("%s:%v", t.config.BindAddrs[0], t.config.BindPort+1)

	r.HandleFunc("/fed", t.pingHandler)
	srv := &http.Server{
		Handler: r,
		Addr:    h,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 3 * time.Second,
		ReadTimeout:  3 * time.Second,
	}

	log.Info("starting web transport listener on ", h)
	t.wsrv = srv
	err := srv.ListenAndServe()
	if err != nil {
		log.Error(err)
	}

}
