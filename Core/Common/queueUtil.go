package Common

import (
	"sync"
)

type SyncQueue struct {
	list  []interface{}
	mutix sync.Mutex
}

func (self *SyncQueue) Offer(msg interface{}) {
	self.syncDoSth(func() {
		self.list = append(self.list, msg)
	})
}
func (self *SyncQueue) Poll() interface{} {
	if self.Length() == 0 {
		return nil
	}

	var elem interface{} = nil
	self.syncDoSth(func() {
		elem = self.list[0]
		self.list = self.list[1:]
	})
	return elem
}
func (self *SyncQueue) Length() int {
	length := 0
	self.syncDoSth(func() {
		length = len(self.list)
	})
	return length
}
func (self *SyncQueue) Clear() {
	self.syncDoSth(func() {
		self.list = self.list[0:0]
	})
}
func (self *SyncQueue) syncDoSth(f func()) {
	self.mutix.Lock()
	defer self.mutix.Unlock()
	SafeCall(f)
}
func NewSyncQueue() *SyncQueue {

	return &SyncQueue{}
}
