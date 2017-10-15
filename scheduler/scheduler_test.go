package scheduler

import (
	"testing"

	"github.com/bytearena/schnapps"
	"github.com/bytearena/schnapps/types"
	"github.com/stretchr/testify/assert"
)

func TestFixedVMPool(t *testing.T) {
	provisionVmFn := func() *vm.VM {
		return &vm.VM{}
	}

	size := 1
	pool, err := NewFixedVMPool(size, provisionVmFn, NoHealtcheck)
	assert.Nil(t, err)

	e1, errPop1 := pool.Pop()
	assert.Nil(t, errPop1)
	assert.NotNil(t, e1)

	assert.Equal(t, pool.GetBackendSize(), 0)

	// Expect error
	e2, errPop2 := pool.Pop()
	assert.Nil(t, e2)
	assert.NotNil(t, errPop2)

	pool.Release(e1)

	assert.Equal(t, pool.GetBackendSize(), 1)
}

func TestPoolDelete(t *testing.T) {
	provisionVmFn := func() *vm.VM {
		return &vm.VM{}
	}

	size := 1
	pool, err := NewFixedVMPool(size, provisionVmFn, NoHealtcheck)
	assert.Nil(t, err)

	e1, errPop1 := pool.Pop()
	assert.Nil(t, errPop1)
	assert.NotNil(t, e1)

	pool.Delete(e1)

	assert.Equal(t, pool.GetBackendSize(), 1)
}

func TestPoolGC(t *testing.T) {
	healtcheckInc := 0
	provisionInc := 0

	healtcheckVmFn := func(x *vm.VM) bool {
		healtcheckInc++
		return false
	}

	provisionVmFn := func() *vm.VM {
		provisionInc++
		var config types.VMConfig

		return vm.NewVM(config)
	}

	size := 1
	pool, err := NewFixedVMPool(size, provisionVmFn, healtcheckVmFn)
	assert.Nil(t, err)

	for i := 0; i < NOK_HEALTCH_BEFORE_REMOVAL; i++ {
		pool.gc()
	}

	assert.Equal(t, healtcheckInc, size*NOK_HEALTCH_BEFORE_REMOVAL)
	assert.Equal(t, provisionInc, size*2)

	assert.Equal(t, pool.GetBackendSize(), 1)
}

func TestPoolSelectAndPop(t *testing.T) {
	provisionVmFn := func() *vm.VM {
		return &vm.VM{}
	}

	size := 1
	pool, err := NewFixedVMPool(size, provisionVmFn, NoHealtcheck)
	assert.Nil(t, err)

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
}

func TestPoolProvisionRetriesLimit(t *testing.T) {
	i := 0
	provisionVmFn := func() *vm.VM {
		i++
		return nil
	}

	_, err := NewFixedVMPool(1, provisionVmFn, NoHealtcheck)
	assert.NotNil(t, err)

	assert.Equal(t, err, PROVISION_LIMIT_ERROR)
	assert.Equal(t, i, PROVISION_RETRY_TIMES)
}
