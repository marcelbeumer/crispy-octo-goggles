package now

import (
	"sync"
	"time"
)

type Stub struct {
	Time   time.Time
	Step   time.Duration
	Frozen bool
}

func (n *Stub) Inc() {
	n.Time = n.Time.Add(n.Step)
}

func (n *Stub) Now() time.Time {
	if !n.Frozen {
		n.Inc()
	}
	return n.Time
}

func NewStub() *Stub {
	return &Stub{Time: time.UnixMilli(0), Step: time.Second}
}

var stub *Stub
var stubLock sync.Mutex // for concurrent tests

func SetupStub() *Stub {
	stubLock.Lock()
	stub = NewStub()
	return stub
}

func ClearStub() {
	stub = nil
	stubLock.Unlock()
}

func Now() time.Time {
	if stub != nil {
		return stub.Now()
	} else {
		return time.Now()
	}
}
