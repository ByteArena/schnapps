package scheduler

import (
	"github.com/bytearena/schnapps"
)

func isVmInQueue(queue Queue, vm *vm.VM) bool {
	for _, e := range queue {
		if e == vm {
			return true
		}
	}

	return false
}
