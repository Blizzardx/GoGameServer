package Network

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proc"
)

type ServerObject struct {
	peer       cellnet.Peer
	dispatcher *proc.MessageDispatcher
	Session    cellnet.Session
}

func (self *ServerObject) Stop() {
	self.peer.Stop()
}
