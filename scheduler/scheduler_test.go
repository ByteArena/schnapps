package scheduler

import (
	"testing"

	"github.com/bytearena/schnapps"
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

	e2, errPop2 := pool.Pop()
	assert.Nil(t, e2)
	assert.NotNil(t, errPop2)

	pool.Release(e1)

	e3, errPop3 := pool.Pop()
	assert.Nil(t, errPop3)
	assert.NotNil(t, e3)
}

func TestFixedVMPoolDelete(t *testing.T) {
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

	e2, errPop2 := pool.Pop()
	assert.Nil(t, errPop2)
	assert.NotNil(t, e2)
}

func TestFixedVMPoolGC(t *testing.T) {
	healtcheckInc := 0
	provisionInc := 0

	healtcheckVmFn := func(x *vm.VM) bool {
		healtcheckInc++
		return false
	}

	provisionVmFn := func() *vm.VM {
		provisionInc++
		return &vm.VM{}
	}

	size := 1
	pool, err := NewFixedVMPool(size, provisionVmFn, healtcheckVmFn)
	assert.Nil(t, err)

	pool.gc()

	assert.Equal(t, healtcheckInc, size)
	assert.Equal(t, provisionInc, size*2)
}
