<script setup>
import { ref } from 'vue'

const startDate = ref('')
const endDate = ref('')
const errorMessage = ref('')

const generateReport = () => {
  errorMessage.value = ''
  if (!startDate.value || !endDate.value) {
    errorMessage.value = 'Tanggal mulai dan selesai wajib diisi.'
    return
  }
  const url = `/api/reports/aggregate/pdf?start_date=${startDate.value}&end_date=${endDate.value}`
  window.open(url, '_blank')
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">Laporan Agregat</h1>
      <p class="text-sm text-slate-500">Buat laporan PDF berdasarkan rentang tanggal.</p>
    </div>

    <div v-if="errorMessage" class="rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">
      {{ errorMessage }}
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="grid gap-4 md:grid-cols-3">
        <div>
          <label class="text-sm font-medium text-slate-700">Tanggal Mulai</label>
          <input v-model="startDate" type="date" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Tanggal Selesai</label>
          <input v-model="endDate" type="date" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" />
        </div>
        <div class="flex items-end">
          <button class="w-full rounded-xl bg-red-600 px-4 py-2 text-sm font-semibold text-white" @click="generateReport">
            Generate PDF
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
