package methods

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

// JAILER CONFIGURATION
func JailerEnabledVM() {
	UID := 123
	GID := 100

	const id = "4580"
	const socketPath = "api.socket"
	const kernelImagePath = "./fc_kernel"
	const rfsPath = "./fc_rfs"
	const firecrackerPath = "./firecracker"
	const jailerPath = "./jailer"
	const ChrootBaseDir = "/srv/jailer"

	nsPath, err := CreateContainer("testVm")
	if err != nil {
		panic(err)
	}
	if err := SetUpSandBoxNetwork(nsPath, UID, GID); err != nil {
		panic(err)
	}

	ctx := context.Background()
	vmmCtx, vmmCancel := context.WithCancel(ctx)
	defer vmmCancel()

	networkIfaces := []firecracker.NetworkInterface{{
		StaticConfiguration: &firecracker.StaticNetworkConfiguration{
			MacAddress:  "52:54:00:ab:cd:ef",
			HostDevName: "tap0", // interface in the sandbox net namespace
			IPConfiguration: &firecracker.IPConfiguration{
				IPAddr: net.IPNet{
					IP:   net.IPv4(172, 16, 0, 2),      // Ip Address of the vm.
					Mask: net.IPMask{255, 255, 255, 0}, // subnet mask
				},
				Gateway:     net.IPv4(172, 16, 0, 1),
				Nameservers: []string{"8.8.8.8"}, // nameserver set to google dns
				IfName:      "eth0",              // interface of vm
			},
		},
	}}

	fcCfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: kernelImagePath,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off",
		Drives:          firecracker.NewDrivesBuilder(rfsPath).Build(),
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
			JailerBinary:   jailerPath,
			ChrootBaseDir:  ChrootBaseDir,
			Stdin:          os.Stdin,
			Stdout:         os.Stdout,
			Stderr:         os.Stderr,
			CgroupVersion:  "2",
			ChrootStrategy: firecracker.NewNaiveChrootStrategy(kernelImagePath),
			ExecFile:       firecrackerPath,
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
