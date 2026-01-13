<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../lib/api'

const router = useRouter()
const users = ref([])
const loading = ref(false)
const status = ref('active')
const search = ref('')

const fetchUsers = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams({
      draw: '1',
      start: '0',
      length: '100',
      status: status.value,
    })
    if (search.value) {
      params.append('search[value]', search.value)
    }
    const { data } = await api.get(`/users?${params.toString()}`)
    users.value = Array.isArray(data?.data) ? data.data : []
  } catch (error) {
    users.value = []
  } finally {
    loading.value = false
  }
}

const toggleStatus = async (user) => {
  if (user.is_active === false || user.deleted_at) {
    await api.post(`/users/${user.id}/activate`)
  } else {
    await api.delete(`/users/${user.id}`)
  }
  fetchUsers()
}

const goEdit = (user) => {
  router.push({ name: 'users-edit', params: { id: user.id } })
}

onMounted(fetchUsers)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold text-slate-800">Manajemen Pengguna</h1>
        <p class="text-sm text-slate-500">Kelola akun petugas dan aksesnya.</p>
      </div>
      <RouterLink class="rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white" to="/users/new">
        Tambah Pengguna
      </RouterLink>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div class="flex gap-2">
          <button class="rounded-lg px-3 py-2 text-sm" :class="status === 'active' ? 'bg-primary-600 text-white' : 'border border-slate-200'" @click="status = 'active'; fetchUsers()">Aktif</button>
          <button class="rounded-lg px-3 py-2 text-sm" :class="status === 'inactive' ? 'bg-primary-600 text-white' : 'border border-slate-200'" @click="status = 'inactive'; fetchUsers()">Nonaktif</button>
        </div>
        <div class="flex items-center gap-2">
          <input v-model="search" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Cari nama atau NRP" />
          <button class="rounded-xl bg-slate-900 px-3 py-2 text-sm text-white" @click="fetchUsers">Cari</button>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead class="bg-slate-50 text-left text-xs uppercase text-slate-500">
            <tr>
              <th class="px-3 py-2">Nama</th>
              <th class="px-3 py-2">NRP</th>
              <th class="px-3 py-2">Pangkat</th>
              <th class="px-3 py-2">Jabatan</th>
              <th class="px-3 py-2">Regu</th>
              <th class="px-3 py-2">Peran</th>
              <th class="px-3 py-2">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="7" class="px-3 py-4 text-center text-slate-500">Memuat data...</td>
            </tr>
            <tr v-else-if="users.length === 0">
              <td colspan="7" class="px-3 py-4 text-center text-slate-500">Belum ada data.</td>
            </tr>
            <tr v-for="user in users" :key="user.id" class="border-t border-slate-100">
              <td class="px-3 py-2 font-semibold text-slate-700">{{ user.nama_lengkap }}</td>
              <td class="px-3 py-2">{{ user.nrp }}</td>
              <td class="px-3 py-2">{{ user.pangkat }}</td>
              <td class="px-3 py-2">{{ user.jabatan }}</td>
              <td class="px-3 py-2">{{ user.regu || '-' }}</td>
              <td class="px-3 py-2">{{ user.peran }}</td>
              <td class="px-3 py-2">
                <div class="flex flex-wrap gap-2">
                  <button class="rounded-lg border border-slate-200 px-2 py-1 text-xs" @click="goEdit(user)">Edit</button>
                  <button class="rounded-lg px-2 py-1 text-xs" :class="status === 'active' ? 'bg-red-50 text-red-600' : 'bg-emerald-50 text-emerald-700'" @click="toggleStatus(user)">
                    {{ status === 'active' ? 'Nonaktifkan' : 'Aktifkan' }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
