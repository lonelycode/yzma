package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lonelycode/yzma/logger"
	"github.com/lonelycode/yzma/oplog"
	"github.com/lonelycode/yzma/server"
	"io/ioutil"
	"net/http"
	"strings"
)

var log = logger.GetLogger("api")

type Payload struct {
	Status string
	Error  string
	Data   interface{}
}

type JoinReq struct {
	Peers []string
}

type WebAPI struct {
	server *server.Server
	mux    *mux.Router
	op     *oplog.Handler
	cfg    *APICfg
}

func (a *WebAPI) Start(srv *server.Server, cfg *APICfg) {
	a.server = srv
	a.mux = mux.NewRouter()
	a.initEndpoints(a.mux, a)

	log.Info("API listening on ", cfg.Bind)
	err := http.ListenAndServe(cfg.Bind, a.mux)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *WebAPI) ClusterJoin(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		a.wErr(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	var obj JoinReq
	err = json.Unmarshal(b, &obj)
	if err != nil {
		a.wErr(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = a.server.Join(obj.Peers)
	if err != nil {
		a.wErr(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	a.wOk(w, r, fmt.Sprintf("joined %s", strings.Join(obj.Peers, ",")), http.StatusOK)
}

func (a *WebAPI) ClusterLeave(w http.ResponseWriter, r *http.Request) {
	err := a.server.Leave()
	if err != nil {
		a.wErr(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	a.wOk(w, r, "leave ok", http.StatusOK)
}

func (a *WebAPI) AddObject(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	k, ok := v["key"]
	if !ok {
		a.wErr(w, r, "key required", http.StatusBadRequest)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		a.wErr(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	//var obj interface{}
	//err = json.Unmarshal(b, &obj)
	//if err != nil {
	//	a.wErr(w, r, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	a.server.Add(k, b)
	a.wOk(w, r, fmt.Sprintf("added %s", k), http.StatusOK)
}

func (a *WebAPI) RemObject(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	k, ok := v["key"]
	if !ok {
		a.wErr(w, r, "key required", http.StatusBadRequest)
		return
	}

	a.server.Remove(k)
	a.wOk(w, r, fmt.Sprintf("deleted %s", k), http.StatusOK)
}

func (a *WebAPI) LoadObject(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	k, ok := v["key"]
	if !ok {
		a.wErr(w, r, "key required", http.StatusBadRequest)
		return
	}

	dat, ok := a.server.Load(k)
	if !ok {
		a.wErr(w, r, "not found", http.StatusNotFound)
		return
	}

	a.wOk(w, r, dat.Extract(), http.StatusOK)
}

func (a *WebAPI) wOk(w http.ResponseWriter, r *http.Request, msg interface{}, code int) {
	pl := &Payload{
		Status: "ok",
		Data:   msg,
	}

	a.writeToClient(w, r, pl, code)
}

func (a *WebAPI) wErr(w http.ResponseWriter, r *http.Request, errMsg string, errCode int) {
	pl := &Payload{
		Status: "error",
		Error:  errMsg,
	}

	a.writeToClient(w, r, pl, errCode)
}

func (a *WebAPI) writeToClient(w http.ResponseWriter, r *http.Request, data *Payload, code int) {
	asJson, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error("write to client failed: ", err, " payload was: ", data)
		return
	}

	w.WriteHeader(code)
	w.Write(asJson)
	return
}
