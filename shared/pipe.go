package utils

import (
	"sync"
)

type Pipe struct {
	list      []interface{}
	listGuard sync.Mutex
	listCond  *sync.Cond
}

func (self *Pipe) Add(msg interface{}) {
	self.listGuard.Lock()
	self.list = append(self.list, msg)
	self.listGuard.Unlock()

	self.listCond.Signal()
}

func (self *Pipe) Reset() {
	self.list = self.list[0:0]
}

func (self *Pipe) Pick(retList *[]interface{}) (exit bool) {

	self.listGuard.Lock()
	for len(self.list) == 0 {
		self.listCond.Wait()
	}

	for _, data := range self.list {

		if data == nil {
			exit = true
			break
		} else {
			*retList = append(*retList, data)
		}
	}

	self.Reset()
	self.listGuard.Unlock()
	return
}

func NewPipe() *Pipe {
	self := &Pipe{}
	self.listCond = sync.NewCond(&self.listGuard)

	return self
}
