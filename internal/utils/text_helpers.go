package utils

import (
	"strconv" // <-- PERBAIKAN: TAMBAHKAN IMPORT INI
)

// Daftar kata untuk konversi
var (
	satuan = []string{"", "satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan"}
	belasan = []string{"sepuluh", "sebelas", "dua belas", "tiga belas", "empat belas", "lima belas", "enam belas", "tujuh belas", "delapan belas", "sembilan belas"}
	puluhan = []string{"", "sepuluh", "dua puluh", "tiga puluh", "empat puluh", "lima puluh", "enam puluh", "tujuh puluh", "delapan puluh", "sembilan puluh"}
)

// IntToIndonesianWords mengonversi angka integer (0-999) menjadi teks bahasa Indonesia.
// Fungsi ini akan mengembalikan teks dalam huruf kecil, sesuai format di surat.
func IntToIndonesianWords(num int) string {
	if num < 0 {
		return "minus " + IntToIndonesianWords(-num)
	}
	if num == 0 {
		return "nol"
	}

	if num < 10 {
		// Kasus 1-9
		return satuan[num]
	}
	if num < 20 {
		// Kasus 10-19
		return belasan[num-10]
	}
	if num < 100 {
		// Kasus 20-99
		sisa := num % 10
		if sisa == 0 {
			// Kasus 20, 30, 40, ...
			return puluhan[num/10]
		}
		// Kasus 21, 22, ..., 99
		return puluhan[num/10] + " " + satuan[sisa]
	}
	
	if num < 200 {
		// Kasus 100-199
		if num == 100 { return "seratus" }
		return "seratus " + IntToIndonesianWords(num-100)
	}
	if num < 1000 {
		// Kasus 200-999
		sisa := num % 100
		if sisa == 0 {
			return satuan[num/100] + " ratus"
		}
		return satuan[num/100] + " ratus " + IntToIndonesianWords(sisa)
	}

	// Fallback jika angka terlalu besar
	return strconv.Itoa(num)
}