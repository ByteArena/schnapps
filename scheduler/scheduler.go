package scheduler

import (
	"errors"
	"time"

	"github.com/bytearena/schnapps"
)

type Queue []*vm.VM

type Pool struct {
	size  int
	queue Queue

	provisionVmFn func() *vm.VM
}

func NewFixedVMPool(size int, provisionVmFn func() *vm.VM) (*Pool, error) {
	if size < 0 {
		return nil, errors.New("Pool size cannot be negative")
	}

	pool := &Pool{
		size:          size,
		queue:         make(Queue, 0),
		provisionVmFn: provisionVmFn,
	}

	err := pool.init()

	go pool.runBackgroundGC(time.Duration(5 * time.Second))

	return pool, err
}

/*
	The garbage collection is reponsible for maintaining a healthy set of VM.
	Currently it only checks for VM derefernces (nil)
*/
func (p *Pool) runBackgroundGC(interval time.Duration) {
	for {
		for _, vm := range p.queue {
			if vm == nil {
				p.Delete(vm)
			}
		}

		<-time.After(interval)
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
		vm := p.provisionVmFn()

		if vm == nil {
			return nil, errors.New("Cannot pop element: backend is empty")
		} else {
			return vm, nil
		}
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
	newQueue := make(Queue, 0)

	for _, vm := range p.queue {
		if deletedVm != vm {
			newQueue = append(newQueue, vm)
		}
	}

	if len(newQueue) <= p.size {
		newVm := p.provisionVmFn()

		if newVm != nil {
			if err := p.Release(newVm); err != nil {
				return err
			}
		}
	}

	p.queue = newQueue

	return nil
}
