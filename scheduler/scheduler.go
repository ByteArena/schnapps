package scheduler

import (
	"errors"
	"time"

	"github.com/bytearena/schnapps"
)

const (
	GC_INTERVAL = time.Duration(5 * time.Second)
)

type Queue []*vm.VM

type Pool struct {
	size  int
	queue Queue

	provisionVmFn  func() *vm.VM
	healtcheckVmFn func(*vm.VM) bool

	tickGC *time.Ticker
	stopGC chan bool
}

func NewFixedVMPool(size int, provisionVmFn func() *vm.VM, healtcheckVmFn func(*vm.VM) bool) (*Pool, error) {
	if size < 0 {
		return nil, errors.New("Pool size cannot be negative")
	}

	pool := &Pool{
		size:           size,
		queue:          make(Queue, 0),
		provisionVmFn:  provisionVmFn,
		healtcheckVmFn: healtcheckVmFn,

		tickGC: time.NewTicker(GC_INTERVAL),
		stopGC: make(chan bool),
	}

	err := pool.init()

	go pool.runBackgroundGC()

	return pool, err
}

func (p *Pool) gc() {

	if len(p.queue) == 0 {
		return // Nothing to check here
	}

	for _, vm := range p.queue {

		if p.healtcheckVmFn(vm) == false {
			p.Delete(vm)
		}
	}
}

/*
	The garbage collection is reponsible for maintaining a healthy set of VM.
*/
func (p *Pool) runBackgroundGC() {

	for {
		select {
		case <-p.stopGC:
			return

		case <-p.tickGC.C:
			p.gc()

		}

	}
}

func (p *Pool) init() error {
	var i = 0

	for {
		vm := p.provisionVmFn()

		if vm != nil {
			err := p.Release(vm)

			if err != nil {
				return err
			}

			i++
		}

		if i >= p.size {
			break
		}
	}

	return nil
}

func (p *Pool) SelectAndPop(take func(*vm.VM) bool) (*vm.VM, error) {
	var takeElement *vm.VM
	newQueue := p.queue

	if len(newQueue) == 0 {
		vm := p.provisionVmFn()

		if vm != nil {
			return nil, errors.New("Cannot pop element: backend is empty")
		} else {
			newQueue = append(newQueue, vm)
		}
	}

	for _, e := range newQueue {
		if takeElement == nil && take(e) == true {
			takeElement = e
		} else {
			newQueue = append(newQueue, e)
		}
	}

	p.queue = newQueue

	return takeElement, nil
}

func (p *Pool) Pop() (*vm.VM, error) {
	if len(p.queue) == 0 {
		return nil, errors.New("Cannot pop element: backend is empty")
	} else {
		queueLen := len(p.queue)

		e := p.queue[queueLen-1]
		p.queue = p.queue[:queueLen-1]

		return e, nil
	}
}

func (p *Pool) Release(e *vm.VM) error {
	if p.size <= len(p.queue) {
		return errors.New("Cannot release element: backend reached the limit")
	}

	p.queue = append(p.queue, e)

	return nil
}

func (p *Pool) Delete(deletedVm *vm.VM) error {
	for {
		vm := p.provisionVmFn()

		if vm != nil {
			p.Release(vm)
			break
		}
	}

	return nil
}
