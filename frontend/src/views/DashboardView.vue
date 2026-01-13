<script setup>
import { onMounted, ref } from 'vue'
import api from '../lib/api'

const stats = ref({
  docs_today: 0,
  docs_monthly: 0,
  docs_yearly: 0,
  active_users: 0,
})
const loading = ref(false)

const fetchStats = async () => {
  try {
    loading.value = true
    const { data } = await api.get('/stats')
    stats.value = data
  } catch (error) {
    // fallback stays
  } finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<template>
  <div class="space-y-6">
    <div class="rounded-3xl bg-gradient-to-r from-primary-600 via-primary-500 to-indigo-400 p-6 text-white shadow-xl">
      <h1 class="text-2xl font-semibold">Dashboard SIMDOKPOL</h1>
      <p class="mt-1 text-sm text-white/80">Ringkasan aktivitas surat dan pengguna aktif.</p>
    </div>

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <p class="text-xs uppercase tracking-wide text-slate-500">Surat Hari Ini</p>
        <p class="mt-2 text-3xl font-semibold text-slate-800">{{ loading ? '...' : stats.docs_today }}</p>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <p class="text-xs uppercase tracking-wide text-slate-500">Surat Bulan Ini</p>
        <p class="mt-2 text-3xl font-semibold text-slate-800">{{ loading ? '...' : stats.docs_monthly }}</p>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <p class="text-xs uppercase tracking-wide text-slate-500">Surat Tahun Ini</p>
        <p class="mt-2 text-3xl font-semibold text-slate-800">{{ loading ? '...' : stats.docs_yearly }}</p>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <p class="text-xs uppercase tracking-wide text-slate-500">Pengguna Aktif</p>
        <p class="mt-2 text-3xl font-semibold text-slate-800">{{ loading ? '...' : stats.active_users }}</p>
      </div>
    </div>

    <div class="rounded-2xl border border-dashed border-slate-200 bg-white/60 p-6 text-sm text-slate-500">
      Grafik bulanan dan komposisi barang akan muncul setelah migrasi modul chart selesai.
    </div>
  </div>
</template>
