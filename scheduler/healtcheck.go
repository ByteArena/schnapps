package scheduler

import (
	"github.com/bytearena/schnapps"
)

func NoHealtcheck(*vm.VM) bool {
	return true
}
