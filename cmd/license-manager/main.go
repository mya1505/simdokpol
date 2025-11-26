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
	//"io/ioutil"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	//"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// --- RAHASIA DAPUR (HARUS SAMA DENGAN APLIKASI UTAMA) ---
// 1. Payload untuk ECDSA (jika masih pakai) / Payload umum
const licensePayload = "SIMDOKPOL-PRO-LICENSE-PAYLOAD"

// 2. Secret Key untuk HMAC (Yang kita pakai di update terakhir)
// PASTIKAN INI SAMA DENGAN 'appSecretKey' di 'internal/services/license_service.go'
const appSecretKey = "GANTI_STRING_INI_DENGAN_KATA_SANDI_RAHASIA_YANG_PANJANG_DAN_RUMIT_12345"

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("SIMDOKPOL License Manager (Developer Only)")
	myWindow.Resize(fyne.NewSize(600, 500))

	// ============================================================
	// TAB 1: HMAC SERIAL GENERATOR (METODE UTAMA SAAT INI)
	// ============================================================
	
	lblHwid := widget.NewLabel("Masukkan Hardware ID User:")
	entryHwid := widget.NewEntry()
	entryHwid.SetPlaceHolder("Contoh: A1B2-C3D4-E5F6-G7H8")

	lblResult := widget.NewLabel("Serial Key (Untuk User):")
	entrySerial := widget.NewEntry()
	entrySerial.SetPlaceHolder("Serial key akan muncul di sini...")
	entrySerial.Disable() // Read only

	btnGenerateSerial := widget.NewButtonWithIcon("Generate Serial Key", theme.ConfirmIcon(), func() {
		hwid := strings.TrimSpace(entryHwid.Text)
		if hwid == "" {
			dialog.ShowError(fmt.Errorf("Hardware ID tidak boleh kosong"), myWindow)
			return
		}

		// --- LOGIKA HMAC (SAMA DENGAN APP UTAMA) ---
		h := hmac.New(sha256.New, []byte(appSecretKey))
		h.Write([]byte(hwid))
		hash := h.Sum(nil)

		// Ambil 15 byte
		truncatedHash := hash[:15]
		// Encode Base32
		rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)

		// Format XXXXX-XXXXX
		var formattedKey strings.Builder
		for i, r := range rawKey {
			if i > 0 && i%5 == 0 {
				formattedKey.WriteRune('-')
			}
			formattedKey.WriteRune(r)
		}
		// -------------------------------------------

		entrySerial.SetText(formattedKey.String())
	})

	btnCopySerial := widget.NewButtonWithIcon("Copy Serial", theme.ContentCopyIcon(), func() {
		myWindow.Clipboard().SetContent(entrySerial.Text)
	})

	tabHmac := container.NewVBox(
		widget.NewLabelWithStyle("Generator Lisensi (HMAC-SHA256)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		lblHwid,
		entryHwid,
		widget.NewSeparator(),
		btnGenerateSerial,
		widget.NewSeparator(),
		lblResult,
		entrySerial,
		btnCopySerial,
	)


	// ============================================================
	// TAB 2: ECDSA KEY PAIR GENERATOR (OPSIONAL / BACKUP)
	// ============================================================
	
	entryPrivate := widget.NewMultiLineEntry()
	entryPrivate.SetPlaceHolder("Private Key (PEM)")
	entryPublic := widget.NewMultiLineEntry()
	entryPublic.SetPlaceHolder("Public Key (PEM)")

	btnGenKeys := widget.NewButton("Generate New Key Pair (ECDSA)", func() {
		privKey, pubKeyPEM, privKeyPEM, err := generateECDSAKeys()
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		_ = privKey // unused in GUI display directly
		entryPrivate.SetText(string(privKeyPEM))
		entryPublic.SetText(string(pubKeyPEM))
	})

	btnSaveKeys := widget.NewButton("Simpan ke File (private.pem & public.pem)", func() {
		if entryPrivate.Text == "" {
			dialog.ShowError(fmt.Errorf("Generate kunci terlebih dahulu"), myWindow)
			return
		}
		err := os.WriteFile("private.pem", []byte(entryPrivate.Text), 0600)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		err = os.WriteFile("public.pem", []byte(entryPublic.Text), 0644)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		dialog.ShowInformation("Sukses", "Kunci berhasil disimpan di folder aplikasi ini.", myWindow)
	})

	tabEcdsa := container.NewVBox(
		widget.NewLabel("Tool ini untuk membuat pasangan kunci kriptografi baru."),
		widget.NewLabel("Gunakan HANYA jika Anda ingin mereset sistem keamanan aplikasi."),
		btnGenKeys,
		container.NewGridWithColumns(2, 
			container.NewPadded(entryPrivate), 
			container.NewPadded(entryPublic),
		),
		btnSaveKeys,
	)


	// ============================================================
	// MAIN LAYOUT
	// ============================================================
	
	tabs := container.NewAppTabs(
		container.NewTabItem("License Generator", container.NewPadded(tabHmac)),
		container.NewTabItem("Key Management (Advanced)", container.NewPadded(tabEcdsa)),
	)

	// Tambahkan footer info
	footer := widget.NewLabelWithStyle("SIMDOKPOL Developer Tool v1.0", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	content := container.NewBorder(nil, footer, nil, nil, tabs)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

// Helper untuk ECDSA (Logic lama, disimpan untuk referensi/kebutuhan masa depan)
func generateECDSAKeys() (*ecdsa.PrivateKey, []byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, nil, err
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyBytes})

	return privateKey, publicKeyPEM, privateKeyPEM, nil
}
