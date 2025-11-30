# Sistem Informasi Manajemen Dokumen Kepolisian (SIMDOKPOL)

![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey.svg)
![Database](https://img.shields.io/badge/Database-SQLite%20%7C%20MySQL%20%7C%20PostgreSQL-blue.svg)

SIMDOKPOL adalah aplikasi desktop cross-platform yang dirancang untuk membantu unit kepolisian dalam manajemen dan penerbitan surat keterangan secara efisien, cepat, dan aman. Aplikasi ini dapat berjalan dalam dua mode yaitu standalone yang sepenuhnya offline menggunakan database SQLite portabel untuk penggunaan di satu komputer, serta mode client-server yang terhubung ke database terpusat MySQL atau PostgreSQL untuk penggunaan bersama di jaringan kantor.

![Dasbor SIMDOKPOL](.github/assets/guide-dashboard.png)

---

## Fitur Utama

Aplikasi ini menyediakan kemampuan manajemen dokumen yang komprehensif dengan dukungan multi-database mencakup SQLite, MySQL, dan PostgreSQL. Pengguna mendapatkan kemudahan melalui wizard setup lima langkah yang terpandu, termasuk tes koneksi database dan pembuatan akun administrator. Aplikasi desktop standalone berjalan dengan integrasi system tray dan notifikasi native untuk pengalaman pengguna yang mulus.

Fitur tier profesional mencakup editor template dengan antarmuka drag-and-drop yang memungkinkan administrator menyesuaikan formulir barang hilang seperti KTP, SIM, dan BPKB sesuai kebutuhan organisasi. Alat migrasi data streaming memungkinkan pemindahan dataset lengkap secara aman dan real-time dari SQLite ke MySQL atau PostgreSQL. Kemampuan pelaporan agregat menghasilkan laporan PDF ringkasan yang berisi statistik operator dan analisis komposisi barang hilang berdasarkan rentang tanggal yang ditentukan.

Sistem lisensi mengaktifkan fitur profesional melalui serial key yang diamankan dengan identifikasi hardware menggunakan autentikasi HMAC. Fungsionalitas CRUD lengkap mendukung pembuatan, pembacaan, pembaruan, dan penghapusan dokumen surat keterangan. Generasi dokumen presisi tinggi mengekspor ke format PDF dengan kemampuan ekspor Excel pada tier profesional.

Kontrol akses berbasis peran menyediakan dua tingkat otorisasi untuk peran Super Admin dan Operator dengan kemampuan aktivasi dan deaktivasi. Modul audit log mencatat semua aktivitas penting termasuk identifikasi pengguna, tindakan yang dilakukan, dan timestamp dengan fungsionalitas ekspor Excel. Pengguna Super Admin mengakses fungsionalitas hot-backup dan restore untuk database SQLite.

Mode HTTPS secure mendukung enkripsi SSL lokal dengan instalasi sertifikat otomatis ke Windows Trusted Root. File konfigurasi dan database secara otomatis disimpan di folder data pengguna seperti AppData Roaming untuk keamanan dan portabilitas yang lebih baik.

---

## Galeri Fitur

Tangkapan layar berikut mendemonstrasikan kemampuan utama aplikasi di berbagai area fungsional.

**Konfigurasi Setup Multi-Database**

Antarmuka setup memungkinkan pengguna memilih dialek database pilihan mereka selama konfigurasi awal, mendukung SQLite untuk deployment standalone, MySQL untuk arsitektur client-server tradisional, dan PostgreSQL untuk kebutuhan tingkat enterprise.

![Antarmuka Setup](.github/assets/guide-setup.png)

**Antarmuka Manajemen Dokumen Aktif**

Layar manajemen dokumen menyediakan pengawasan komprehensif terhadap surat keterangan aktif dengan kemampuan filtering, pencarian, dan operasi bulk untuk manajemen siklus hidup dokumen yang efisien.

![Dokumen Aktif](.github/assets/guide-doc-active.png)

**Antarmuka Pengaturan Sistem dengan Tab**

Administrator mengakses konfigurasi kop surat, manajemen koneksi database, operasi backup, dan alat migrasi data melalui antarmuka pengaturan dengan tab terpadu yang memusatkan fungsi administrasi sistem.

![Panel Pengaturan](.github/assets/guide-settings-general.png)

**Editor Kustomisasi Template (Tier Profesional)**

Editor template visual memungkinkan administrator mengkonfigurasi kategori barang hilang dan field dinamis terkait melalui antarmuka intuitif yang tidak memerlukan keahlian teknis khusus.

![Editor Template](.github/assets/guide-template-list.png)

**Modul Pelaporan Agregat (Tier Profesional)**

Administrator menghasilkan laporan PDF komprehensif yang berisi analisis statistik aktivitas operator dan distribusi barang hilang di rentang tanggal yang ditentukan pengguna untuk keperluan review manajemen dan kepatuhan.

![Generasi Laporan](.github/assets/guide-report.png)

**Dashboard Manajemen Pengguna**

Antarmuka administrasi pengguna menyediakan kontrol lengkap atas akun pengguna aktif dan non-aktif dengan penugasan peran, manajemen status aktivasi, dan visibilitas jejak audit.

![Manajemen Pengguna](.github/assets/guide-user-list.png)

---

## Tumpukan Teknologi

Arsitektur aplikasi memanfaatkan teknologi modern untuk memberikan kinerja dan maintainability yang robust. Implementasi backend menggunakan Go versi 1.23 atau lebih tinggi dengan framework web Gin yang menyediakan routing HTTP dan dukungan middleware yang efisien. GORM berfungsi sebagai lapisan object-relational mapping dengan golang-migrate menangani versioning skema database di platform SQLite, MySQL, dan PostgreSQL.

Antarmuka pengguna desktop menggabungkan template HTML Go dengan styling CSS terinspirasi Metro dan library JavaScript termasuk jQuery untuk manipulasi DOM, Bootstrap untuk layout responsif, dan Chart.js untuk visualisasi data. Generasi dokumen menggunakan gofpdf untuk output PDF presisi tinggi dan excelize untuk pembuatan spreadsheet Excel.

Integrasi desktop mengandalkan getlantern systray untuk fungsionalitas system tray dan gen2brain beeep untuk pengiriman notifikasi native. Pipeline build dan packaging menggunakan NSIS untuk installer Windows, format DEB dan RPM untuk distribusi Linux, dan paket DMG untuk deployment macOS.

---

## Memulai

### Instalasi Pengguna Akhir

Unduh installer terbaru dari halaman Releases resmi di GitHub. Pengguna Windows harus menjalankan file installer dengan hak akses administrator dan mengikuti wizard instalasi, setelah itu aplikasi tersedia melalui shortcut Desktop atau Start Menu. Pengguna Linux menginstal paket DEB atau RPM yang sesuai menggunakan package manager distribusi mereka sebelum meluncurkan aplikasi dari menu aplikasi lingkungan desktop. Pengguna macOS membuka file DMG dan menyeret bundle aplikasi ke folder Applications untuk akses melalui Launchpad atau Finder.

### Proses Setup Awal

Saat peluncuran pertama, aplikasi secara otomatis membuka jendela browser ke halaman setup di localhost port 8080. Pengguna memilih antara membuat konfigurasi baru melalui wizard lima langkah yang mencakup pemilihan database, konfigurasi kop surat, dan pembuatan akun administrator, atau memulihkan dari file backup SQLite sebelumnya dengan mengunggah database yang sudah ada.

### Konfigurasi Domain dan HTTPS

Pengguna yang mengaktifkan mode HTTPS di pengaturan sistem menerima prompt yang meminta izin untuk menginstal sertifikat SSL ke Windows Trusted Root store. Instalasi ini mencegah peringatan keamanan browser dan memerlukan elevasi administrator untuk penetapan rantai kepercayaan sertifikat yang tepat.

### Setup Pengembang

Pengembang memerlukan Go versi 1.23 atau lebih tinggi, kontrol versi Git, dan compiler C seperti TDM-GCC atau Mingw-w64 untuk Windows atau paket build-essential untuk distribusi Linux berbasis Debian. Ketergantungan compiler C berasal dari persyaratan CGO SQLite untuk operasi database native.

Mulailah dengan mengkloning repository dan navigasi ke direktori proyek. Instal semua dependensi menggunakan sistem modul Go dengan perintah tidy. Untuk alur kerja pengembangan, jalankan aplikasi menggunakan Air untuk reload otomatis saat file sumber berubah. File konfigurasi air.toml yang disediakan berisi pengaturan yang dioptimalkan untuk iterasi pengembangan.

Build produksi memerlukan kompilasi dengan flag linker spesifik untuk menghapus simbol debugging, mengurangi ukuran binary, dan menyembunyikan jendela console pada platform Windows. Jalankan binary yang dihasilkan secara langsung untuk deployment produksi dengan loading konfigurasi berbasis environment.

---

## Struktur Proyek

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
│   ├── services/           # Business logic (Lisensi, Update, PDF, Backup, dll)
│   └── utils/              # Helper (PDF Generator, AppDir, HWID, Certs)
├── migrations/             # File migrasi skema database SQL (Arsip)
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

Organisasi codebase memisahkan concern di direktori logis untuk maintainability dan skalabilitas. Direktori cmd berisi entry point termasuk binary aplikasi utama, utilitas generasi key pair, GUI manajemen lisensi, alat seeding data, dan antarmuka command-line penandatanganan serial key.

Paket internal mengimplementasikan fungsionalitas inti dengan config menangani loading environment, controllers mengelola handler permintaan HTTP, data transfer objects mendefinisikan kontrak API, middleware menyediakan filter autentikasi dan otorisasi, models merepresentasikan entitas database, mock interfaces mendukung unit testing, repositories mengabstraksi pola akses data, services berisi logika bisnis untuk lisensi, update, generasi PDF, dan operasi backup, serta utilities menawarkan fungsi helper untuk generasi dokumen, direktori aplikasi, identifikasi hardware, dan manajemen sertifikat.

Migrasi database memelihara riwayat versi skema di dialek SQLite, MySQL, dan PostgreSQL dalam direktori terpisah. Direktori web mengorganisir aset statis termasuk stylesheet CSS, library JavaScript, dan gambar bersama template HTML dan partials, dengan konfigurasi Go embed memungkinkan bundling aset ke dalam binary yang dikompilasi.

Direktori khusus GitHub berisi screenshot marketing dan definisi workflow continuous integration untuk testing otomatis dan release builds. File konfigurasi mencakup pengaturan Air untuk hot-reload pengembangan dan spesifikasi dependensi modul Go.

---

## Manajemen Konfigurasi

Aplikasi menyimpan konfigurasi dalam file environment yang terletak di direktori data pengguna daripada folder instalasi untuk portabilitas dan keamanan yang lebih baik. Instalasi Windows menggunakan path direktori AppData Roaming, sistem Linux menggunakan standar direktori config XDG, dan macOS mengikuti konvensi Library Application Support. File environment dibuat secara otomatis selama setup awal dengan permission dan ownership yang sesuai.

---

## Roadmap Pengembangan

### Fitur yang Telah Diselesaikan

Rilis saat ini mencakup arsitektur backend dan frontend yang bersih dengan pemisahan concern yang komprehensif. Dukungan multi-database menyediakan implementasi production-ready untuk SQLite, MySQL, dan PostgreSQL dengan penanganan migrasi otomatis. Wizard setup lima langkah memandu pengguna melalui testing koneksi database dan konfigurasi awal.

Kemampuan tier profesional mencakup editor template drag-and-drop untuk kustomisasi formulir barang hilang, generasi laporan PDF agregat dengan analisis statistik, dan migrasi data streaming antar backend database. Sistem lisensi mengimplementasikan fungsionalitas freemium dengan hardware binding yang diamankan HMAC.

Pengecekan update otomatis memantau rilis GitHub untuk notifikasi versi. Manajemen workflow sertifikat lengkap menyediakan operasi CRUD penuh dengan validasi dan jejak audit. Autentikasi dan otorisasi berbasis peran menegakkan kontrol akses di tier pengguna.

Subsistem manajemen pengguna menangani operasi siklus hidup akun termasuk pembuatan, modifikasi, aktivasi, dan deaktivasi. Analitik dashboard dinamis memvisualisasikan metrik operasional melalui chart interaktif. Logging audit komprehensif menangkap semua tindakan signifikan dengan kemampuan ekspor Excel.

Fungsionalitas hot backup dan restore SQLite memungkinkan perlindungan data tanpa downtime. Aplikasi desktop mengintegrasikan kontrol system tray native dan pengiriman notifikasi. Dukungan HTTPS mencakup generasi sertifikat otomatis dan instalasi ke system trust store.

Packaging multi-platform menghadirkan installer native untuk NSIS Windows, DEB dan RPM Linux, dan format DMG macOS. Pipeline continuous integration menjalankan unit test, analisis statis, dan integration test di ketiga backend database yang didukung. Ekspor Excel tier profesional melampaui PDF untuk pertukaran data yang fleksibel. Generasi PDF server-side memastikan output dokumen berkualitas tinggi yang konsisten di seluruh platform klien.

### Peningkatan yang Direncanakan

Prioritas pengembangan masa depan mencakup eksposur API publik melalui spesifikasi Swagger atau OpenAPI yang memungkinkan peluang integrasi pihak ketiga. Implementasi server lisensi online akan menyediakan infrastruktur validasi terpusat untuk keamanan yang ditingkatkan dan analitik penggunaan.

Editor PDF visual akan memungkinkan administrator memodifikasi layout dan styling dokumen melalui antarmuka grafis tanpa manipulasi kode template. Integrasi notifikasi email akan mengirimkan alert otomatis untuk sertifikat yang akan kedaluwarsa dan event sistem penting ke penerima yang ditentukan.

---

## Panduan Troubleshooting

### Peringatan Keamanan HTTPS

Browser mungkin menampilkan peringatan sertifikat saat mengakses aplikasi melalui koneksi HTTPS. Navigasi ke panel System Settings dan akses tab General. Toggle mode HTTPS off lalu on lagi untuk memicu dialog instalasi sertifikat. Setujui prompt elevasi administrator Windows saat diminta untuk menginstal sertifikat ke system trust store, yang akan menghilangkan peringatan browser di masa mendatang.

### Error Permission Write Database pada Windows

Aplikasi yang dijalankan langsung dari arsip terkompresi kekurangan akses write ke direktori yang diperlukan. Ekstrak installer atau arsip portabel ke lokasi permanen sebelum eksekusi. Aplikasi memerlukan permission write ke direktori AppData pengguna untuk penyimpanan database dan konfigurasi.

### Masalah Visibilitas Icon System Tray

Build binary harus menyertakan aset statis embedded untuk rendering icon yang tepat. Verifikasi bahwa konten direktori web static img disertakan dalam binary yang dikompilasi melalui direktif Go embed. Sistem Linux memerlukan paket libayatana-appindicator3-1 untuk fungsionalitas system tray di sebagian besar lingkungan desktop.

Untuk dukungan teknis tambahan atau laporan bug, silakan buat laporan issue terperinci di halaman GitHub Issues dengan informasi sistem, pesan error, dan langkah reproduksi untuk memfasilitasi penyelesaian yang tepat waktu.

---

## Lisensi

Proyek ini didistribusikan di bawah ketentuan MIT License. Rujuk ke file LICENSE di root repository untuk teks legal lengkap dan izin.

---

## Tim Pengembangan

**Lead Developer:** Muhammad Yusuf Abdurrohman

Proyek ini menyambut kontribusi dari komunitas. Lihat daftar kontributor lengkap di halaman contributors repository GitHub.

---

## Dukungan dan Kontak

Tim pengembangan memelihara beberapa saluran dukungan untuk membantu pengguna SIMDOKPOL dengan masalah teknis dan pertanyaan umum.

**Saluran Dukungan Teknis**

Laporan bug harus dikirimkan melalui GitHub Issues tracker dengan langkah reproduksi terperinci dan informasi sistem. Diskusi umum, permintaan fitur, dan pertanyaan implementasi disambut di forum GitHub Discussions.

**Kontak Langsung**

Untuk pertanyaan, konsultasi, atau dukungan teknis, Anda dapat menghubungi kami melalui berbagai saluran komunikasi berikut:

📧 **Email:** emailbaruku50@gmail.com - Untuk pertanyaan bisnis, kemitraan, atau konsultasi mendalam mengenai implementasi sistem.

💬 **WhatsApp Business:** +62 823-0001-4685 - Dukungan real-time dan konsultasi cepat selama jam kerja Indonesia (Senin-Jumat, 08:00-17:00 WIB).

📱 **Telegram:** @simdokpol_support - Alternatif komunikasi instant untuk diskusi teknis dan update informasi produk terbaru.

👥 **Facebook Page:** [SIMDOKPOL Official](https://facebook.com/simdokpol) - Ikuti halaman resmi kami untuk update fitur, tips penggunaan, dan pengumuman penting.

Kami berkomitmen untuk merespons setiap pertanyaan dalam waktu maksimal 1x24 jam pada hari kerja. Untuk masalah kritis yang memerlukan penanganan segera, silakan hubungi melalui WhatsApp atau Telegram dengan mencantumkan label "URGENT" di awal pesan.

---

<div align="center">

**Dikembangkan dengan dedikasi untuk Unit Kepolisian Indonesia**

[Website](https://github.com/muhammad1505/simdokpol) • [Dokumentasi](https://github.com/muhammad1505/simdokpol/wiki) • [Releases](https://github.com/muhammad1505/simdokpol/releases) • [Changelog](https://github.com/muhammad1505/simdokpol/blob/main/CHANGELOG.md)

</div>