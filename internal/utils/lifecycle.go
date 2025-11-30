package utils

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// RestartApp me-restart aplikasi saat ini (Cross-Platform)
func RestartApp() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return err
	}

	log.Println("🔄 SYSTEM RESTART TRIGGERED...")

	// Coba exec (Unix)
	err = syscall.Exec(exePath, os.Args, os.Environ())
	if err != nil {
		// Fallback (Windows/Others)
		cmd := exec.Command(exePath, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Start(); err != nil {
			return err
		}
		os.Exit(0)
	}
	return nil
}