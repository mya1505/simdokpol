<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">Panduan Pengguna</h1>
      <p class="text-sm text-slate-500">Ringkasan cara menggunakan sistem.</p>
    </div>

    <div class="grid gap-6 lg:grid-cols-3">
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Setup Awal</h2>
        <ul class="mt-3 list-disc space-y-2 pl-5 text-sm text-slate-600">
          <li>Pilih jenis database dan isi kredensial jika menggunakan server.</li>
          <li>Lengkapi identitas instansi (KOP surat).</li>
          <li>Atur format nomor surat, kode surat, dan kode arsip.</li>
          <li>Buat akun Super Admin pertama.</li>
        </ul>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Operasional Dokumen</h2>
        <ul class="mt-3 list-disc space-y-2 pl-5 text-sm text-slate-600">
          <li>Buat surat baru lewat menu Dokumen.</li>
          <li>Tambahkan barang hilang sesuai template.</li>
          <li>Preview/unduh PDF untuk cetak presisi.</li>
          <li>Dokumen lama akan berpindah ke arsip otomatis.</li>
        </ul>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Administrasi</h2>
        <ul class="mt-3 list-disc space-y-2 pl-5 text-sm text-slate-600">
          <li>Kelola pengguna dan master jabatan.</li>
          <li>Atur template barang untuk kebutuhan lokal.</li>
          <li>Gunakan backup/restore atau migrasi database.</li>
          <li>Aktifkan HTTPS untuk keamanan jaringan.</li>
        </ul>
      </div>
    </div>
  </div>
</template>
