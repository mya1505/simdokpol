<script setup>
import { ref } from 'vue'
import api from '../lib/api'

const updateInfo = ref(null)
const loading = ref(false)
const errorMessage = ref('')

const checkUpdate = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/updates/check')
    updateInfo.value = data
  } catch (error) {
    errorMessage.value = 'Gagal mengecek pembaruan.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">Tentang Aplikasi</h1>
      <p class="text-sm text-slate-500">Informasi teknis dan pembaruan aplikasi.</p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-slate-800">SIMDOKPOL</h2>
      <p class="mt-2 text-sm text-slate-600">
        Sistem Informasi Manajemen Dokumen Kepolisian untuk SPKT. Fokus pada penerbitan surat keterangan hilang, arsip, dan pelaporan.
      </p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-lg font-semibold text-slate-800">Cek Pembaruan</h2>
          <p class="text-sm text-slate-500">Periksa versi terbaru dari server update.</p>
        </div>
        <button class="rounded-xl bg-primary-600 px-4 py-2 text-sm text-white" @click="checkUpdate" :disabled="loading">
          {{ loading ? 'Memeriksa...' : 'Cek Update' }}
        </button>
      </div>

      <div v-if="errorMessage" class="mt-4 rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">
        {{ errorMessage }}
      </div>

      <div v-if="updateInfo" class="mt-4 rounded-xl border border-slate-200 p-4 text-sm">
        <p><strong>Versi Saat Ini:</strong> {{ updateInfo.current_version || '-' }}</p>
        <p><strong>Versi Terbaru:</strong> {{ updateInfo.latest_version || '-' }}</p>
        <p><strong>Status:</strong> {{ updateInfo.has_update ? 'Update tersedia' : 'Sudah terbaru' }}</p>
        <a v-if="updateInfo.download_url" :href="updateInfo.download_url" target="_blank" class="mt-2 inline-block text-primary-600">
          Buka halaman rilis
        </a>
        <pre v-if="updateInfo.release_notes" class="mt-3 max-h-48 overflow-auto rounded-lg bg-slate-50 p-3 text-xs text-slate-600">{{ updateInfo.release_notes }}</pre>
      </div>
    </div>
  </div>
</template>
