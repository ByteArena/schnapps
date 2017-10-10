package scheduler

import (
	"errors"

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

	return pool, err
}

func (p *Pool) init() error {
	for i := 1; i <= p.size; i++ {
		vm := p.provisionVmFn()

		if vm != nil {
			err := p.Release(vm)

			if err != nil {
				return err
			}
		}
	}

	return nil
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
