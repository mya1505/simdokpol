<script setup>
import { onMounted, ref } from 'vue'
import api from '../lib/api'

const hwid = ref('')
const licenseStatus = ref('')
const activationCode = ref('')
const message = ref('')
const errorMessage = ref('')

const fetchStatus = async () => {
  try {
    const { data } = await api.get('/config/limits')
    licenseStatus.value = data?.license_status || 'UNLICENSED'
  } catch {
    licenseStatus.value = 'UNLICENSED'
  }
}

const fetchHwid = async () => {
  try {
    const { data } = await api.get('/license/hwid')
    hwid.value = data?.hardware_id || ''
  } catch {
    hwid.value = ''
  }
}

const copyHwid = async () => {
  if (!hwid.value) return
  try {
    await navigator.clipboard.writeText(hwid.value)
    message.value = 'HWID disalin ke clipboard.'
  } catch {
    message.value = 'Gagal menyalin HWID.'
  }
}

const pasteCode = async () => {
  try {
    const text = await navigator.clipboard.readText()
    activationCode.value = text
  } catch {
    // ignore
  }
}

const activate = async () => {
  message.value = ''
  errorMessage.value = ''
  if (!activationCode.value.trim()) {
    errorMessage.value = 'Activation Code wajib diisi.'
    return
  }
  try {
    const { data } = await api.post('/license/activate', { key: activationCode.value.trim() })
    message.value = data?.message || 'Lisensi berhasil diaktifkan.'
    await fetchStatus()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal aktivasi lisensi.'
  }
}

onMounted(async () => {
  await fetchStatus()
  await fetchHwid()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">Upgrade & Lisensi</h1>
      <p class="text-sm text-slate-500">Aktifkan lisensi Professional untuk membuka fitur lengkap.</p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p class="text-sm text-slate-500">Status Lisensi</p>
          <p class="text-lg font-semibold text-slate-800">{{ licenseStatus }}</p>
        </div>
        <div class="rounded-full px-3 py-1 text-xs" :class="licenseStatus === 'PRO' ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-200 text-slate-600'">
          {{ licenseStatus === 'PRO' ? 'Aktif' : 'Belum Aktif' }}
        </div>
      </div>
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Hardware ID</h2>
        <p class="text-sm text-slate-500">Kirim HWID ke admin untuk mendapatkan Activation Code.</p>
        <div class="mt-4 flex flex-wrap items-center gap-3">
          <code class="rounded-xl bg-slate-100 px-3 py-2 text-sm">{{ hwid || '...' }}</code>
          <button class="rounded-xl border border-slate-200 px-3 py-2 text-sm" @click="copyHwid">Copy</button>
        </div>
        <div class="mt-4">
          <img v-if="hwid" :src="`/api/license/hwid/qr?size=160`" alt="QR HWID" class="rounded-xl border border-slate-200" />
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Aktivasi Lisensi</h2>
        <p class="text-sm text-slate-500">Masukkan Activation Code untuk mengaktifkan lisensi.</p>
        <div class="mt-4 flex gap-2">
          <input v-model="activationCode" type="text" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Activation Code" />
          <button class="rounded-xl border border-slate-200 px-3 py-2 text-sm" type="button" @click="pasteCode">Paste</button>
        </div>
        <div class="mt-4 flex justify-end">
          <button class="rounded-xl bg-primary-600 px-4 py-2 text-sm text-white" @click="activate">Aktifkan</button>
        </div>
        <div v-if="message" class="mt-3 rounded-xl bg-emerald-50 px-3 py-2 text-sm text-emerald-700">{{ message }}</div>
        <div v-if="errorMessage" class="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm text-red-600">{{ errorMessage }}</div>
      </div>
    </div>
  </div>
</template>
