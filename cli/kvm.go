package cli

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/bytearena/schnapps/types"
)

func CreateKVMCommand(kvmbin string, config types.VMConfig) *exec.Cmd {

	args := []string{
		"-name", strconv.Itoa(config.Id),
		"-m", strconv.Itoa(config.MegMemory) + "M",
		"-snapshot",
		"-smp", strconv.Itoa(config.CPUAmount) + ",cores=" + strconv.Itoa(config.CPUCoreAmount),
		"-nographic",
		"-no-fd-bootchk",
		"-drive", "file=" + config.ImageLocation + ",if=virtio,cache=none,format=raw,index=1",
	}

	args = append(args, buildNetArgs(config.NICs)...)
	args = append(args, buildQMPServer(config.QMPServer)...)

	cmd := exec.Command(kvmbin, args...)
	cmd.Env = nil

	return cmd
}

func buildNetArgs(NICs []interface{}) []string {
	args := []string{}

	for _, e := range NICs {
		switch nic := e.(type) {
		case types.NICBridge:
			args = append(
				args,
				[]string{
					"-netdev",
					fmt.Sprintf("bridge,br=%s,id=net0", nic.Bridge),
					"-device",
					fmt.Sprintf("virtio-net,netdev=net0,mac=%s", nic.MAC),
				}...,
			)

		case types.NICIface:
			args = append(
				args,
				[]string{
					"-net",
					fmt.Sprintf("nic,model=%s", nic.Model),
				}...,
			)

		case types.NICTap:
			args = append(
				args,
				[]string{
					"-net",
					fmt.Sprintf("tap,ifname=%s,script=no,downscript=no", nic.Ifname),
				}...,
			)
		case types.NICUser:
			args = append(
				args,
				[]string{
					"-net",
					fmt.Sprintf("user,dhcpstart=%s,net=%s", nic.DHCPStart, nic.Net),
				}...,
			)
		case types.NICSocket:
			args = append(
				args,
				[]string{
					"-net",
					fmt.Sprintf("socket,connect=%s", nic.Connect),
				}...,
			)
		default:
			panic("Unknow NIC type")
		}
	}

	return args
}

func buildQMPServer(config *types.QMPServer) []string {
	args := []string{}

	if config != nil {
		return []string{"-qmp", config.Protocol + ":" + config.Addr + ",server"}
	}

	return args
}
