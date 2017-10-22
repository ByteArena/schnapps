package vm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/bytearena/schnapps/cli"
	"github.com/bytearena/schnapps/libvirt"
	schnappsqmp "github.com/bytearena/schnapps/qmp"
	"github.com/bytearena/schnapps/types"
	"github.com/bytearena/schnapps/utils"
	"github.com/digitalocean/go-qemu/qmp"
)

type VM struct {
	Config  types.VMConfig
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	process *os.Process
	qmp     *qmp.SocketMonitor
	events  chan qmp.Event
}

func NewVM(config types.VMConfig) *VM {
	qmpport := schnappsqmp.GetNextPort()

	config.QMPServer = &types.QMPServer{
		Protocol: "tcp",
		Addr:     "localhost:" + strconv.Itoa(qmpport),
	}

	return &VM{
		Config: config,
	}
}

func (vm *VM) readStdout(reader io.Reader) {
	buffReader := bufio.NewReader(reader)

	for {
		line, _, readErr := buffReader.ReadLine()

		if readErr == io.EOF {
			break
		}

		if len(line) == 0 {
			continue
		}

		vm.Log(string(line))
	}
}

func (vm *VM) Log(msg string) {
	fmt.Printf("[VM %d] %s\n", vm.Config.Id, msg)
}

func (vm *VM) Quit() error {
	vm.Log("Halting...")

	command := []byte("{ \"execute\": \"quit\" }")

	if vm.qmp == nil {
		return errors.New("Cannot halt VM: not connected to the QMP server")
	}

	_, err := vm.qmp.Run(command)

	if err != nil {
		return err
	}

	timeout := time.After(3 * time.Second)

	for {
		select {
		case e := <-vm.events:
			if e.Event == libvirt.EVENT_SHUTDOWN {
				break
			}
		case <-timeout:
			err := vm.killProcess()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) killProcess() error {
	vm.Log("Killing process...")

	if vm.process == nil {
		return errors.New("Could not kill process: process not available")
	}

	vm.process.Kill()

	return nil
}

func (vm *VM) Close() {
	vm.Log("Releasing resources...")

	var closeErr error

	if vm.qmp != nil {
		closeErr = vm.qmp.Disconnect()
		utils.RecoverableCheck(closeErr, "Could not disconnect from qmp server")
	}

	closeErr = vm.stdout.Close()
	utils.RecoverableCheck(closeErr, "Could not close stdout")

	closeErr = vm.stderr.Close()
	utils.RecoverableCheck(closeErr, "Could not close stderr")

	closeErr = vm.process.Release()
	utils.RecoverableCheck(closeErr, "Could not close process")

	vm.process = nil
}

// FIXME(sven): determine if KVM has booted the VM
func (vm *VM) WaitUntilBooted() error {
	fakeProcess := time.After(10 * time.Second)

	<-fakeProcess
	return nil
}

func (vm *VM) Start() error {
	kvmbin, err := exec.LookPath("kvm")

	if err != nil {
		return errors.New("Error: kvm not found in $PATH")
	}

	cmd := cli.CreateKVMCommand(kvmbin, vm.Config)

	stdout, stdoutErr := cmd.StdoutPipe()
	utils.Check(stdoutErr, "Could not get stdout")

	stderr, stderrErr := cmd.StderrPipe()
	utils.Check(stderrErr, "Could not get stderr")

	vm.stdout = stdout
	vm.stderr = stderr

	vm.Log("Starting...")

	err = cmd.Start()

	if err != nil {
		return errors.New("Error: VM could not be Started: " + err.Error())
	}

	vm.process = cmd.Process

	go vm.readStdout(stdout)
	go vm.readStdout(stderr)

	// Connect QMP
	<-time.After(1 * time.Second)

	qmp, socketMonitorErr := qmp.NewSocketMonitor(vm.Config.QMPServer.Protocol, vm.Config.QMPServer.Addr, 20*time.Second)
	utils.Check(socketMonitorErr, "Could not connect to QMP socket")

	monitorErr := qmp.Connect()

	if monitorErr != nil {
		vm.Close()

		return errors.New("Could not connect monitoring to QMP server")
	}

	vm.qmp = qmp

	// Register event consumer
	events, eventsErr := vm.qmp.Events()
	utils.Check(eventsErr, "could not consume events")

	go func() {
		for {
			select {
			case e := <-events:
				if e.Event != "" {
					vm.events <- e
				}
			}

		}
	}()

	go func() {
		waitErr := cmd.Wait()
		utils.Check(waitErr, "Could not wait VM process")

		vm.Log("Stopped")
		vm.Close()
	}()

	return nil
}
