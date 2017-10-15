package scheduler

import (
	"errors"
	"time"

	"github.com/bytearena/schnapps"
)

var (
	GC_INTERVAL                = time.Duration(5 * time.Second)
	NOK_HEALTCH_BEFORE_REMOVAL = 5

	PROVISION_RETRY_TIMES = 3
	PROVISION_LIMIT_ERROR = errors.New("Cannot provision pool: retry limit reached")
)

type Queue []*vm.VM

type Pool struct {
	size  int
	queue Queue

	provisionVmFn  func() *vm.VM
	healtcheckVmFn func(*vm.VM) bool

	nokHealthChecksByVm map[*vm.VM]int

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

		nokHealthChecksByVm: make(map[*vm.VM]int),

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
			if _, hasCount := p.nokHealthChecksByVm[vm]; !hasCount {
				p.nokHealthChecksByVm[vm] = 0
			}

			p.nokHealthChecksByVm[vm]++

			if p.nokHealthChecksByVm[vm] >= NOK_HEALTCH_BEFORE_REMOVAL {
				err := vm.Quit()

				if err != nil {
					delete(p.nokHealthChecksByVm, vm)
					p.Delete(vm)
				}
			}
		} else {

			// Reset healthchecks fail incr
			if _, hasCount := p.nokHealthChecksByVm[vm]; hasCount {
				p.nokHealthChecksByVm[vm] = 0
			}
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

out:
	for {
		if i >= p.size {
			break
		}

		provisionRetries := 0

		for {
			if provisionRetries >= PROVISION_RETRY_TIMES {
				return PROVISION_LIMIT_ERROR
			}

			vm := p.provisionVmFn()

			if vm != nil {
				err := p.Release(vm)

				if err != nil {
					return err
				}

				i++
				continue out
			} else {
				provisionRetries++
			}
		}
	}

	return nil
}

func (p *Pool) SelectAndPop(take func(*vm.VM) bool) (*vm.VM, error) {
	if len(p.queue) == 0 {
		return nil, errors.New("Cannot pop element: backend is empty")
	}

	for k, e := range p.queue {
		if take(e) == true {
			p.queue = append(p.queue[:k], p.queue[k+1:]...)
			return e, nil
		}
	}

	return nil, nil
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

func (p *Pool) GetBackendSize() int {
	return len(p.queue)
}

func (p *Pool) Delete(deletedVm *vm.VM) error {
	i := 0

	for {
		if i >= PROVISION_RETRY_TIMES {
			return PROVISION_LIMIT_ERROR
		}

		i++

		vm := p.provisionVmFn()

		if vm != nil {
			p.Release(vm)
			break
		}
	}

	return nil
}
