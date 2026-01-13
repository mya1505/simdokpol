<script setup>
import { onMounted, ref } from 'vue'
import api from '../lib/api'

const config = ref({})
const loading = ref(false)
const message = ref('')
const errorMessage = ref('')
const restoreFile = ref(null)

const fetchSettings = async () => {
  try {
    const { data } = await api.get('/settings')
    config.value = data || {}
  } catch (error) {
    errorMessage.value = 'Gagal memuat pengaturan.'
  }
}

const saveSettings = async () => {
  loading.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const payload = {
      kop_baris_1: config.value.kop_baris_1 || '',
      kop_baris_2: config.value.kop_baris_2 || '',
      kop_baris_3: config.value.kop_baris_3 || '',
      nama_kantor: config.value.nama_kantor || '',
      tempat_surat: config.value.tempat_surat || '',
      kode_surat: config.value.kode_surat || '',
      kode_arsip: config.value.kode_arsip || '',
      format_nomor_surat: config.value.format_nomor_surat || '',
      nomor_surat_terakhir: config.value.nomor_surat_terakhir || '',
      zona_waktu: config.value.zona_waktu || '',
      archive_duration_days: String(config.value.archive_duration_days || ''),
      enable_https: config.value.enable_https ? 'true' : 'false',
      db_dialect: config.value.db_dialect || 'sqlite',
      db_host: config.value.db_host || '',
      db_port: config.value.db_port || '',
      db_user: config.value.db_user || '',
      db_name: config.value.db_name || '',
      db_dsn: config.value.db_dsn || '',
      db_sslmode: config.value.db_sslmode || '',
      db_pass: config.value.db_pass || '',
    }
    const { data } = await api.put('/settings', payload)
    message.value = data?.message || 'Pengaturan tersimpan.'
    if (data?.check_https_cert) {
      message.value += ' Silakan install sertifikat.'
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal menyimpan pengaturan.'
  } finally {
    loading.value = false
  }
}

const downloadCert = () => {
  window.open('/api/settings/download-cert', '_blank')
}

const installCert = async () => {
  try {
    const { data } = await api.post('/settings/install-cert')
    alert(data?.message || 'Sertifikat terpasang.')
  } catch (error) {
    alert(error?.response?.data?.error || 'Gagal install sertifikat.')
  }
}

const backupDb = async () => {
  try {
    const response = await api.post('/backups', null, { responseType: 'blob' })
    const url = window.URL.createObjectURL(response.data)
    const link = document.createElement('a')
    link.href = url
    link.download = 'simdokpol_backup.db'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (error) {
    alert(error?.response?.data?.error || 'Gagal backup.')
  }
}

const restoreDb = async () => {
  if (!restoreFile.value) {
    alert('Pilih file backup terlebih dahulu.')
    return
  }
  const formData = new FormData()
  formData.append('restore-file', restoreFile.value)
  try {
    await api.post('/restore', formData)
    alert('Restore berhasil. Sistem akan restart otomatis.')
  } catch (error) {
    alert(error?.response?.data?.error || 'Gagal restore.')
  }
}

onMounted(fetchSettings)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">Pengaturan Sistem</h1>
      <p class="text-sm text-slate-500">Kelola identitas instansi, database, dan keamanan.</p>
    </div>

    <div v-if="message" class="rounded-xl bg-emerald-50 px-4 py-2 text-sm text-emerald-700">
      {{ message }}
    </div>
    <div v-if="errorMessage" class="rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">
      {{ errorMessage }}
    </div>

    <form class="space-y-6" @submit.prevent="saveSettings">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Identitas Instansi</h2>
        <div class="mt-4 grid gap-4 md:grid-cols-3">
          <input v-model="config.kop_baris_1" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Baris 1" />
          <input v-model="config.kop_baris_2" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Baris 2" />
          <input v-model="config.kop_baris_3" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Baris 3" />
        </div>
        <div class="mt-4 grid gap-4 md:grid-cols-2">
          <input v-model="config.nama_kantor" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Nama Kantor" />
          <input v-model="config.tempat_surat" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Tempat Terbit" />
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Penomoran & Arsip</h2>
        <div class="mt-4 grid gap-4 md:grid-cols-3">
          <input v-model="config.kode_surat" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Kode Surat" />
          <input v-model="config.kode_arsip" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Kode Arsip" />
          <input v-model="config.nomor_surat_terakhir" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Nomor Terakhir" />
        </div>
        <div class="mt-4 grid gap-4 md:grid-cols-2">
          <input v-model="config.format_nomor_surat" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Format Nomor" />
          <input v-model="config.zona_waktu" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Zona Waktu" />
        </div>
        <div class="mt-4">
          <input v-model="config.archive_duration_days" type="number" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Durasi Arsip (hari)" />
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Database</h2>
        <div class="mt-4 grid gap-4 md:grid-cols-3">
          <select v-model="config.db_dialect" class="rounded-xl border border-slate-200 px-3 py-2 text-sm">
            <option value="sqlite">SQLite</option>
            <option value="mysql">MySQL</option>
            <option value="postgres">Postgres</option>
          </select>
          <input v-model="config.db_host" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Host" />
          <input v-model="config.db_port" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Port" />
          <input v-model="config.db_name" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="DB Name" />
          <input v-model="config.db_user" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="DB User" />
          <input v-model="config.db_pass" type="password" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="DB Pass" />
          <input v-model="config.db_sslmode" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="SSL Mode" />
          <input v-model="config.db_dsn" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="SQLite DSN" />
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Keamanan HTTPS</h2>
        <div class="mt-4 flex flex-wrap items-center gap-3">
          <label class="flex items-center gap-2 text-sm text-slate-600">
            <input v-model="config.enable_https" type="checkbox" class="h-4 w-4" />
            Aktifkan HTTPS
          </label>
          <button type="button" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" @click="downloadCert">Download Sertifikat</button>
          <button type="button" class="rounded-xl bg-emerald-600 px-3 py-2 text-sm text-white" @click="installCert">Install Sertifikat</button>
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Backup & Restore</h2>
        <div class="mt-4 flex flex-wrap items-center gap-3">
          <button type="button" class="rounded-xl bg-primary-600 px-3 py-2 text-sm text-white" @click="backupDb">Backup Database</button>
          <input type="file" class="text-sm" @change="(e) => (restoreFile = e.target.files[0])" />
          <button type="button" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" @click="restoreDb">Restore</button>
        </div>
      </div>

      <div class="flex justify-end">
        <button type="submit" class="rounded-xl bg-slate-900 px-4 py-2 text-sm text-white" :disabled="loading">
          {{ loading ? 'Menyimpan...' : 'Simpan Pengaturan' }}
        </button>
      </div>
    </form>
  </div>
</template>
