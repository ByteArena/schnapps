package id

import (
	"testing"

	"github.com/bytearena/schnapps"
	"github.com/bytearena/schnapps/types"
	"github.com/stretchr/testify/assert"
)

func TestGetVMMac(t *testing.T) {
	value := "foo"

	config := types.VMConfig{
		NICs: []interface{}{
			types.NICBridge{
				Bridge: "br",
				MAC:    value,
			},
		},
		Id:            1,
		MegMemory:     1,
		CPUAmount:     1,
		CPUCoreAmount: 1,
		ImageLocation: ".",
	}

	vm := vm.NewVM(config)

	mac, hasMac := GetVMMAC(vm)

	assert.True(t, hasMac)
	assert.Equal(t, mac, value)
}

func TestGetVMMacNoNics(t *testing.T) {
	config := types.VMConfig{
		NICs:          []interface{}{},
		Id:            1,
		MegMemory:     1,
		CPUAmount:     1,
		CPUCoreAmount: 1,
		ImageLocation: ".",
	}

	vm := vm.NewVM(config)

	mac, hasMac := GetVMMAC(vm)

	assert.False(t, hasMac)
	assert.Equal(t, mac, "")
}

func TestGetVMMacNilVm(t *testing.T) {
	mac, hasMac := GetVMMAC(nil)

	assert.False(t, hasMac)
	assert.Equal(t, mac, "")
}
