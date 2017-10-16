package scheduler

import (
	"github.com/bytearena/schnapps"
)

// TODO(sven): use struct inheritance

type HEALTHCHECK struct{ VM *vm.VM }
type HEALTHCHECK_RESULT struct {
	VM  *vm.VM
	Res bool
}

// Note that it has already been deleted from the scheduler's internal state
type VM_UNHEALTHY struct{ VM *vm.VM }

type PROVISION struct{}
type PROVISION_RESULT struct {
	VM *vm.VM
}

type gc struct{}
type READY struct{}
