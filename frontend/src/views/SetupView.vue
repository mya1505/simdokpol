<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../lib/api'

const router = useRouter()
const loading = ref(false)
const message = ref('')
const errorMessage = ref('')
const restoreFile = ref(null)

const form = ref({
  db_dialect: 'sqlite',
  db_host: '',
  db_port: '',
  db_name: '',
  db_user: '',
  db_pass: '',
  db_sslmode: 'disable',

  kop_baris_1: '',
  kop_baris_2: '',
  kop_baris_3: '',
  nama_kantor: '',
  tempat_surat: '',

  kode_surat: 'SKH',
  kode_arsip: 'TUK.7.2.1',
  format_nomor_surat: '{KODE_SURAT}/{NOMOR}/{BULAN_ROMAWI}/{KODE_ARSIP}/{TAHUN}',
  nomor_surat_terakhir: '0',
  zona_waktu: 'Asia/Jakarta',
  archive_duration_days: '15',

  admin_nama_lengkap: '',
  admin_nrp: '',
  admin_pangkat: '',
  admin_jabatan: '',
  admin_password: '',
  admin_password_confirm: '',
})

const checkSetup = async () => {
  try {
    const { data } = await api.get('/config/limits')
    if (data?.is_setup_complete) {
      router.push({ name: 'login' })
    }
  } catch {
    // ignore
  }
}

const testDb = async () => {
  try {
    await api.post('/db/test', {
      db_dialect: form.value.db_dialect,
      db_host: form.value.db_host,
      db_port: form.value.db_port,
      db_name: form.value.db_name,
      db_user: form.value.db_user,
      db_pass: form.value.db_pass,
      db_sslmode: form.value.db_sslmode,
    })
    message.value = 'Koneksi database berhasil.'
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal koneksi database.'
  }
}

const submit = async () => {
  errorMessage.value = ''
  message.value = ''
  if (form.value.admin_password.length < 8) {
    errorMessage.value = 'Kata sandi admin minimal 8 karakter.'
    return
  }
  if (form.value.admin_password !== form.value.admin_password_confirm) {
    errorMessage.value = 'Konfirmasi kata sandi admin tidak cocok.'
    return
  }

  loading.value = true
  try {
    await api.post('/setup', {
      db_dialect: form.value.db_dialect,
      db_host: form.value.db_host,
      db_port: form.value.db_port,
      db_name: form.value.db_name,
      db_user: form.value.db_user,
      db_pass: form.value.db_pass,
      db_sslmode: form.value.db_sslmode,

      kop_baris_1: form.value.kop_baris_1,
      kop_baris_2: form.value.kop_baris_2,
      kop_baris_3: form.value.kop_baris_3,
      nama_kantor: form.value.nama_kantor,
      tempat_surat: form.value.tempat_surat,

      format_nomor_surat: form.value.format_nomor_surat,
      kode_surat: form.value.kode_surat,
      kode_arsip: form.value.kode_arsip,
      nomor_surat_terakhir: form.value.nomor_surat_terakhir,
      zona_waktu: form.value.zona_waktu,
      archive_duration_days: form.value.archive_duration_days,

      admin_nama_lengkap: form.value.admin_nama_lengkap,
      admin_nrp: form.value.admin_nrp,
      admin_pangkat: form.value.admin_pangkat,
      admin_jabatan: form.value.admin_jabatan,
      admin_password: form.value.admin_password,
    })
    message.value = 'Setup berhasil. Sistem akan dimuat ulang.'
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal menyimpan setup.'
  } finally {
    loading.value = false
  }
}

const restoreSetup = async () => {
  if (!restoreFile.value) {
    errorMessage.value = 'Pilih file backup terlebih dahulu.'
    return
  }
  const formData = new FormData()
  formData.append('restore-file', restoreFile.value)
  try {
    await api.post('/setup/restore', formData)
    message.value = 'Restore berhasil. Sistem akan restart.'
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal restore setup.'
  }
}

onMounted(checkSetup)
</script>

<template>
  <div class="min-h-screen bg-slate-950">
    <div class="mx-auto max-w-5xl px-6 py-10">
      <div class="rounded-3xl bg-white p-8 shadow-2xl">
        <h1 class="text-2xl font-semibold text-slate-900">Setup Awal SIMDOKPOL</h1>
        <p class="mt-1 text-sm text-slate-500">Lengkapi konfigurasi sistem sebelum digunakan.</p>

        <div v-if="message" class="mt-4 rounded-xl bg-emerald-50 px-4 py-2 text-sm text-emerald-700">
          {{ message }}
        </div>
        <div v-if="errorMessage" class="mt-4 rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">
          {{ errorMessage }}
        </div>

        <form class="mt-6 space-y-6" @submit.prevent="submit">
          <div class="rounded-2xl border border-slate-200 p-5">
            <h2 class="text-lg font-semibold text-slate-800">Database</h2>
            <div class="mt-3 grid gap-4 md:grid-cols-3">
              <select v-model="form.db_dialect" class="rounded-xl border border-slate-200 px-3 py-2 text-sm">
                <option value="sqlite">SQLite</option>
                <option value="mysql">MySQL</option>
                <option value="postgres">Postgres</option>
              </select>
              <input v-model="form.db_host" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Host" />
              <input v-model="form.db_port" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Port" />
              <input v-model="form.db_name" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Nama DB" />
              <input v-model="form.db_user" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="User" />
              <input v-model="form.db_pass" type="password" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Password" />
              <input v-model="form.db_sslmode" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="SSL Mode" />
            </div>
            <button type="button" class="mt-4 rounded-xl border border-slate-200 px-3 py-2 text-sm" @click="testDb">
              Tes Koneksi
            </button>
          </div>

          <div class="rounded-2xl border border-slate-200 p-5">
            <h2 class="text-lg font-semibold text-slate-800">Identitas Instansi</h2>
            <div class="mt-3 grid gap-4 md:grid-cols-3">
              <input v-model="form.kop_baris_1" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="KOP Baris 1" />
              <input v-model="form.kop_baris_2" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="KOP Baris 2" />
              <input v-model="form.kop_baris_3" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="KOP Baris 3" />
              <input v-model="form.nama_kantor" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Nama Kantor" />
              <input v-model="form.tempat_surat" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Tempat Terbit" />
            </div>
          </div>

          <div class="rounded-2xl border border-slate-200 p-5">
            <h2 class="text-lg font-semibold text-slate-800">Penomoran & Arsip</h2>
            <div class="mt-3 grid gap-4 md:grid-cols-3">
              <input v-model="form.kode_surat" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Kode Surat" />
              <input v-model="form.kode_arsip" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Kode Arsip" />
              <input v-model="form.nomor_surat_terakhir" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Nomor Terakhir" />
              <input v-model="form.format_nomor_surat" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Format Nomor" />
              <input v-model="form.zona_waktu" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Zona Waktu" />
              <input v-model="form.archive_duration_days" type="number" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Durasi Arsip (hari)" />
            </div>
          </div>

          <div class="rounded-2xl border border-slate-200 p-5">
            <h2 class="text-lg font-semibold text-slate-800">Akun Super Admin</h2>
            <div class="mt-3 grid gap-4 md:grid-cols-3">
              <input v-model="form.admin_nama_lengkap" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Nama Lengkap" />
              <input v-model="form.admin_nrp" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="NRP" />
              <input v-model="form.admin_pangkat" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Pangkat" />
              <input v-model="form.admin_jabatan" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Jabatan" />
              <input v-model="form.admin_password" type="password" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Password" />
              <input v-model="form.admin_password_confirm" type="password" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Konfirmasi Password" />
            </div>
          </div>

          <div class="rounded-2xl border border-slate-200 p-5">
            <h2 class="text-lg font-semibold text-slate-800">Restore Setup</h2>
            <div class="mt-3 flex flex-wrap items-center gap-3">
              <input type="file" @change="(e) => (restoreFile = e.target.files[0])" />
              <button type="button" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" @click="restoreSetup">Restore</button>
            </div>
          </div>

          <div class="flex justify-end">
            <button type="submit" class="rounded-xl bg-slate-900 px-4 py-2 text-sm text-white" :disabled="loading">
              {{ loading ? 'Menyimpan...' : 'Simpan Setup' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
