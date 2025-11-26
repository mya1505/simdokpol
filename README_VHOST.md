# Virtual Host Setup - SIMDOKPOL

## Deskripsi

Fitur Virtual Host Setup memungkinkan aplikasi SIMDOKPOL diakses melalui domain lokal yang lebih user-friendly (`simdokpol.local`) dibandingkan menggunakan `localhost:8080`.

## Fitur

✅ **Auto-detection**: Deteksi otomatis sistem operasi (Windows, Linux, macOS)  
✅ **Auto-setup**: Mencoba setup otomatis saat pertama kali dijalankan  
✅ **Manual fallback**: Memberikan instruksi manual jika setup otomatis gagal  
✅ **Permission check**: Memeriksa izin administrator/root sebelum modifikasi  
✅ **DNS flush**: Otomatis membersihkan DNS cache setelah setup  
✅ **Safe removal**: Fungsi untuk menghapus konfigurasi vhost dengan aman  

## Cara Kerja

### Automatic Setup (Direkomendasikan)

1. **Jalankan aplikasi pertama kali**
   - Aplikasi akan otomatis mendeteksi apakah virtual host sudah dikonfigurasi
   - Jika belum, akan mencoba setup otomatis

2. **Windows**
   - Aplikasi harus dijalankan sebagai Administrator
   - Entry akan ditambahkan ke `C:\Windows\System32\drivers\etc\hosts`
   - DNS cache akan di-flush dengan `ipconfig /flushdns`

3. **Linux**
   - Aplikasi memerlukan sudo/root access
   - Entry akan ditambahkan ke `/etc/hosts`
   - systemd-resolved akan direstart jika tersedia

4. **macOS**
   - Aplikasi memerlukan sudo/root access
   - Entry akan ditambahkan ke `/etc/hosts`
   - DNS cache akan di-flush dengan `dscacheutil`

### Manual Setup

Jika setup otomatis gagal (biasanya karena permission), aplikasi akan menampilkan instruksi manual:

#### Windows

```cmd
# Buka Command Prompt sebagai Administrator
notepad C:\Windows\System32\drivers\etc\hosts

# Tambahkan baris ini di akhir file:
127.0.0.1 simdokpol.local

# Simpan dan flush DNS
ipconfig /flushdns
```

#### Linux

```bash
# Buka Terminal
sudo sh -c 'echo "127.0.0.1 simdokpol.local" >> /etc/hosts'

# (Opsional) Restart network service
sudo systemctl restart systemd-resolved
```

#### macOS

```bash
# Buka Terminal
sudo sh -c 'echo "127.0.0.1 simdokpol.local" >> /etc/hosts'

# Flush DNS cache
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

## Struktur File

```
simdokpol/
├── internal/
│   └── utils/
│       └── vhost_setup.go    # Utility untuk setup virtual host
└── cmd/
    └── main.go               # Integrasi dengan aplikasi utama
```

## Kode Implementasi

### File: `internal/utils/vhost_setup.go`

File ini berisi:
- `VHostSetup` struct untuk mengelola konfigurasi
- `IsSetup()` - Cek apakah vhost sudah dikonfigurasi
- `Setup()` - Setup virtual host otomatis
- `Remove()` - Hapus konfigurasi virtual host
- `checkPermission()` - Validasi permission
- `addHostsEntry()` - Tambah entry ke hosts file
- `flushDNSCache()` - Bersihkan DNS cache
- `showManualInstructions()` - Tampilkan instruksi manual

### Integrasi di `cmd/main.go`

```go
// Di main()
vhostSetup = utils.NewVHostSetup()
setupVirtualHost()

// Tentukan URL dinamis
isSetup, _ := vhostSetup.IsSetup()
if isSetup {
    appURL = vhostSetup.GetURL("8080")
} else {
    appURL = defaultURL
}
```

## Testing

### Test Setup

```bash
# Build aplikasi
go build -o simdokpol cmd/main.go

# Windows: Jalankan sebagai Administrator
./simdokpol.exe

# Linux/macOS: Jalankan dengan sudo (untuk first-time setup)
sudo ./simdokpol
```

### Verifikasi Setup

```bash
# Windows
type C:\Windows\System32\drivers\etc\hosts | findstr simdokpol

# Linux/macOS
cat /etc/hosts | grep simdokpol

# Test akses
ping simdokpol.local
curl http://simdokpol.local:8080
```

### Test Remove

```go
// Tambahkan fungsi CLI untuk testing
func removeVHost() {
    vhost := utils.NewVHostSetup()
    if err := vhost.Remove(); err != nil {
        log.Fatal(err)
    }
}
```

## Troubleshooting

### Problem: "Permission Denied"

**Solusi**:
- **Windows**: Jalankan aplikasi sebagai Administrator (klik kanan > Run as Administrator)
- **Linux/macOS**: Jalankan dengan sudo: `sudo ./simdokpol`

### Problem: "Domain tidak bisa diakses setelah setup"

**Solusi**:
```bash
# Windows
ipconfig /flushdns

# macOS
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder

# Linux
sudo systemctl restart systemd-resolved

# Atau restart browser
```

### Problem: "File hosts terproteksi"

**Solusi Linux/macOS**:
```bash
# Cek permission
ls -l /etc/hosts

# Jika perlu, ubah permission sementara
sudo chmod 644 /etc/hosts
```

**Solusi Windows**:
- Disable antivirus sementara
- Cek Windows Defender protection

## Keamanan

### Best Practices

1. **Backup hosts file** sebelum modifikasi:
   ```bash
   # Windows
   copy C:\Windows\System32\drivers\etc\hosts hosts.backup
   
   # Linux/macOS
   sudo cp /etc/hosts /etc/hosts.backup
   ```

2. **Validasi entry** sebelum menulis:
   - Cek format IP dan domain valid
   - Hindari duplikasi entry
   - Gunakan komentar untuk identifikasi

3. **Minimal permission**:
   - Hanya minta elevated permission saat diperlukan
   - Informasikan user kenapa permission diperlukan

### Risiko

⚠️ **Modifikasi hosts file** memerlukan elevated permission  
⚠️ **Malware** bisa menyalahgunakan hosts file untuk redirect  
⚠️ **Backup** penting untuk recovery jika terjadi error  

## Kustomisasi

### Mengubah Domain Default

Edit di `internal/utils/vhost_setup.go`:

```go
const (
    LocalDomain = "myapp.local"  // Ubah sesuai keinginan
    LocalIP     = "127.0.0.1"
)
```

### Menambah Multiple Domain

```go
func (v *VHostSetup) Setup() error {
    domains := []string{
        "simdokpol.local",
        "www.simdokpol.local",
        "app.simdokpol.local",
    }
    
    for _, domain := range domains {
        v.domain = domain
        if err := v.addHostsEntry(); err != nil {
            return err
        }
    }
    return nil
}
```

## Future Enhancements

- [ ] Support untuk IPv6
- [ ] GUI untuk setup manual
- [ ] Backup otomatis hosts file
- [ ] Multi-domain support
- [ ] Integration dengan SSL/TLS certificate
- [ ] Auto-remove saat uninstall

## Lisensi

Bagian dari aplikasi SIMDOKPOL

---

**Catatan**: Pastikan untuk menjalankan aplikasi dengan permission yang sesuai saat pertama kali untuk memastikan virtual host dapat dikonfigurasi dengan benar.