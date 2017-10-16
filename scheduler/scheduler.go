package scheduler

import (
	"errors"
	"sync"
	"time"

	"github.com/bytearena/schnapps"
)

var (
	GC_INTERVAL                = time.Duration(5 * time.Second)
	NOK_HEALTCH_BEFORE_REMOVAL = 15
	HEALTHCHECK_TIMEOUT        = time.Duration(3 * time.Second)

	PROVISION_LIMIT_ERROR = errors.New("Cannot provision pool: retry limit reached")
)

type Queue []*vm.VM

type ProducerChan chan interface{}
type ConsumerChan chan interface{}

type Pool struct {
	size  int
	queue Queue

	producer     ProducerChan
	consumer     ConsumerChan
	stopConsumer chan bool

	initwg        sync.WaitGroup
	healthcheckwg sync.WaitGroup

	healthcheckConsumerQueue map[*vm.VM]bool
	nokHealthChecksByVm      map[*vm.VM]int

	tickGC *time.Ticker
}

func NewFixedVMPool(size int) (*Pool, error) {
	if size < 0 {
		return nil, errors.New("Pool size cannot be negative")
	}

	pool := &Pool{
		size:  size,
		queue: make(Queue, 0),

		producer:     make(ProducerChan),
		consumer:     make(ConsumerChan),
		stopConsumer: make(chan bool),

		nokHealthChecksByVm: make(map[*vm.VM]int),

		tickGC: time.NewTicker(GC_INTERVAL),
	}

	go pool.init()
	go pool.consumeEvents()

	return pool, nil
}

/*
	The garbage collection is reponsible for maintaining a healthy set of VM.
*/
func (p *Pool) gc() {

	if len(p.queue) == 0 {
		return // Nothing to check here
	}

	p.healthcheckConsumerQueue = make(map[*vm.VM]bool)

	for _, vm := range p.queue {
		p.healthcheckConsumerQueue[vm] = false

		p.produceEvent(HEALTHCHECK{vm})
		p.healthcheckwg.Add(1)
	}

	waitChan := make(chan bool)
	timeoutChan := time.After(HEALTHCHECK_TIMEOUT)

	go func() {
		p.healthcheckwg.Wait()
		waitChan <- true
	}()

	select {
	case <-timeoutChan:
	case <-waitChan:
		p.processHealtcheckConsumerQueue()
		close(waitChan)
		return
	}
}

func (p *Pool) processHealtcheckConsumerQueue() {
	if len(p.healthcheckConsumerQueue) == 0 {
		return
	}

	for vm, res := range p.healthcheckConsumerQueue {

		if res == false {
			if _, hasCount := p.nokHealthChecksByVm[vm]; !hasCount {
				p.nokHealthChecksByVm[vm] = 0
			}

			p.nokHealthChecksByVm[vm]++

			if p.nokHealthChecksByVm[vm] >= NOK_HEALTCH_BEFORE_REMOVAL {
				delete(p.nokHealthChecksByVm, vm)
				p.Delete(vm)
			}
		} else {

			// Reset healthchecks fail incr
			if _, hasCount := p.nokHealthChecksByVm[vm]; hasCount {
				p.nokHealthChecksByVm[vm] = 0
			}
		}
	}
}

func (p *Pool) init() error {
	p.initwg.Add(p.size)

	for i := 0; i < p.size; i++ {
		p.produceEvent(PROVISION{})
	}

	p.initwg.Wait()

	p.produceEvent(READY{})

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
	p.healthcheckwg.Add(1)

	p.produceEvent(VM_UNHEALTHY{deletedVm})
	p.produceEvent(PROVISION{})

	return nil
}

func (p *Pool) Events() ProducerChan {
	return p.producer
}

func (p *Pool) Consumer() ConsumerChan {
	return p.consumer
}

func (p *Pool) Stop() {
	p.stopConsumer <- true

	close(p.stopConsumer)
}

// Asynchronously emit a action form the scheduler
func (p *Pool) produceEvent(msg interface{}) {
	go func() {
		p.producer <- msg
	}()
}

func (p *Pool) consumeEvents() {
	for {
		select {
		case <-p.stopConsumer:
			return

		case msg := <-p.consumer:
			switch msg := msg.(type) {

			case HEALTHCHECK_RESULT:
				p.healthcheckConsumerQueue[msg.VM] = msg.Res
				p.healthcheckwg.Done()

			case PROVISION_RESULT:
				err := p.Release(msg.VM)

				if err == nil {
					p.initwg.Done()
				} else {
					p.produceEvent(ERROR{err})
				}

			default:
				panic("Received unsupported message")
			}

		case <-p.tickGC.C:
			p.gc()
		}
	}
}
