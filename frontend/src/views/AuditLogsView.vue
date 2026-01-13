<script setup>
import { onMounted, ref } from 'vue'
import api from '../lib/api'

const logs = ref([])
const loading = ref(false)
const search = ref('')

const fetchLogs = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams({
      draw: '1',
      start: '0',
      length: '100',
    })
    if (search.value) {
      params.append('search[value]', search.value)
    }
    const { data } = await api.get(`/audit-logs?${params.toString()}`)
    logs.value = Array.isArray(data?.data) ? data.data : []
  } catch (error) {
    logs.value = []
  } finally {
    loading.value = false
  }
}

const exportLogs = () => {
  window.open('/api/audit-logs/export', '_blank')
}

const formatDate = (value) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('id-ID')
}

onMounted(fetchLogs)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold text-slate-800">Log Audit</h1>
        <p class="text-sm text-slate-500">Pantau aktivitas penting di sistem.</p>
      </div>
      <button class="rounded-xl bg-emerald-600 px-4 py-2 text-sm font-semibold text-white" @click="exportLogs">
        Ekspor Excel
      </button>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div class="mb-4 flex items-center gap-2">
        <input v-model="search" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Cari log..." />
        <button class="rounded-xl bg-slate-900 px-3 py-2 text-sm text-white" @click="fetchLogs">Cari</button>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead class="bg-slate-50 text-left text-xs uppercase text-slate-500">
            <tr>
              <th class="px-3 py-2">Waktu</th>
              <th class="px-3 py-2">Pengguna</th>
              <th class="px-3 py-2">Aksi</th>
              <th class="px-3 py-2">Detail</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="4" class="px-3 py-4 text-center text-slate-500">Memuat data...</td>
            </tr>
            <tr v-else-if="logs.length === 0">
              <td colspan="4" class="px-3 py-4 text-center text-slate-500">Belum ada log.</td>
            </tr>
            <tr v-for="log in logs" :key="log.id" class="border-t border-slate-100">
              <td class="px-3 py-2">{{ formatDate(log.timestamp) }}</td>
              <td class="px-3 py-2">{{ log.user?.nama_lengkap || '-' }}</td>
              <td class="px-3 py-2 font-semibold text-slate-700">{{ log.aksi }}</td>
              <td class="px-3 py-2 text-slate-600">{{ log.detail || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
