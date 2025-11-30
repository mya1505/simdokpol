package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// --- RAHASIA DAPUR ---
// FIX: Gunakan 'var' agar bisa di-inject via -ldflags di Makefile
var appSecretKey = "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA"

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("SIMDOKPOL License Manager (Admin)")
	myWindow.Resize(fyne.NewSize(600, 500))

	// ============================================================
	// TAB 1: HMAC SERIAL GENERATOR (METODE UTAMA)
	// ============================================================

	lblInfo := widget.NewLabel("Tools ini menggunakan Secret Key yang di-inject saat build.\nPastikan Anda menggunakan file .exe hasil build Makefile.")
	lblInfo.Wrapping = fyne.TextWrapWord

	lblHwid := widget.NewLabel("Masukkan Hardware ID User:")
	entryHwid := widget.NewEntry()
	entryHwid.SetPlaceHolder("Contoh: A1B2-C3D4-E5F6-G7H8")

	lblResult := widget.NewLabel("Serial Key (Untuk User):")
	entrySerial := widget.NewEntry()
	entrySerial.SetPlaceHolder("Serial key akan muncul di sini...")
	entrySerial.Disable() 

	btnGenerateSerial := widget.NewButtonWithIcon("Generate Serial Key", theme.ConfirmIcon(), func() {
		hwid := strings.TrimSpace(entryHwid.Text)
		if hwid == "" {
			dialog.ShowError(fmt.Errorf("Hardware ID tidak boleh kosong"), myWindow)
			return
		}

		// --- LOGIKA HMAC ---
		h := hmac.New(sha256.New, []byte(appSecretKey))
		h.Write([]byte(hwid))
		hash := h.Sum(nil)

		// Ambil 15 byte & Encode Base32
		truncatedHash := hash[:15]
		rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)

		// Format XXXXX-XXXXX
		var formattedKey strings.Builder
		for i, r := range rawKey {
			if i > 0 && i%5 == 0 {
				formattedKey.WriteRune('-')
			}
			formattedKey.WriteRune(r)
		}
		// -------------------

		entrySerial.SetText(formattedKey.String())
	})

	btnCopySerial := widget.NewButtonWithIcon("Copy Serial", theme.ContentCopyIcon(), func() {
		myWindow.Clipboard().SetContent(entrySerial.Text)
	})

	tabHmac := container.NewVBox(
		widget.NewLabelWithStyle("Generator Lisensi (HMAC-SHA256)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		lblInfo,
		widget.NewSeparator(),
		lblHwid,
		entryHwid,
		btnGenerateSerial,
		widget.NewSeparator(),
		lblResult,
		entrySerial,
		btnCopySerial,
	)

	// ============================================================
	// TAB 2: ECDSA KEY GEN (BACKUP/LEGACY)
	// ============================================================

	entryPrivate := widget.NewMultiLineEntry()
	entryPrivate.SetPlaceHolder("Private Key (PEM)")
	entryPublic := widget.NewMultiLineEntry()
	entryPublic.SetPlaceHolder("Public Key (PEM)")

	btnGenKeys := widget.NewButton("Generate New Key Pair (ECDSA)", func() {
		_, pubKeyPEM, privKeyPEM, err := generateECDSAKeys()
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		entryPrivate.SetText(string(privKeyPEM))
		entryPublic.SetText(string(pubKeyPEM))
	})

	btnSaveKeys := widget.NewButton("Simpan ke File", func() {
		if entryPrivate.Text == "" {
			dialog.ShowError(fmt.Errorf("Generate kunci terlebih dahulu"), myWindow)
			return
		}
		_ = os.WriteFile("private.pem", []byte(entryPrivate.Text), 0600)
		_ = os.WriteFile("public.pem", []byte(entryPublic.Text), 0644)
		dialog.ShowInformation("Sukses", "Kunci disimpan: private.pem & public.pem", myWindow)
	})

	tabEcdsa := container.NewVBox(
		widget.NewLabel("Buat pasangan kunci kriptografi baru (ECDSA P-256)."),
		widget.NewLabel("Hanya gunakan jika ingin mereset sistem keamanan."),
		btnGenKeys,
		container.NewGridWithColumns(2, 
			container.NewPadded(entryPrivate), 
			container.NewPadded(entryPublic),
		),
		btnSaveKeys,
	)

	// ============================================================
	// LAYOUT
	// ============================================================

	tabs := container.NewAppTabs(
		container.NewTabItem("License Generator", container.NewPadded(tabHmac)),
		container.NewTabItem("Key Management", container.NewPadded(tabEcdsa)),
	)

	footer := widget.NewLabelWithStyle("SIMDOKPOL Admin Tool v1.0", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	content := container.NewBorder(nil, footer, nil, nil, tabs)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

// Helper ECDSA
func generateECDSAKeys() (*ecdsa.PrivateKey, []byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil { return nil, nil, nil, err }

	privateKeyBytes, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes})

	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyBytes})

	return privateKey, publicKeyPEM, privateKeyPEM, nil
}