package methods

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/weaveworks/ignite/pkg/logs"
)

// JAILER CONFIGURATION
func JailerEnabledVM() {
	UID := 123
	GID := 100
	nsPath, err := CreateContainer("testVm")
	if err != nil {
		panic(err)
	}
	if err := SetUpSandBoxNetwork(nsPath, UID, GID); err != nil {
		panic(err)
	}
	const socketPath = "api.socket"
	ctx := context.Background()
	vmmCtx, vmmCancel := context.WithCancel(ctx)
	defer vmmCancel()

	const id = "4569"
	//
	const kernelImagePath = "../vmlinux-5.10.210"
	networkIfaces := []firecracker.NetworkInterface{{
		StaticConfiguration: &firecracker.StaticNetworkConfiguration{
			// MacAddress:  "AA:FC:00:00:00:01", // potential bug here
			MacAddress:  "52:54:00:ab:cd:ef",
			HostDevName: "tap0",
			IPConfiguration: &firecracker.IPConfiguration{
				IPAddr: net.IPNet{
					IP:   net.IPv4(172, 16, 0, 2),
					Mask: net.IPMask{255, 255, 255, 0},
				},
				Gateway:     net.IPv4(172, 16, 0, 1),
				Nameservers: []string{"8.8.8.8"},
				IfName:      "eth0",
			},
		},
	}}
	stdOutPath := "/dev/null"
	stdout, err := os.OpenFile(stdOutPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logs.Logger.Errorf("failed to create stdout file: %v", err)
	}
	stdErrPath := "/dev/null"
	stderr, err := os.OpenFile(stdErrPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logs.Logger.Errorf("failed to create stderr file: %v", err)
	}
	stdInPath := "/dev/null"
	stdin, err := os.OpenFile(stdInPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logs.Logger.Errorf("failed to create stderr file: %v", err)
	}

	fcCfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: kernelImagePath,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off",
		Drives:          firecracker.NewDrivesBuilder("../ubuntu-22.04.ext4.3").Build(),
		LogLevel:        "Debug",
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(2),
			Smt:        firecracker.Bool(false),
			MemSizeMib: firecracker.Int64(2048),
		},
		JailerCfg: &firecracker.JailerConfig{
			UID:            &UID,
			GID:            &GID,
			Daemonize:      false,
			ID:             id,
			NumaNode:       firecracker.Int(0),
			JailerBinary:   "../jailer",
			ChrootBaseDir:  "/srv/jailer",
			Stdin:          stdin,
			Stdout:         stdout,
			Stderr:         stderr,
			CgroupVersion:  "2",
			ChrootStrategy: firecracker.NewNaiveChrootStrategy(kernelImagePath),
			ExecFile:       "../firecracker",
		},
		NetNS:             nsPath,
		NetworkInterfaces: networkIfaces,
	}

	// Check if kernel image is readable
	f, err := os.Open(fcCfg.KernelImagePath)
	if err != nil {
		panic(fmt.Errorf("failed to open kernel image: %v", err))
	}
	f.Close()

	// Check each drive is readable and writable
	for _, drive := range fcCfg.Drives {
		drivePath := firecracker.StringValue(drive.PathOnHost)
		f, err := os.OpenFile(drivePath, os.O_RDWR, 0666)
		if err != nil {
			panic(fmt.Errorf("failed to open drive with read/write permissions: %v", err))
		}
		f.Close()
	}

	m, err := firecracker.NewMachine(vmmCtx, fcCfg)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	if err := m.Start(vmmCtx); err != nil {
		log.Println(err)
		panic(err)
	}
	defer m.StopVMM()

	// wait for the VMM to exit
	if err := m.Wait(vmmCtx); err != nil {
		log.Println(err)
		panic(err)
	}
}
