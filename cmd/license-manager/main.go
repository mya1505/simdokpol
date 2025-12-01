package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Variabel ini akan DI-TIMPA oleh Makefile saat build release
var appSecretKey = "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA"

func main() {
	a := app.New()
	w := a.NewWindow("SIMDOKPOL License Manager")
	w.Resize(fyne.NewSize(500, 300))

	lblInfo := widget.NewLabel("Generator Lisensi Pro (Admin Only)")
	lblInfo.Alignment = fyne.TextAlignCenter
	lblInfo.TextStyle = fyne.TextStyle{Bold: true}

	entryHwid := widget.NewEntry()
	entryHwid.SetPlaceHolder("Masukkan Hardware ID User...")

	entryResult := widget.NewEntry()
	entryResult.SetPlaceHolder("Serial Key akan muncul di sini...")
	entryResult.Disable()

	btnGenerate := widget.NewButtonWithIcon("Generate Serial Key", theme.ConfirmIcon(), func() {
		hwid := strings.TrimSpace(entryHwid.Text)
		if hwid == "" {
			dialog.ShowError(fmt.Errorf("Hardware ID kosong"), w)
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

	btnCopy := widget.NewButtonWithIcon("Salin Key", theme.ContentCopyIcon(), func() {
		w.Clipboard().SetContent(entryResult.Text)
	})

	content := container.NewVBox(
		lblInfo,
		widget.NewSeparator(),
		widget.NewLabel("Hardware ID:"),
		entryHwid,
		btnGenerate,
		widget.NewSeparator(),
		widget.NewLabel("Result Serial Key:"),
		entryResult,
		btnCopy,
	)

	w.SetContent(container.NewPadded(content))
	w.ShowAndRun()
}