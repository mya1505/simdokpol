<script setup>
import { onMounted, ref } from 'vue'
import api from '../lib/api'

const positions = ref([])
const form = ref({ nama: '', is_active: true })
const loading = ref(false)
const errorMessage = ref('')

const fetchPositions = async () => {
  loading.value = true
  try {
    const { data } = await api.get('/jabatans')
    positions.value = Array.isArray(data) ? data : []
  } catch (error) {
    positions.value = []
  } finally {
    loading.value = false
  }
}

const addPosition = async () => {
  errorMessage.value = ''
  if (!form.value.nama.trim()) {
    errorMessage.value = 'Nama jabatan wajib diisi.'
    return
  }
  try {
    await api.post('/jabatans', {
      nama: form.value.nama,
      is_active: form.value.is_active,
    })
    form.value.nama = ''
    form.value.is_active = true
    fetchPositions()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal menambah jabatan.'
  }
}

const toggleStatus = async (item) => {
  try {
    await api.put(`/jabatans/${item.id}`, {
      nama: item.nama,
      is_active: !item.is_active,
    })
    fetchPositions()
  } catch (error) {
    alert(error?.response?.data?.error || 'Gagal mengubah status.')
  }
}

const removePosition = async (item) => {
  if (!confirm(`Nonaktifkan jabatan ${item.nama}?`)) return
  try {
    await api.delete(`/jabatans/${item.id}`)
    fetchPositions()
  } catch (error) {
    alert(error?.response?.data?.error || 'Gagal menghapus jabatan.')
  }
}

onMounted(fetchPositions)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">Master Jabatan</h1>
      <p class="text-sm text-slate-500">Kelola daftar jabatan agar input konsisten.</p>
    </div>

    <div v-if="errorMessage" class="rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">
      {{ errorMessage }}
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-slate-800">Tambah Jabatan</h2>
      <div class="mt-4 flex flex-wrap items-center gap-3">
        <input v-model="form.nama" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Nama jabatan" />
        <label class="flex items-center gap-2 text-sm text-slate-600">
          <input v-model="form.is_active" type="checkbox" class="h-4 w-4" />
          Aktif
        </label>
        <button class="rounded-xl bg-primary-600 px-4 py-2 text-sm text-white" @click="addPosition">Tambah</button>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead class="bg-slate-50 text-left text-xs uppercase text-slate-500">
            <tr>
              <th class="px-3 py-2">Nama Jabatan</th>
              <th class="px-3 py-2">Status</th>
              <th class="px-3 py-2">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="3" class="px-3 py-4 text-center text-slate-500">Memuat data...</td>
            </tr>
            <tr v-else-if="positions.length === 0">
              <td colspan="3" class="px-3 py-4 text-center text-slate-500">Belum ada data.</td>
            </tr>
            <tr v-for="item in positions" :key="item.id" class="border-t border-slate-100">
              <td class="px-3 py-2 font-semibold text-slate-700">{{ item.nama }}</td>
              <td class="px-3 py-2">
                <span class="rounded-full px-2 py-1 text-xs" :class="item.is_active ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-200 text-slate-600'">
                  {{ item.is_active ? 'Aktif' : 'Nonaktif' }}
                </span>
              </td>
              <td class="px-3 py-2">
                <div class="flex flex-wrap gap-2">
                  <button class="rounded-lg border border-slate-200 px-2 py-1 text-xs" @click="toggleStatus(item)">
                    {{ item.is_active ? 'Nonaktifkan' : 'Aktifkan' }}
                  </button>
                  <button class="rounded-lg bg-red-50 px-2 py-1 text-xs text-red-600" @click="removePosition(item)">Hapus</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
