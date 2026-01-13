<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../lib/api'

const router = useRouter()
const templates = ref([])
const loading = ref(false)

const fetchTemplates = async () => {
  loading.value = true
  try {
    const { data } = await api.get('/item-templates')
    templates.value = Array.isArray(data) ? data : []
  } catch (error) {
    templates.value = []
  } finally {
    loading.value = false
  }
}

const removeTemplate = async (item) => {
  if (!confirm(`Hapus template ${item.nama_barang}?`)) return
  try {
    await api.delete(`/item-templates/${item.id}`)
    fetchTemplates()
  } catch (error) {
    alert(error?.response?.data?.error || 'Gagal menghapus template.')
  }
}

const editTemplate = (item) => {
  router.push({ name: 'templates-edit', params: { id: item.id } })
}

onMounted(fetchTemplates)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold text-slate-800">Template Barang</h1>
        <p class="text-sm text-slate-500">Kelola formulir dinamis untuk barang hilang.</p>
      </div>
      <RouterLink class="rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white" to="/templates/new">
        Tambah Template
      </RouterLink>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead class="bg-slate-50 text-left text-xs uppercase text-slate-500">
            <tr>
              <th class="px-3 py-2">Nama Barang</th>
              <th class="px-3 py-2">Urutan</th>
              <th class="px-3 py-2">Status</th>
              <th class="px-3 py-2">Jumlah Field</th>
              <th class="px-3 py-2">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="5" class="px-3 py-4 text-center text-slate-500">Memuat data...</td>
            </tr>
            <tr v-else-if="templates.length === 0">
              <td colspan="5" class="px-3 py-4 text-center text-slate-500">Belum ada template.</td>
            </tr>
            <tr v-for="item in templates" :key="item.id" class="border-t border-slate-100">
              <td class="px-3 py-2 font-semibold text-slate-700">{{ item.nama_barang }}</td>
              <td class="px-3 py-2">{{ item.urutan }}</td>
              <td class="px-3 py-2">
                <span class="rounded-full px-2 py-1 text-xs" :class="item.is_active ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-200 text-slate-600'">
                  {{ item.is_active ? 'Aktif' : 'Nonaktif' }}
                </span>
              </td>
              <td class="px-3 py-2">{{ item.fields_config?.length || 0 }}</td>
              <td class="px-3 py-2">
                <div class="flex flex-wrap gap-2">
                  <button class="rounded-lg border border-slate-200 px-2 py-1 text-xs" @click="editTemplate(item)">Edit</button>
                  <button class="rounded-lg bg-red-50 px-2 py-1 text-xs text-red-600" @click="removeTemplate(item)">Hapus</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
