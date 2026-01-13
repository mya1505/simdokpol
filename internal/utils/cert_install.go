package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func InstallCertificate(certPath string) error {
	switch runtime.GOOS {
	case "windows":
		return runCertCommand("certutil", "-addstore", "-f", "Root", certPath)
	case "darwin":
		return runCertCommand("security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", certPath)
	case "linux":
		return installCertificateLinux(certPath)
	default:
		return fmt.Errorf("platform tidak didukung untuk auto-install")
	}
}

func installCertificateLinux(certPath string) error {
	if _, err := os.Stat("/etc/ca-certificates/trust-source/anchors"); err == nil {
		dest := "/etc/ca-certificates/trust-source/anchors/simdokpol.crt"
		if err := copyFile(certPath, dest); err != nil {
			return err
		}
		return runCertCommand("update-ca-trust")
	}

	dest := "/usr/local/share/ca-certificates/simdokpol.crt"
	if err := copyFile(certPath, dest); err != nil {
		return err
	}
	return runCertCommand("update-ca-certificates")
}

func runCertCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s gagal: %v (%s)", name, err, string(output))
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
