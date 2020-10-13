package workgroup

import (
	"context"
	"sync"
)

// WorkGroup resizable group of goroutine
type WorkGroup struct {
	fn      func()
	cancels []func()
	closed  []context.Context
	lock    sync.Locker
}

// NewWorkGroup init a group of goroutine with size and function
func NewWorkGroup(size int, fn func()) *WorkGroup {
	wg := &WorkGroup{
		fn:      fn,
		cancels: make([]func(), 0),
		closed:  make([]context.Context, 0),
		lock:    new(sync.Mutex),
	}
	wg.Resize(size)
	return wg
}

func (wg *WorkGroup) runFN(ctx context.Context, closed func()) {
	defer closed()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			wg.fn()
		}
	}
}

// Resize group size shutdown unnecessary goroutine or start more
func (wg *WorkGroup) Resize(n int) (size int) {
	size = len(wg.cancels)
	if n == size || n < 0 {
		return
	}
	wg.lock.Lock()
	defer wg.lock.Unlock()
	for i := size; n > i; i++ {
		wg.run()
	}
	for i := n; i < size; i++ {
		wg.stopI(i)
		wg.joinI(i)
	}
	wg.cancels = wg.cancels[:n]
	wg.closed = wg.closed[:n]
	return
	/*if n > size {
	//	wg.waitG.Add(n - size)
	//	for i := 0; i < n-size; i++ {
	//		wg.run()
	//	}
	//} else {
	//	for i := n; i < size; i++ {
	//		wg.stopI(i)
	//		wg.joinI(i)
	//	}
	//	wg.cancels = wg.cancels[:n]
	//	wg.closed = wg.closed[:n]
	//}*/
}

func (wg *WorkGroup) joinI(i int) struct{} {
	return <-wg.closed[i].Done()
}

func (wg *WorkGroup) stopI(i int) {
	wg.cancels[i]()
}

func (wg *WorkGroup) run() {
	ctx, cancel := context.WithCancel(context.TODO())
	wg.cancels = append(wg.cancels, cancel)
	closeCtx, closed := context.WithCancel(context.TODO())
	wg.closed = append(wg.closed, closeCtx)
	go wg.runFN(ctx, closed)
}

// Close shutdown all goroutines and wait them exit
func (wg *WorkGroup) Close() {
	wg.Resize(0)
}
