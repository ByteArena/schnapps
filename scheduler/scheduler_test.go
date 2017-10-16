package scheduler

import (
	"testing"

	"github.com/bytearena/schnapps"
	"github.com/stretchr/testify/assert"
)

func TestFixedVMPool(t *testing.T) {
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
}

func TestPoolDelete(t *testing.T) {
	size := 1
	pool, err := NewFixedVMPool(size)
	assert.Nil(t, err)

	test := func() {

		e1, errPop1 := pool.Pop()
		assert.Nil(t, errPop1)
		assert.NotNil(t, e1)

		pool.Delete(e1)

		assert.Equal(t, pool.GetBackendSize(), 0)

		pool.Stop()
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

}

func TestPoolGC(t *testing.T) {
	healtcheckInc := 0
	provisionInc := 0
	size := 1

	pool, err := NewFixedVMPool(size)
	assert.Nil(t, err)

	test := func() {

		for i := 0; i < NOK_HEALTCH_BEFORE_REMOVAL; i++ {
			pool.gc()
		}

		assert.Equal(t, healtcheckInc, size*NOK_HEALTCH_BEFORE_REMOVAL)
		assert.Equal(t, provisionInc, size*2)

		assert.Equal(t, pool.GetBackendSize(), 1)

		pool.Stop()
	}

	go func() {
		events := pool.Events()

		for {
			select {
			case msg := <-events:
				switch msg := msg.(type) {
				case HEALTHCHECK:
					healtcheckInc++
					pool.Consumer() <- HEALTHCHECK_RESULT{
						VM:  msg.VM,
						Res: false,
					}
				case PROVISION:
					provisionInc++
					vm := &vm.VM{}
					pool.Consumer() <- PROVISION_RESULT{vm}
				case READY:
					test()
				}
			}
		}
	}()
}

func TestPoolSelectAndPop(t *testing.T) {
	size := 1
	pool, err := NewFixedVMPool(size)
	assert.Nil(t, err)

	test := func() {
		vm1, popErr1 := pool.SelectAndPop(func(vm *vm.VM) bool {
			return true
		})

		assert.Nil(t, popErr1)
		assert.NotNil(t, vm1)
		assert.Equal(t, pool.GetBackendSize(), 0)

		vm2, popErr2 := pool.SelectAndPop(func(vm *vm.VM) bool {
			return true
		})

		assert.Nil(t, vm2)
		assert.NotNil(t, popErr2)

		pool.Stop()

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
}
