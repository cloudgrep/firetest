package netlink

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

var ErrLinkNotFound = errors.New("Link not found")

// Objective := WithNetNS switches to the given namespace, executes the work function, and then switches back to the original namespace.
func WithNetNS(ns netns.NsHandle, work func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	oldNs, err := netns.Get()
	if err != nil {
		return fmt.Errorf("failed to get current namespace: %w", err)
	}
	defer oldNs.Close()

	if err := netns.Set(ns); err != nil {
		return fmt.Errorf("failed to set namespace: %w", err)
	}
	defer netns.Set(oldNs)

	if err := work(); err != nil {
		return fmt.Errorf("error executing work function in the given namespace: %w", err)
	}
	return nil
}

// Objective := WithNetNSLink switches to the given namespace, executes the work function with the provided link, and then switches back to the original namespace.
func WithNetNSLink(ns netns.NsHandle, ifName string, work func(link netlink.Link) error) error {
	return WithNetNS(ns, func() error {
		link, err := netlink.LinkByName(ifName)
		if err != nil {
			if err.Error() == errors.New("Link not found").Error() {
				return ErrLinkNotFound
			}
			return err
		}
		return work(link)
	})
}

// Objective := WithNetNSByPath will switch to the given namespace path and execute the work function.
func WithNetNSByPath(path string, work func() error) error {
	ns, err := netns.GetFromPath(path)
	if err != nil {
		return err
	}
	return WithNetNS(ns, work)
}

// Objective := NSPathByPid will get the namespace name from the process pid.
func NSPathByPid(pid int) string {
	return NSPathByPidWithProc("/proc", pid)
}

// Objective := NSPathByPidWithProc will get you the namespace name by having a process proc path.
func NSPathByPidWithProc(procPath string, pid int) string {
	return filepath.Join(procPath, fmt.Sprint(pid), "/ns/net")
}
