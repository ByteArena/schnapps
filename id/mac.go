package id

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bytearena/schnapps"
	"github.com/bytearena/schnapps/types"
)

const (
	HEXCHARS = "0123456789abcdef"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomHex(strlen int) string {
	result := make([]byte, strlen)
	for i := range result {
		result[i] = HEXCHARS[r.Intn(len(HEXCHARS))]
	}
	return string(result)
}

func GenerateRandomMAC() string {
	return fmt.Sprintf("00:f0:%s:%s:%s:%s", RandomHex(2), RandomHex(2), RandomHex(2), RandomHex(2))
}

func GetVMMAC(vm *vm.VM) (mac string, found bool) {

	for _, nic := range vm.Config.NICs {
		if bridge, ok := nic.(types.NICBridge); ok {
			return bridge.MAC, true
		}
	}

	return "", false
}
