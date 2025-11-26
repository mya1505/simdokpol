/**
 * FILE: internal/utils/vhost_setup.go
 * 
 * PURPOSE:
 * Utility untuk setup virtual host di multiplatform (Windows, Linux, macOS)
 * Menambahkan entry hosts file untuk domain lokal aplikasi
 */
package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// Domain lokal yang akan digunakan
	LocalDomain = "simdokpol.local"
	LocalIP     = "127.0.0.1"
)

// VHostSetup struct untuk mengelola setup virtual host
type VHostSetup struct {
	hostsFile string
	domain    string
	ip        string
}

// NewVHostSetup membuat instance baru VHostSetup
func NewVHostSetup() *VHostSetup {
	return &VHostSetup{
		hostsFile: getHostsFilePath(),
		domain:    LocalDomain,
		ip:        LocalIP,
	}
}

// getHostsFilePath mengembalikan path file hosts berdasarkan OS
func getHostsFilePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts")
	case "linux", "darwin":
		return "/etc/hosts"
	default:
		log.Printf("PERINGATAN: Sistem operasi %s tidak didukung untuk setup vhost", runtime.GOOS)
		return ""
	}
}

// IsSetup memeriksa apakah virtual host sudah dikonfigurasi
func (v *VHostSetup) IsSetup() (bool, error) {
	if v.hostsFile == "" {
		return false, fmt.Errorf("hosts file tidak didukung untuk OS ini")
	}

	file, err := os.Open(v.hostsFile)
	if err != nil {
		return false, fmt.Errorf("gagal membuka file hosts: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip komentar
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Cek apakah baris mengandung IP dan domain kita
		if strings.Contains(line, v.ip) && strings.Contains(line, v.domain) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error membaca file hosts: %w", err)
	}

	return false, nil
}

// Setup melakukan setup virtual host
func (v *VHostSetup) Setup() error {
	// Cek apakah sudah setup
	isSetup, err := v.IsSetup()
	if err != nil {
		return err
	}

	if isSetup {
		log.Printf("INFO: Virtual host %s sudah dikonfigurasi", v.domain)
		return nil
	}

	log.Printf("INFO: Memulai setup virtual host untuk domain: %s", v.domain)

	// Cek permission sebelum menulis
	if err := v.checkPermission(); err != nil {
		log.Printf("PERINGATAN: Tidak dapat setup virtual host secara otomatis: %v", err)
		v.showManualInstructions()
		return err
	}

	// Tambahkan entry ke hosts file
	if err := v.addHostsEntry(); err != nil {
		log.Printf("ERROR: Gagal menambahkan entry ke hosts file: %v", err)
		v.showManualInstructions()
		return err
	}

	log.Printf("INFO: Virtual host berhasil dikonfigurasi!")
	log.Printf("INFO: Anda sekarang dapat mengakses aplikasi melalui: http://%s:8080", v.domain)

	return nil
}

// checkPermission memeriksa apakah memiliki permission untuk menulis hosts file
func (v *VHostSetup) checkPermission() error {
	// Coba buka file dengan mode append untuk test write permission
	file, err := os.OpenFile(v.hostsFile, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("tidak memiliki izin administrator/root untuk memodifikasi file hosts")
		}
		return err
	}
	file.Close()
	return nil
}

// addHostsEntry menambahkan entry baru ke hosts file
func (v *VHostSetup) addHostsEntry() error {
	// Baca file hosts terlebih dahulu
	content, err := os.ReadFile(v.hostsFile)
	if err != nil {
		return fmt.Errorf("gagal membaca file hosts: %w", err)
	}

	// Siapkan entry baru
	newEntry := fmt.Sprintf("\n# SIMDOKPOL - Ditambahkan secara otomatis pada %s\n%s %s\n",
		getTimestamp(), v.ip, v.domain)

	// Append entry baru
	newContent := string(content) + newEntry

	// Tulis kembali ke file
	if err := os.WriteFile(v.hostsFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("gagal menulis ke file hosts: %w", err)
	}

	// Flush DNS cache
	v.flushDNSCache()

	return nil
}

// flushDNSCache membersihkan DNS cache setelah modifikasi hosts file
func (v *VHostSetup) flushDNSCache() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ipconfig", "/flushdns")
	case "darwin":
		cmd = exec.Command("dscacheutil", "-flushcache")
		// macOS 10.10.4+ juga butuh perintah tambahan
		exec.Command("killall", "-HUP", "mDNSResponder").Run()
	case "linux":
		// Linux biasanya tidak perlu flush DNS untuk hosts file
		// Tapi beberapa distro menggunakan systemd-resolved
		cmd = exec.Command("systemctl", "restart", "systemd-resolved")
	}

	if cmd != nil {
		if err := cmd.Run(); err != nil {
			log.Printf("PERINGATAN: Gagal flush DNS cache (ini mungkin normal): %v", err)
		} else {
			log.Println("INFO: DNS cache berhasil dibersihkan")
		}
	}
}

// showManualInstructions menampilkan instruksi manual jika setup otomatis gagal
func (v *VHostSetup) showManualInstructions() {
	log.Println("\n" + strings.Repeat("=", 70))
	log.Println("INSTRUKSI SETUP MANUAL VIRTUAL HOST")
	log.Println(strings.Repeat("=", 70))

	switch runtime.GOOS {
	case "windows":
		log.Println("1. Buka Command Prompt atau PowerShell sebagai Administrator")
		log.Println("2. Jalankan perintah berikut:")
		log.Printf("   echo %s %s >> %s\n", v.ip, v.domain, v.hostsFile)
		log.Println("\n   ATAU edit manual dengan notepad:")
		log.Printf("   notepad %s\n", v.hostsFile)
		log.Println("   Kemudian tambahkan baris berikut di akhir file:")
		log.Printf("   %s %s\n", v.ip, v.domain)
		log.Println("\n3. Simpan file (jika menggunakan notepad)")
		log.Println("4. Flush DNS cache dengan perintah:")
		log.Println("   ipconfig /flushdns")

	case "darwin":
		log.Println("1. Buka Terminal")
		log.Println("2. Jalankan perintah:")
		log.Printf("   sudo sh -c 'echo \"%s %s\" >> %s'\n", v.ip, v.domain, v.hostsFile)
		log.Println("3. Masukkan password administrator Anda")
		log.Println("4. Flush DNS cache:")
		log.Println("   sudo dscacheutil -flushcache")
		log.Println("   sudo killall -HUP mDNSResponder")

	case "linux":
		log.Println("1. Buka Terminal")
		log.Println("2. Jalankan perintah:")
		log.Printf("   sudo sh -c 'echo \"%s %s\" >> %s'\n", v.ip, v.domain, v.hostsFile)
		log.Println("3. Masukkan password sudo Anda")
		log.Println("4. (Opsional) Restart network service atau systemd-resolved jika diperlukan")
	}

	log.Println("\nSetelah setup manual selesai, restart aplikasi SIMDOKPOL.")
	log.Printf("Anda akan dapat mengakses aplikasi melalui: http://%s:8080\n", v.domain)
	log.Println(strings.Repeat("=", 70) + "\n")
}

// Remove menghapus entry virtual host dari hosts file
func (v *VHostSetup) Remove() error {
	log.Printf("INFO: Menghapus virtual host %s dari hosts file...", v.domain)

	// Cek permission
	if err := v.checkPermission(); err != nil {
		return fmt.Errorf("tidak memiliki izin untuk memodifikasi hosts file: %w", err)
	}

	// Baca file hosts
	file, err := os.Open(v.hostsFile)
	if err != nil {
		return fmt.Errorf("gagal membuka file hosts: %w", err)
	}
	defer file.Close()

	// Baca dan filter baris
	var newLines []string
	scanner := bufio.NewScanner(file)
	skipNext := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip baris yang mengandung domain kita atau komentar SIMDOKPOL
		if strings.Contains(trimmed, v.domain) || strings.Contains(trimmed, "# SIMDOKPOL") {
			skipNext = true
			continue
		}

		if skipNext && trimmed == "" {
			skipNext = false
			continue
		}

		newLines = append(newLines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error membaca file hosts: %w", err)
	}

	// Tulis kembali file hosts tanpa entry SIMDOKPOL
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(v.hostsFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("gagal menulis file hosts: %w", err)
	}

	// Flush DNS cache
	v.flushDNSCache()

	log.Printf("INFO: Virtual host %s berhasil dihapus", v.domain)
	return nil
}

// GetDomain mengembalikan domain yang dikonfigurasi
func (v *VHostSetup) GetDomain() string {
	return v.domain
}

// GetURL mengembalikan URL lengkap aplikasi
func (v *VHostSetup) GetURL(port string) string {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	return fmt.Sprintf("http://%s%s", v.domain, port)
}

// getTimestamp helper function untuk mendapatkan timestamp
func getTimestamp() string {
	now := time.Now()
	return now.Format("2006-01-02 15:04:05")
}