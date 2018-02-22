package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// createDescription is used to describe create command in detail and auto generate command doc.
var createDescription = "Create a static container object in Pouchd. " +
	"When creating, all configuration user input will be stored in memory store of Pouchd. " +
	"This is useful when you wish to create a container configuration ahead of time so that Pouchd will preserve the resource in advance. " +
	"The container you created is ready to start when you need it."

// CreateCommand use to implement 'create' command, it create a container.
type CreateCommand struct {
	container
	baseCommand
}

// Init initialize create command.
func (cc *CreateCommand) Init(c *Cli) {
	cc.cli = c
	cc.cmd = &cobra.Command{
		Use:   "create [OPTIONS] IMAGE [ARG...]",
		Short: "Create a new container with specified image",
		Long:  createDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.runCreate(args)
		},
		Example: createExample(),
	}
	cc.addFlags()
}

// addFlags adds flags for specific command.
func (cc *CreateCommand) addFlags() {
	flagSet := cc.cmd.Flags()
	flagSet.SetInterspersed(false)
	flagSet.StringVar(&cc.name, "name", "", "Specify name of container")
	flagSet.BoolVarP(&cc.tty, "tty", "t", false, "Allocate a tty device")
	flagSet.StringSliceVarP(&cc.volume, "volume", "v", nil, "Bind mount volumes to container")
	flagSet.StringVar(&cc.runtime, "runtime", "", "Specify oci runtime")
	flagSet.StringSliceVarP(&cc.env, "env", "e", nil, "Set environment variables for container")
	flagSet.StringSliceVarP(&cc.labels, "label", "l", nil, "Set label for a container")
	flagSet.StringVar(&cc.entrypoint, "entrypoint", "", "Overwrite the default entrypoint")
	flagSet.StringVarP(&cc.workdir, "workdir", "w", "", "Set the working directory in a container")
	flagSet.StringVarP(&cc.user, "user", "u", "", "UID")
	flagSet.StringVar(&cc.hostname, "hostname", "", "Set container's hostname")

	// cpu
	flagSet.Int64Var(&cc.cpushare, "cpu-share", 0, "CPU shares")
	flagSet.StringVar(&cc.cpusetcpus, "cpuset-cpus", "", "CPUs in cpuset")
	flagSet.StringVar(&cc.cpusetmems, "cpuset-mems", "", "MEMs in cpuset")

	// memory
	flagSet.Int64Var(&cc.memorySwappiness, "memory-wappiness", -1, "Container memory swappiness [0, 100]")
	flagSet.StringVarP(&cc.memory, "memory", "m", "", "Container memory limit")
	flagSet.StringVar(&cc.memorySwap, "memory-swap", "", "Container swap limit")
	// alios memory
	flagSet.Int64Var(&cc.memoryWmarkRatio, "memory-wmark-ratio", 0, "Represent this container's memory low water mark percentage, range in [0, 100]. The value of memory low water mark is memory.limit_in_bytes * MemoryWmarkRatio")
	flagSet.Int64Var(&cc.memoryExtra, "memory-extra", 0, "Represent container's memory high water mark percentage, range in [0, 100]")
	flagSet.Int64Var(&cc.memoryForceEmptyCtl, "memory-force-empty-ctl", 0, "Whether to reclaim page cache when deleting the cgroup of container")
	flagSet.Int64Var(&cc.scheLatSwitch, "sche-lat-switch", 0, "Whether to enable scheduler latency count in cpuacct")

	flagSet.StringSliceVarP(&cc.devices, "device", "", nil, "Add a host device to the container")
	flagSet.BoolVar(&cc.enableLxcfs, "enableLxcfs", false, "Enable lxcfs")
	flagSet.StringVar(&cc.restartPolicy, "restart", "", "Restart policy to apply when container exits")

	// namespace mode
	flagSet.StringVar(&cc.ipcMode, "ipc", "", "IPC namespace to use")
	flagSet.StringVar(&cc.pidMode, "pid", "", "PID namespace to use")
	flagSet.StringVar(&cc.utsMode, "uts", "", "UTS namespace to use")

	flagSet.StringSliceVar(&cc.sysctls, "sysctl", nil, "Sysctl options")
	flagSet.StringSliceVar(&cc.networks, "net", nil, "Set networks to container")
	flagSet.StringSliceVar(&cc.securityOpt, "security-opt", nil, "Security Options")

	// capabilities
	flagSet.BoolVar(&cc.privileged, "privileged", false, "Give extended privileges to the container")
	flagSet.StringSliceVar(&cc.capAdd, "cap-add", nil, "Add Linux capabilities")
	flagSet.StringSliceVar(&cc.capDrop, "cap-drop", nil, "Drop Linux capabilities")

	// blkio
	flagSet.Uint16Var(&cc.blkioWeight, "blkio-weight", 0, "Block IO (relative weight), between 10 and 1000, or 0 to disable")
	flagSet.Var(&cc.blkioWeightDevice, "blkio-weight-device", "Block IO weight (relative device weight)")
	flagSet.Var(&cc.blkioDeviceReadBps, "device-read-bps", "Limit read rate (bytes per second) from a device")
	flagSet.Var(&cc.blkioDeviceReadIOps, "device-read-iops", "Limit read rate (IO per second) from a device")
	flagSet.Var(&cc.blkioDeviceWriteBps, "device-write-bps", "Limit write rate (bytes per second) from a device")
	flagSet.Var(&cc.blkioDeviceWriteIOps, "device-write-iops", "Limit write rate (IO per second) from a device")

	// Intel RDT
	flagSet.StringVar(&cc.IntelRdtL3Cbm, "intel-rdt-l3-cbm", "", "Limit container resource for Intel RDT/CAT which introduced in Linux 4.10 kernel")
}

// runCreate is the entry of create command.
func (cc *CreateCommand) runCreate(args []string) error {
	config, err := cc.config()
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	config.Image = args[0]
	if len(args) > 1 {
		config.Cmd = args[1:]
	}
	containerName := cc.name

	ctx := context.Background()
	apiClient := cc.cli.Client()
	result, err := apiClient.ContainerCreate(ctx, config.ContainerConfig, config.HostConfig, config.NetworkingConfig, containerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	if len(result.Warnings) != 0 {
		fmt.Printf("WARNING: %s \n", strings.Join(result.Warnings, "\n"))
	}
	fmt.Printf("container ID: %s, name: %s \n", result.ID, result.Name)
	return nil
}

// createExample shows examples in create command, and is used in auto-generated cli docs.
func createExample() string {
	return `$ pouch create --name foo busybox:latest
container ID: e1d541722d68dc5d133cca9e7bd8fd9338603e1763096c8e853522b60d11f7b9, name: foo`
}
