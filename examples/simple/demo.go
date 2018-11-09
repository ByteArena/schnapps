package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/bytearena/schnapps"
	vmid "github.com/bytearena/schnapps/id"
	vmtypes "github.com/bytearena/schnapps/types"
)

const (
	MEG_MEMORY            = 2048
	CPU_AMOUNT            = 1
	CPU_CORE_AMOUNT       = 1
	BRIDGE_NAME           = "brtest"
	CIDR                  = "10.1.0.1/24"
	VM_RAW_IMAGE_LOCATION = "./linuxkit/linuxkit.raw"
)

func main() {
	// keep track of vm in order to clean them later
	vms := map[int]*vm.VM{}

	for id := 0; id < 5; id++ {
		vm, err := SpawnWorker(id)

		if err != nil {
			panic(err)
		}

		if err := vm.WaitUntilBooted(); err != nil {
			panic(err)
		}

		log.Println("vm", "VM ("+strconv.Itoa(id)+") booted")

		vms[id] = vm
	}

	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)

	<-signal_channel

	for _, vm := range vms {
		vm.Close()
	}
}

func SpawnWorker(id int) (*vm.VM, error) {
	mac := vmid.GenerateRandomMAC()

	config := vmtypes.VMConfig{
		NICs: []interface{}{
			vmtypes.NICBridge{
				Bridge: BRIDGE_NAME,
				MAC:    mac,
			},
		},
		Id:            id,
		MegMemory:     MEG_MEMORY,
		CPUAmount:     CPU_AMOUNT,
		CPUCoreAmount: CPU_CORE_AMOUNT,
		ImageLocation: VM_RAW_IMAGE_LOCATION,
		Metadata:      vmtypes.VMMetadata{},
	}

	workerVm := vm.NewVM(config)

	startErr := workerVm.Start()

	if startErr != nil {
		return nil, startErr
	}

	log.Println("Started new VM (" + mac + ")")

	return workerVm, nil
}
