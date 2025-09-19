//go:build darwin || linux || freebsd || netbsd || openbsd
// +build darwin linux freebsd netbsd openbsd

package main

import (
    "fmt"
    "syscall"
    "golang.org/x/sys/unix"
)

func setUlimit() error {
    var rLimit syscall.Rlimit
    if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
        return err
    }
    maxFiles, err := unix.SysctlUint32("kern.maxfilesperproc")
    if err != nil {
        return fmt.Errorf("failed to get kern.maxfilesperproc: %v", err)
    }
    desired := uint64(4096)
    if desired > uint64(maxFiles) {
        desired = uint64(maxFiles)
    }
    if rLimit.Cur < desired {
        rLimit.Cur = desired
    }
    // fmt.Printf("ulimit setting soft limit to: %d (max allowed: %d)\n", rLimit.Cur, maxFiles)
    return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
}
