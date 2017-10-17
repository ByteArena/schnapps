package scheduler

import (
	"errors"
	"sync"
	"sync/atomic"
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

	initCount        int32
	healthcheckCount int32

	healthcheckConsumerQueue map[*vm.VM]bool
	nokHealthChecksByVm      map[*vm.VM]int

	stopTheWorldMutex sync.Mutex
	tickGC            *time.Ticker
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
	resume := p.stopTheWorld()
	p.sendHealthcheckRequests()
	resume()

	waitChan := make(chan bool)
	timeoutChan := time.After(HEALTHCHECK_TIMEOUT)

	go func() {
		pollUntil(func() bool {
			count := atomic.LoadInt32(&p.healthcheckCount)

			if count <= 0 {
				atomic.StoreInt32(&p.healthcheckCount, 0)
				return true
			} else {
				return false
			}
		})

		waitChan <- true
	}()

	select {
	case <-timeoutChan:
	case <-waitChan:
		queue := make(map[*vm.VM]bool)

		// Copy array to avoid having to lock it
		resume := p.stopTheWorld()
		for k, v := range p.healthcheckConsumerQueue {
			queue[k] = v
		}
		resume()

		p.processHealtcheckConsumerQueue(queue)
		close(waitChan)
		return
	}
}

func (p *Pool) sendHealthcheckRequests() {
	p.healthcheckConsumerQueue = make(map[*vm.VM]bool)

	atomic.StoreInt32(&p.healthcheckCount, int32(len(p.queue)))

	for _, vm := range p.queue {
		p.healthcheckConsumerQueue[vm] = false

		p.produceEvent(HEALTHCHECK{vm})
	}
}

func (p *Pool) stopTheWorld() (resume func()) {
	p.stopTheWorldMutex.Lock()

	return func() {
		p.stopTheWorldMutex.Unlock()
	}
}

func (p *Pool) processHealtcheckConsumerQueue(queue map[*vm.VM]bool) {

	if len(queue) == 0 {
		return
	}

	for vm, res := range queue {

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
	atomic.StoreInt32(&p.initCount, int32(p.size))

	for i := 0; i < p.size; i++ {
		p.produceEvent(PROVISION{})
	}

	pollUntil(func() bool {
		count := atomic.LoadInt32(&p.initCount)

		if count <= 0 {
			atomic.StoreInt32(&p.initCount, 0)
			return true
		} else {
			return false
		}
	})

	p.produceEvent(READY{})

	return nil
}

func (p *Pool) Pop() (*vm.VM, error) {
	resume := p.stopTheWorld()
	defer resume()

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
	atomic.AddInt32(&p.healthcheckCount, 1)

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
				resume := p.stopTheWorld()

				if len(p.healthcheckConsumerQueue) <= len(p.queue) && isVmInQueue(p.queue, msg.VM) {

					if !p.healthcheckConsumerQueue[msg.VM] {
						atomic.AddInt32(&p.healthcheckCount, -1)
					}

					p.healthcheckConsumerQueue[msg.VM] = msg.Res
				} else {
					err := errors.New("Unexpected healtchcheck")
					p.produceEvent(ERROR{err})
				}

				resume()

			case PROVISION_RESULT:
				resume := p.stopTheWorld()
				err := p.Release(msg.VM)
				resume()

				if err == nil {
					atomic.AddInt32(&p.initCount, -1)
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
