package shortlivedpool

import "time"
import "sync/atomic"

type noCopy struct{}

const (
	maxIdleGet = 60
)

// Pool is a drop-in replacement for sync.Pool
// Its difference is that evicts older elements
// so memory can be freed
type Pool struct {
	noCopy  noCopy
	stack   EvictionStack
	lastGet int64
	New     func() interface{}
}

// Put adds x to the pool.
func (p *Pool) Put(x interface{}) {
	l := atomic.LoadInt64(&p.lastGet)
	if l > 0 && l+maxIdleGet < time.Now().Unix() {
		return
	}
	p.stack.Put(x)
}

// Get selects the most recent used item from the Pool,
// removes it from the Pool, and returns it to the caller.
// Get may choose to ignore the pool and treat it as empty.
// Callers should not assume any relation between values passed to Put and
// the values returned by Get.
//
// If Get would otherwise return nil and p.New is non-nil, Get returns
// the result of calling p.New.
func (p *Pool) Get() (x interface{}) {
	x = p.stack.Pop()
	if x != nil {
		return
	}

	atomic.StoreInt64(&p.lastGet, time.Now().Unix())
	if p.New != nil {
		return p.New()
	}
	return
}
