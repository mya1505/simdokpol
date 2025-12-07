package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"simdokpol/internal/utils" // Import Utils

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv" // Import Dotenv
)

// Variabel ini KOSONG. Harus diisi via LDFLAGS atau .env
var appSecretKey = ""

func main() {
	// --- 1. LOAD .ENV OTOMATIS ---
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	_ = godotenv.Overload(envPath)

	// --- 2. VALIDASI SECRET KEY ---
	// Prioritas: LDFLAGS -> ENV (.env)
	if appSecretKey == "" {
		appSecretKey = os.Getenv("APP_SECRET_KEY")
	}

	isKeyValid := appSecretKey != ""

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
		warnText := canvas.NewText(fmt.Sprintf("ERROR: Secret Key tidak ditemukan di %s", envPath), color.RGBA{R: 255, G: 0, B: 0, A: 255})
		warnText.TextStyle = fyne.TextStyle{Bold: true}
		warnText.Alignment = fyne.TextAlignCenter
		warnText.TextSize = 10
		statusContainer = container.New(layout.NewCenterLayout(), warnText)
	} else {
		// Info Key Loaded
		keyInfo := canvas.NewText(fmt.Sprintf("Key Loaded: %s... (Valid)", sha256Sum(appSecretKey)), color.RGBA{R: 0, G: 128, B: 0, A: 255})
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

		// Logic Generate
		h := hmac.New(sha256.New, []byte(appSecretKey))
		h.Write([]byte(hwid))
		hash := h.Sum(nil)

		truncatedHash := hash[:15]
		rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)

		var formattedKey strings.Builder
		for i, r := range rawKey {
			if i > 0 && i%5 == 0 {
				formattedKey.WriteRune('-')
			}
			formattedKey.WriteRune(r)
		}

		entryResult.SetText(formattedKey.String())
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

func sha256Sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}