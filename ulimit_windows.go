// ulimit_windows.go
//go:build windows
// +build windows

package main

func setUlimit() error {
    // no-op on Windows
    return nil
}

