package bcaster

import "github.com/hashicorp/memberlist"

type Broadcast struct {
	Msg    []byte
	Notify chan<- struct{}
}

func (b *Broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *Broadcast) Message() []byte {
	return b.Msg
}

func (b *Broadcast) Finished() {
	if b.Notify != nil {
		close(b.Notify)
	}
}
