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

var stub = NewStub()
var stubLock sync.Mutex // for concurrent tests
var useStub = false

func EnableStub() {
	stubLock.Lock()
	useStub = true
}

func DisableStub() {
	useStub = true
	stubLock.Unlock()
}

func ResetStub() {
	stub = NewStub()
}

func CurrStub() *Stub {
	return stub
}

func Stubbed() bool {
	return useStub
}

func Now() time.Time {
	if useStub {
		return stub.Now()
	} else {
		return time.Now()
	}
}
