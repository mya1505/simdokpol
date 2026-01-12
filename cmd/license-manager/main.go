package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"simdokpol/internal/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	keyPath := resolvePrivateKeyPath("")
	privateKey, keyHash, err := loadPrivateKey(keyPath)
	isKeyValid := err == nil

	a := app.New()
	w := a.NewWindow("SIMDOKPOL License Manager")
	w.Resize(fyne.NewSize(500, 350))

	// --- Header ---
	lblInfo := widget.NewLabel("Generator Lisensi Pro")
	lblInfo.Alignment = fyne.TextAlignCenter
	lblInfo.TextStyle = fyne.TextStyle{Bold: true}

	// --- Form Inputs ---
	entryHwid := widget.NewEntry()
	entryHwid.SetPlaceHolder("Masukkan Hardware ID User (Format: XXXX-XXXX-...)")

	entryResult := widget.NewEntry()
	entryResult.SetPlaceHolder("Serial Key akan muncul di sini...")
	entryResult.Disable()

	// --- Status Warning jika Key Kosong ---
	var statusContainer *fyne.Container
	if !isKeyValid {
		warnText := canvas.NewText(fmt.Sprintf("ERROR: Private key tidak ditemukan di %s", keyPath), color.RGBA{R: 255, G: 0, B: 0, A: 255})
		warnText.TextStyle = fyne.TextStyle{Bold: true}
		warnText.Alignment = fyne.TextAlignCenter
		warnText.TextSize = 10
		statusContainer = container.New(layout.NewCenterLayout(), warnText)
	} else {
		// Info Key Loaded
		keyInfo := canvas.NewText(fmt.Sprintf("Key Loaded: %s... (Valid)", keyHash), color.RGBA{R: 0, G: 128, B: 0, A: 255})
		keyInfo.TextSize = 10
		keyInfo.Alignment = fyne.TextAlignCenter
		statusContainer = container.New(layout.NewCenterLayout(), keyInfo)
	}

	// --- Tombol Generate ---
	btnGenerate := widget.NewButtonWithIcon("Generate Serial Key", theme.ConfirmIcon(), func() {
		hwid := strings.TrimSpace(entryHwid.Text)
		if hwid == "" {
			dialog.ShowError(fmt.Errorf("Hardware ID tidak boleh kosong"), w)
			return
		}

		formattedKey, signErr := utils.SignActivationKey(hwid, privateKey)
		if signErr != nil {
			dialog.ShowError(fmt.Errorf("Gagal generate key: %w", signErr), w)
			return
		}

		entryResult.SetText(formattedKey)
	})
	btnGenerate.Importance = widget.HighImportance

	if !isKeyValid {
		btnGenerate.Disable()
		entryHwid.Disable()
	}

	// --- Tombol Copy ---
	btnCopy := widget.NewButtonWithIcon("Salin Key ke Clipboard", theme.ContentCopyIcon(), func() {
		if entryResult.Text != "" {
			w.Clipboard().SetContent(entryResult.Text)
			dialog.ShowInformation("Sukses", "Serial Key disalin!", w)
		}
	})

	// --- Layout ---
	formContent := container.NewVBox(
		lblInfo,
		statusContainer,
		widget.NewSeparator(),
		widget.NewLabel("Hardware ID User:"),
		entryHwid,
		layout.NewSpacer(),
		btnGenerate,
		widget.NewSeparator(),
		widget.NewLabel("Hasil Serial Key:"),
		entryResult,
		btnCopy,
	)

	w.SetContent(container.NewPadded(formContent))
	w.ShowAndRun()
}

func sha256Sum(b []byte) string {
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:8])
}

func resolvePrivateKeyPath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}

	if envPath := os.Getenv("LICENSE_PRIVATE_KEY"); envPath != "" {
		return envPath
	}

	if _, err := os.Stat("private.pem"); err == nil {
		return "private.pem"
	}

	return filepath.Join(utils.GetAppDataDir(), "private.pem")
}

func loadPrivateKey(path string) (*ecdsa.PrivateKey, string, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	privateKey, err := utils.ParsePrivateKeyPEM(pemBytes)
	if err != nil {
		return nil, "", err
	}

	return privateKey, sha256Sum(pemBytes), nil
}
