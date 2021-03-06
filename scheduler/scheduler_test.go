package scheduler

import (
	"testing"

	"github.com/bytearena/schnapps"
	"github.com/stretchr/testify/assert"
)

func TestFixedVMPool(t *testing.T) {
	wait := make(chan bool)
	size := 1

	pool, err := NewFixedVMPool(size)
	assert.Nil(t, err)

	test := func() {
		e1, errPop1 := pool.Pop()
		assert.Nil(t, errPop1)
		assert.NotNil(t, e1)

		assert.Equal(t, pool.GetBackendSize(), 0)

		// Expect error
		e2, errPop2 := pool.Pop()
		assert.Nil(t, e2)
		assert.NotNil(t, errPop2)

		pool.Release(e1)

		assert.Equal(t, pool.GetBackendSize(), size)
		pool.Stop()
		wait <- false
	}

	go func() {
		events := pool.Events()

		for {
			select {
			case msg := <-events:
				switch msg.(type) {
				case PROVISION:
					vm := &vm.VM{}
					pool.Consumer() <- PROVISION_RESULT{vm}
				case READY:
					test()
				}
			}
		}
	}()
	<-wait
}

func TestPoolDelete(t *testing.T) {
	size := 1
	wait := make(chan bool)
	pool, err := NewFixedVMPool(size)
	assert.Nil(t, err)

	test := func() {

		e1, errPop1 := pool.Pop()
		assert.Nil(t, errPop1)
		assert.NotNil(t, e1)

		pool.Delete(e1)

		assert.Equal(t, pool.GetBackendSize(), 0)

		pool.Stop()
		wait <- false
	}

	go func() {
		events := pool.Events()

		for {
			select {
			case msg := <-events:
				switch msg.(type) {
				case PROVISION:
					vm := &vm.VM{}
					pool.Consumer() <- PROVISION_RESULT{vm}
				case READY:
					test()
				}
			}
		}
	}()

	<-wait
}

func TestPoolGC(t *testing.T) {
	healtcheckInc := 0
	provisionInc := 0
	size := 1
	wait := make(chan bool)

	pool, err := NewFixedVMPool(size)
	assert.Nil(t, err)

	test := func() {

		for i := 0; i < NOK_HEALTCH_BEFORE_REMOVAL; i++ {
			pool.gc()
		}

		wait <- false
	}

	go func() {
		events := pool.Events()

		for {
			select {
			case msg := <-events:
				switch msg := msg.(type) {
				case ERROR:
					assert.Fail(t, msg.err.Error())
					wait <- false

				case HEALTHCHECK:
					pool.Consumer() <- HEALTHCHECK_RESULT{
						VM:  msg.VM,
						Res: false,
					}
					healtcheckInc++

				case PROVISION:
					vm := &vm.VM{}
					pool.Consumer() <- PROVISION_RESULT{vm}
					provisionInc++

				case READY:
					go test()
				}
			case <-wait:
				assert.Equal(t, healtcheckInc, size*NOK_HEALTCH_BEFORE_REMOVAL)
				assert.Equal(t, provisionInc, size)

				assert.Equal(t, pool.GetBackendSize(), size)
				pool.Stop()
				return
			}
		}
	}()
}

func TestPoolGCOverProvision(t *testing.T) {
	errorCount := 0
	size := 1
	wait := make(chan bool)

	pool, err := NewFixedVMPool(size)
	assert.Nil(t, err)

	go func() {
		events := pool.Events()

		for {
			select {
			case msg := <-events:
				switch msg := msg.(type) {
				case HEALTHCHECK:
					pool.Consumer() <- HEALTHCHECK_RESULT{
						VM:  msg.VM,
						Res: false,
					}
				case PROVISION:
					vm := &vm.VM{}
					pool.Consumer() <- PROVISION_RESULT{vm}

					// Send two
					pool.Consumer() <- PROVISION_RESULT{vm}
				case ERROR:
					errorCount++
					wait <- false
				case READY:
					pool.Stop()
				}
			}
		}
	}()

	<-wait

	assert.Equal(t, errorCount, 1)
}

func TestPoolGCOverHealtcheck(t *testing.T) {
	errorCount := 0
	size := 1
	wait := make(chan bool)

	pool, err := NewFixedVMPool(size)
	assert.Nil(t, err)

	go func() {
		events := pool.Events()

		for {
			select {
			case msg := <-events:
				switch msg := msg.(type) {
				case HEALTHCHECK:
					health := HEALTHCHECK_RESULT{
						VM:  msg.VM,
						Res: true,
					}
					pool.Consumer() <- health

					// Send another with the same vm
					pool.Consumer() <- health

					// Send another with an unknown vm
					pool.Consumer() <- HEALTHCHECK_RESULT{
						VM:  &vm.VM{},
						Res: true,
					}
				case PROVISION:
					vm := &vm.VM{}
					pool.Consumer() <- PROVISION_RESULT{vm}
				case ERROR:
					errorCount++
					wait <- false
				case READY:
					go pool.gc()
				}
			}
		}
	}()

	<-wait
	pool.Stop()

	assert.Equal(t, errorCount, 1)
}
