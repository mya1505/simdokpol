package services

// Variabel ini akan diisi saat runtime (dari ENV atau LDFLAGS)
// JANGAN ISI HARDCODED STRING DI SINI!
var (
	AppSecretKeyString string 
	JWTSecretKey       []byte
)