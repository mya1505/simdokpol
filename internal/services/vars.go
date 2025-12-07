package services

// Variabel ini KOSONG. Akan diisi oleh Compiler (ldflags) saat build release.
var (
	AppSecretKeyString string 
	JWTSecretKeyString string // <-- Ini yang dicari sama error tadi
	
	// Ini yang dipakai aplikasi (akan diisi dari string di atas saat runtime)
	JWTSecretKey       []byte
)