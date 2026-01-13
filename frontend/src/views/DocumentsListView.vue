<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../lib/api'

const props = defineProps({
  status: { type: String, default: 'active' },
  title: { type: String, default: 'Dokumen' },
})

const route = useRoute()
const router = useRouter()
const rows = ref([])
const loading = ref(false)
const search = ref('')

const fetchDocuments = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams({
      draw: '1',
      start: '0',
      length: '100',
      status: props.status,
    })
    if (search.value) {
      params.append('search[value]', search.value)
    }
    const { data } = await api.get(`/documents?${params.toString()}`)
    rows.value = Array.isArray(data?.data) ? data.data : []
  } catch (error) {
    rows.value = []
  } finally {
    loading.value = false
  }
}

const formatDate = (value) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleDateString('id-ID')
}

const handleDelete = async (row) => {
  if (!confirm(`Hapus dokumen ${row.nomor_surat}?`)) return
  try {
    await api.delete(`/documents/${row.id}`)
    fetchDocuments()
  } catch (error) {
    alert(error?.response?.data?.error || 'Gagal menghapus dokumen.')
  }
}

const handleEdit = (row) => {
  router.push({ name: 'documents-edit', params: { id: row.id } })
}

const handlePreview = (row) => {
  window.open(`/documents/${row.id}/print`, '_blank')
}

const handlePdf = (row) => {
  window.open(`/api/documents/${row.id}/pdf`, '_blank')
}

const titleText = computed(() => props.title)

onMounted(fetchDocuments)
watch(() => route.fullPath, fetchDocuments)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold text-slate-800">{{ titleText }}</h1>
        <p class="text-sm text-slate-500">Kelola surat keterangan sesuai status.</p>
      </div>
      <div class="flex gap-2">
        <RouterLink class="rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white" to="/documents/new">
          Buat Surat Baru
        </RouterLink>
        <button v-if="status === 'active'" class="rounded-xl border border-slate-200 px-4 py-2 text-sm" @click="fetchDocuments">
          Refresh
        </button>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div class="flex gap-2">
          <RouterLink class="rounded-lg px-3 py-2 text-sm" :class="status === 'active' ? 'bg-primary-600 text-white' : 'border border-slate-200'" to="/documents">
            Aktif
          </RouterLink>
          <RouterLink class="rounded-lg px-3 py-2 text-sm" :class="status === 'archived' ? 'bg-primary-600 text-white' : 'border border-slate-200'" to="/documents/archived">
            Arsip
          </RouterLink>
        </div>
        <div class="flex items-center gap-2">
          <input v-model="search" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Cari nomor atau nama..." />
          <button class="rounded-xl bg-slate-900 px-3 py-2 text-sm text-white" @click="fetchDocuments">Cari</button>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead class="bg-slate-50 text-left text-xs uppercase text-slate-500">
            <tr>
              <th class="px-3 py-2">Nomor</th>
              <th class="px-3 py-2">Pemohon</th>
              <th class="px-3 py-2">Tanggal</th>
              <th class="px-3 py-2">Status</th>
              <th class="px-3 py-2">Operator</th>
              <th class="px-3 py-2">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="6" class="px-3 py-4 text-center text-slate-500">Memuat data...</td>
            </tr>
            <tr v-else-if="rows.length === 0">
              <td colspan="6" class="px-3 py-4 text-center text-slate-500">Belum ada data.</td>
            </tr>
            <tr v-for="row in rows" :key="row.id" class="border-t border-slate-100">
              <td class="px-3 py-2 font-semibold text-slate-700">{{ row.nomor_surat }}</td>
              <td class="px-3 py-2">{{ row.resident?.nama_lengkap || '-' }}</td>
              <td class="px-3 py-2">{{ formatDate(row.tanggal_laporan) }}</td>
              <td class="px-3 py-2">
                <span class="rounded-full px-2 py-1 text-xs" :class="row.status === 'DIARSIPKAN' ? 'bg-slate-200 text-slate-600' : 'bg-emerald-100 text-emerald-700'">
                  {{ row.status || '-' }}
                </span>
              </td>
              <td class="px-3 py-2">{{ row.operator?.nama_lengkap || '-' }}</td>
              <td class="px-3 py-2">
                <div class="flex flex-wrap gap-2">
                  <button class="rounded-lg border border-slate-200 px-2 py-1 text-xs" @click="handlePreview(row)">Preview</button>
                  <button class="rounded-lg border border-slate-200 px-2 py-1 text-xs" @click="handlePdf(row)">PDF</button>
                  <button v-if="status === 'active'" class="rounded-lg border border-slate-200 px-2 py-1 text-xs" @click="handleEdit(row)">Edit</button>
                  <button v-if="status === 'active'" class="rounded-lg bg-red-50 px-2 py-1 text-xs text-red-600" @click="handleDelete(row)">Hapus</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
