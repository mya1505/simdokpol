# Sistem Informasi Manajemen Dokumen Kepolisian (SIMDOKPOL)

![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue.svg)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey.svg)
![Database](https://img.shields.io/badge/Database-SQLite%20%7C%20MySQL%20%7C%20PostgreSQL-blue.svg)

**SIMDOKPOL** adalah aplikasi desktop *cross-platform* yang dirancang untuk membantu unit kepolisian dalam manajemen dan penerbitan surat keterangan secara efisien, cepat, dan aman.

Aplikasi ini dapat berjalan dalam dua mode:
1. **Standalone (100% Offline):** Menggunakan database **SQLite** yang portabel, ideal untuk penggunaan di satu komputer.
2. **Client-Server (Jaringan):** Dapat terhubung ke database terpusat (**MySQL** atau **PostgreSQL**) untuk penggunaan bersama di jaringan kantor.

![Dasbor SIMDOKPOL](.github/assets/guide-dashboard.png)

---

## ✨ Fitur Utama

- **Multi-Database**: Mendukung SQLite, MySQL, dan PostgreSQL.
- **Setup Wizard 5-Langkah**: Konfigurasi awal yang terpandu, termasuk tes koneksi database.
- **Aplikasi Desktop Standalone**: Berjalan sebagai aplikasi mandiri dengan ikon di *system tray* dan notifikasi *native*.
- **Kustomisasi Template Barang**: (Fitur Pro) Mengatur *item* barang hilang (KTP, SIM, STNK) dan *field* dinamisnya (termasuk *drag-and-drop*).
- **Laporan Agregat**: (Fitur Pro) Membuat laporan PDF yang berisi ringkasan statistik operator dan komposisi barang hilang berdasarkan rentang tanggal.
- **Sistem Lisensi**: Aktivasi fitur Pro menggunakan Serial Key yang terkunci pada Hardware ID (HMAC).
- **Manajemen Dokumen (CRUD)**: Sistem penuh untuk Membuat, Membaca, Memperbarui, dan Menghapus surat keterangan.
- **Generasi Dokumen Presisi**: Ekspor ke **PDF** (presisi tinggi) dan **Excel** (Fitur Pro).
- **Manajemen Pengguna (RBAC)**: Dua tingkat hak akses (Super Admin & Operator) dengan fitur aktivasi/non-aktivasi.
- **Modul Audit Log**: Merekam semua aktivitas penting (siapa, apa, kapan) dan dapat diekspor ke Excel.
- **Backup & Restore**: (Super Admin) Fungsionalitas *backup* dan *restore* yang mudah untuk database SQLite.
- **Penyimpanan Data Aman**: Konfigurasi (`.env`) dan database (`simdokpol.db`) secara otomatis disimpan di folder data pengguna (misal: `AppData\Roaming` atau `.config`).

---

## 📸 Galeri Fitur

<table width="100%">
    <tr>
        <td width="50%" align="center">
            <strong>Setup Multi-Database</strong><br>
            Pilih dialek database (SQLite, MySQL, Postgres) saat setup.
            <img src=".github/assets/guide-setup.png" width="90%">
        </td>
        <td width="50%" align="center">
            <strong>Manajemen Dokumen Aktif</strong><br>
            Pantau dan kelola surat keterangan yang masih berlaku.
            <img src=".github/assets/guide-doc-active.png" width="90%">
        </td>
    </tr>
    <tr>
        <td width="50%" align="center">
            <strong>Pengaturan Sistem (Tabs)</strong><br>
            Kelola KOP surat, koneksi DB, dan backup di satu tempat yang rapi.
            <img src=".github/assets/guide-settings-general.png" width="90%">
        </td>
        <td width="50%" align="center">
            <strong>Kustomisasi Template (Pro)</strong><br>
            (Admin) Atur item barang hilang dan field dinamisnya dengan mudah.
            <img src=".github/assets/guide-template-list.png" width="90%">
        </td>
    </tr>
    <tr>
        <td width="50%" align="center">
            <strong>Laporan Agregat (Pro)</strong><br>
            (Admin) Buat laporan PDF ringkasan berdasarkan rentang tanggal.
            <img src=".github/assets/guide-report.png" width="90%">
        </td>
        <td width="50%" align="center">
            <strong>Manajemen Pengguna</strong><br>
            (Admin) Kelola pengguna aktif dan non-aktif dengan mudah.
            <img src=".github/assets/guide-user-list.png" width="90%">
        </td>
    </tr>
</table>

---

## 🛠️ Tumpukan Teknologi (Technology Stack)

- **Backend**: Go (Golang)
- **Web Framework**: Gin
- **ORM & Migrasi**: GORM & golang-migrate
- **Database**: SQLite, MySQL, PostgreSQL
- **Desktop UI**: Go HTML Templates, CSS, JavaScript (jQuery, Bootstrap, Chart.js)
- **Generasi Dokumen**:
  - PDF: `gofpdf`
  - Excel: `excelize`
- **Integrasi Desktop**:
  - System Tray: `getlantern/systray`
  - Notifikasi: `gen2brain/beeep`
- **Build & Packaging**:
  - Installer Windows: NSIS
  - Paket Linux: DEB & RPM
  - Paket macOS: DMG

---

## 🚀 Memulai (Getting Started)

### Untuk Pengguna Akhir

Unduh installer terbaru dari halaman **[Releases](https://github.com/muhammad1505/simdokpol-release/releases)**.

**Di Windows**: Jalankan file `SIMDOKPOL-windows-x64-vX.X.X-installer.exe`. Ikuti wizard instalasi, kemudian jalankan aplikasi dari shortcut di Desktop atau Start Menu.

**Di Linux**: Instal file `.deb` atau `.rpm` menggunakan package manager Anda. Cari "SIMDOKPOL" di application menu Anda dan jalankan.

**Di macOS**: Buka file `.dmg` dan seret `SIMDOKPOL.app` ke folder Applications Anda. Jalankan dari Launchpad atau folder Applications.

#### Setup Pertama Kali

Saat pertama kali dijalankan, browser Anda akan terbuka ke halaman `http://localhost:8080/setup`. Di sini Anda bisa memilih:
1. **Konfigurasi Baru**: Mengikuti wizard 5-langkah untuk setup database (SQLite/MySQL/Postgres), KOP surat, dan membuat akun Admin.
2. **Pulihkan dari Backup**: (Hanya SQLite) Mengunggah file `simdokpol.db` dari instalasi sebelumnya untuk setup instan.

#### Setup Domain (Opsional)

Jika Anda ingin mengakses aplikasi via `http://simdokpol.local:8080`, klik kanan ikon aplikasi di system tray dan pilih **"Setup Domain (simdokpol.local)"**. Aksi ini akan meminta hak Administrator/Sudo.

### Untuk Pengembang

#### Prasyarat

- Go (versi 1.25+)
- Git
- C Compiler (TDM-GCC/Mingw-w64 untuk Windows, `build-essential` untuk Debian/Ubuntu)
- Library C untuk Systray (Lihat `README` `getlantern/systray`)

#### Instalasi & Menjalankan

Kloning repositori dengan perintah:

```bash
git clone https://github.com/muhammad1505/simdokpol.git
cd simdokpol
```

Instalasi dependensi:

```bash
go mod tidy
```

Menjalankan di mode pengembangan (SQLite) dengan live reload menggunakan Air (`air.toml` sudah tersedia):

```bash
air
```

Aplikasi akan berjalan di `http://localhost:8080`.

Untuk menjalankan di mode produksi lokal (menggunakan `.env`):

```bash
# Build binary
go build -o simdokpol ./cmd/main.go

# Jalankan binary
./simdokpol
```

---

## 📂 Struktur Proyek

```
simdokpol/
├── cmd/
│   ├── keygen/             # Generator Key Pair ECDSA (Dev Tool)
│   ├── license-manager/    # GUI Manager Lisensi & Key (Dev Tool)
│   ├── seeder/             # Seeder data dummy & migrator (Dev Tool)
│   ├── signer/             # Generator Serial Key CLI (Dev Tool)
│   └── main.go             # Entrypoint aplikasi utama
├── internal/
│   ├── config/             # Logic pemuatan config & .env
│   ├── controllers/        # HTTP handlers (logic API)
│   ├── dto/                # Data Transfer Objects
│   ├── middleware/         # Auth, Admin, Setup, License middleware
│   ├── models/             # Model data GORM (entities)
│   ├── mocks/              # Mock interface untuk unit test
│   ├── repositories/       # Data access layer (interaksi GORM)
│   ├── services/           # Business logic (Lisensi, Update, PDF, dll)
│   └── utils/              # Helper (PDF Generator, AppDir, HWID)
├── migrations/             # File migrasi skema database SQL
│   ├── mysql/              # Migrasi khusus MySQL
│   ├── sqlite/             # Migrasi khusus SQLite
│   └── postgres/           # Migrasi khusus PostgreSQL
├── web/
│   ├── static/             # Aset statis (CSS, JS, Img)
│   ├── templates/          # Template HTML & Partials
│   └── assets.go           # Go Embed Configuration
├── .github/
│   ├── assets/             # Screenshot untuk README
│   └── workflows/          # CI/CD (Test & Release Build)
├── .air.toml               # Konfigurasi live-reload Air
├── go.mod                  # Manajemen dependensi Go
└── README.md               # Dokumentasi proyek
```

---

## 🔧 Konfigurasi

Aplikasi ini menggunakan file `.env` yang disimpan di folder data pengguna (bukan folder instalasi) untuk portabilitas dan keamanan.

- **Windows**: `C:\Users\<NAMA>\AppData\Roaming\SIMDOKPOL\.env`
- **Linux**: `/home/<nama>/.config/SIMDOKPOL/.env`
- **macOS**: `/Users/<nama>/Library/Application Support/SIMDOKPOL/.env`

File `.env` akan dibuat otomatis saat setup, atau Anda dapat membuatnya manual dengan format berikut:

```ini
# --- PENGATURAN UMUM ---
JWT_SECRET_KEY=secret-anda-yang-sangat-aman
BCRYPT_COST=10
PORT=8080

# --- LISENSI (Diisi Otomatis) ---
LICENSE_KEY=XXXXX-XXXXX-XXXXX-XXXXX-XXXXX

# --- PENGATURAN DATABASE ---
# Pilih salah satu dialek: "sqlite", "mysql", atau "postgres"

# --- OPSI 1: SQLite (DEFAULT) ---
DB_DIALECT=sqlite
DB_DSN=simdokpol.db?_foreign_keys=on

# --- OPSI 2: MySQL (Contoh) ---
#DB_DIALECT=mysql
#DB_HOST=127.0.0.1
#DB_PORT=3306
#DB_USER=root
#DB_PASS=password
#DB_NAME=simdokpol

# --- OPSI 3: PostgreSQL (Contoh) ---
#DB_DIALECT=postgres
#DB_HOST=127.0.0.1
#DB_PORT=5432
#DB_USER=postgres
#DB_PASS=password
#DB_NAME=simdokpol
```

---

## 🛣️ Rencana Pengembangan (Roadmap)

### Selesai (Versi Terkini)

- ✅ Arsitektur Backend & Frontend yang Bersih
- ✅ Arsitektur Multi-Database (SQLite, MySQL, Postgres)
- ✅ Wizard Setup 5-Langkah dengan Tes Koneksi DB
- ✅ Kustomisasi Template Barang (Formulir Dinamis) - Fitur Pro
- ✅ Pembuatan Laporan Agregat PDF - Fitur Pro
- ✅ Sistem Lisensi & Aktivasi (Freemium)
- ✅ Update Checker Otomatis (via GitHub)
- ✅ Alur Kerja Surat Keterangan Hilang (CRUD Lengkap)
- ✅ Otentikasi & Otorisasi Berbasis Peran
- ✅ Manajemen Pengguna (CRUD, Aktivasi/Deaktivasi)
- ✅ Dasbor Analitik dengan Grafik Dinamis
- ✅ Modul Audit Log & Fitur Backup/Restore
- ✅ Aplikasi Desktop Standalone (via Systray & Beeep)
- ✅ Setup VHost Opsional (Tanpa Sudo/Admin saat startup)
- ✅ Installer Multi-Platform (NSIS, DEB, RPM, DMG)
- ✅ CI/CD Penuh (Unit Test, Linter, & Integration Test 3 DB)
- ✅ Ekspor Data ke Excel (Dokumen & Audit) - Fitur Pro
- ✅ Generasi PDF Presisi Tinggi (Server-Side)

### Rencana Selanjutnya

- [ ] Server Lisensi Online: Validasi lisensi terpusat untuk keamanan lebih tinggi.
- [ ] Template Editor (Drag & Drop): Memungkinkan Admin mengubah layout cetak PDF secara visual.
- [ ] Migrasi Data Antar DB: Alat untuk memindahkan data dari SQLite ke MySQL/Postgres secara otomatis.
- [ ] Notifikasi Email: Mengirim notifikasi surat kadaluwarsa via email.

---

## 🐛 Troubleshooting

### Virtual Host Issues

**Problem**: Domain `simdokpol.local` tidak bisa diakses

**Solution**: Klik kanan ikon aplikasi di system tray Anda, pilih "Setup Domain (simdokpol.local)", setujui permintaan Administrator/Sudo, kemudian restart aplikasi Anda.

### "Attempt to write a readonly database" (Windows)

**Problem**: Aplikasi di Windows gagal menyimpan data setelah instalasi.

**Solution**: Ini adalah bug di versi lama (v1.0.3 ke bawah). Bug ini sudah diperbaiki di v1.0.4+ dengan memindahkan database ke folder data pengguna Anda di `C:\Users\<NAMA>\AppData\Roaming\SIMDOKPOL\` untuk Windows atau `/home/<nama>/.config/SIMDOKPOL/` untuk Linux.

### Icon Not Showing in System Tray

**Solution**: Pastikan file `web/static/img/icon.ico` (Windows) atau `icon.png` (Linux/macOS) ada di folder instalasi Anda atau di sebelah binary jika menjalankan dari build lokal. Untuk Linux, pastikan Anda memiliki `libayatana-appindicator3-1` terinstal.

Untuk bug lainnya, silakan buat issue baru di halaman [GitHub Issues](https://github.com/muhammad1505/simdokpol/issues).

---

## 📄 Lisensi

Proyek ini dilisensikan di bawah [MIT License](LICENSE) - lihat file LICENSE untuk detail.

---

## 👥 Tim Pengembang

- **Lead Developer**: Muhammad Yusuf Abdurrohman
- **Contributors**: [View all contributors](https://github.com/muhammad1505/simdokpol/graphs/contributors)

---

## 📞 Dukungan & Kontak

Kami berkomitmen untuk memberikan dukungan terbaik bagi pengguna SIMDOKPOL. Silakan hubungi kami melalui saluran berikut:

### Dukungan Teknis

Untuk pertanyaan teknis, laporan bug, atau permintaan fitur, silakan gunakan platform GitHub kami yang memungkinkan kolaborasi transparan dan pelacakan issue yang terstruktur.

- **🐛 Pelaporan Bug**: Laporkan masalah teknis atau bug yang Anda temukan melalui [GitHub Issues](https://github.com/muhammad1505/simdokpol/issues). Pastikan untuk menyertakan detail lengkap tentang masalah, langkah-langkah reproduksi, dan informasi sistem Anda.

- **💬 Diskusi & Pertanyaan**: Untuk diskusi umum, pertanyaan implementasi, atau berbagi pengalaman dengan pengguna lain, kunjungi [GitHub Discussions](https://github.com/muhammad1505/simdokpol/discussions).

### Kontak Langsung

Untuk pertanyaan khusus, konsultasi implementasi, atau kebutuhan dukungan enterprise, Anda dapat menghubungi kami secara langsung.

- **📧 Email**: emailbaruku50@gmail.com  
  *Waktu respons: 1-2 hari kerja*

- **💬 WhatsApp Business**: +62 823-0001-4685  
  *Tersedia: Senin - Jumat, 09:00 - 17:00 WIB*

### Informasi Penting

Sebelum menghubungi kami, pastikan Anda telah:

- Memeriksa dokumentasi di README dan halaman Wiki (jika tersedia)
- Mencari solusi di GitHub Issues yang sudah ada
- Menyiapkan informasi sistem (versi aplikasi, sistem operasi, konfigurasi database)

Untuk pertanyaan mengenai lisensi Pro atau kerja sama institusional, silakan hubungi kami melalui email dengan subject line "SIMDOKPOL: Lisensi Enterprise".

---

<div align="center">

**Dikembangkan dengan ❤️ untuk Unit Kepolisian Indonesia**

[Website](https://github.com/muhammad1505/simdokpol) • [Documentation](https://github.com/muhammad1505/simdokpol/wiki) • [Releases](https://github.com/muhammad1505/simdokpol-release/releases) • [Changelog](https://github.com/muhammad1505/simdokpol/blob/main/CHANGELOG.md)

</div>