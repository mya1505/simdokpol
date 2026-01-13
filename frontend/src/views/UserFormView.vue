<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../lib/api'

const route = useRoute()
const router = useRouter()
const userId = computed(() => route.params.id)
const isEdit = computed(() => Boolean(userId.value))

const jabatanOptions = ref([])
const form = ref({
  nama_lengkap: '',
  nrp: '',
  pangkat: '',
  peran: 'OPERATOR',
  jabatan: '',
  regu: '',
  kata_sandi: '',
  konfirmasi: '',
})

const customJabatan = ref('')
const errorMessage = ref('')
const loading = ref(false)

const fetchJabatan = async () => {
  try {
    const { data } = await api.get('/jabatans/active')
    jabatanOptions.value = Array.isArray(data) ? data : []
  } catch {
    jabatanOptions.value = []
  }
}

const fetchUser = async () => {
  if (!isEdit.value) return
  try {
    const { data } = await api.get(`/users/${userId.value}`)
    form.value.nama_lengkap = data?.nama_lengkap || ''
    form.value.nrp = data?.nrp || ''
    form.value.pangkat = data?.pangkat || ''
    form.value.peran = data?.peran || 'OPERATOR'
    form.value.jabatan = data?.jabatan || ''
    form.value.regu = data?.regu || ''

    const match = jabatanOptions.value.find((j) => j.nama === form.value.jabatan)
    if (!match && form.value.jabatan) {
      customJabatan.value = form.value.jabatan
      form.value.jabatan = '__OTHER__'
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal memuat data pengguna.'
  }
}

const submit = async () => {
  errorMessage.value = ''
  if (!isEdit.value && form.value.kata_sandi.length < 8) {
    errorMessage.value = 'Kata sandi minimal 8 karakter.'
    return
  }
  if (form.value.kata_sandi && form.value.kata_sandi !== form.value.konfirmasi) {
    errorMessage.value = 'Konfirmasi kata sandi tidak cocok.'
    return
  }

  loading.value = true
  try {
    let jabatanValue = form.value.jabatan
    if (jabatanValue === '__OTHER__') {
      jabatanValue = customJabatan.value.trim()
    }
    const payload = {
      nama_lengkap: form.value.nama_lengkap,
      nrp: form.value.nrp,
      pangkat: form.value.pangkat,
      peran: form.value.peran,
      jabatan: jabatanValue,
      regu: form.value.regu,
      kata_sandi: form.value.kata_sandi,
    }
    if (isEdit.value) {
      await api.put(`/users/${userId.value}`, payload)
    } else {
      await api.post('/users', payload)
    }
    router.push({ name: 'users' })
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal menyimpan pengguna.'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await fetchJabatan()
  await fetchUser()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">{{ isEdit ? 'Edit Pengguna' : 'Tambah Pengguna' }}</h1>
      <p class="text-sm text-slate-500">Atur detail akun petugas dan perannya.</p>
    </div>

    <div v-if="errorMessage" class="rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">
      {{ errorMessage }}
    </div>

    <form class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm" @submit.prevent="submit">
      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <label class="text-sm font-medium text-slate-700">Nama Lengkap</label>
          <input v-model="form.nama_lengkap" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">NRP</label>
          <input v-model="form.nrp" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Pangkat</label>
          <input v-model="form.pangkat" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Peran</label>
          <select v-model="form.peran" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required>
            <option value="OPERATOR">Operator</option>
            <option value="SUPER_ADMIN">Super Admin</option>
          </select>
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Jabatan</label>
          <select v-model="form.jabatan" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required>
            <option value="">Pilih Jabatan</option>
            <option v-for="jab in jabatanOptions" :key="jab.id" :value="jab.nama">{{ jab.nama }}</option>
            <option value="__OTHER__">LAINNYA</option>
          </select>
        </div>
        <div v-if="form.jabatan === '__OTHER__'">
          <label class="text-sm font-medium text-slate-700">Jabatan Lainnya</label>
          <input v-model="customJabatan" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Regu</label>
          <input v-model="form.regu" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="I / II / -" />
        </div>
      </div>

      <div class="mt-6 grid gap-4 md:grid-cols-2">
        <div>
          <label class="text-sm font-medium text-slate-700">Kata Sandi</label>
          <input v-model="form.kata_sandi" type="password" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" :required="!isEdit" />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Konfirmasi Kata Sandi</label>
          <input v-model="form.konfirmasi" type="password" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" :required="!isEdit" />
        </div>
      </div>

      <div class="mt-6 flex justify-end gap-3">
        <RouterLink class="rounded-xl border border-slate-200 px-4 py-2 text-sm" to="/users">Kembali</RouterLink>
        <button type="submit" class="rounded-xl bg-slate-900 px-4 py-2 text-sm text-white" :disabled="loading">
          {{ loading ? 'Menyimpan...' : 'Simpan' }}
        </button>
      </div>
    </form>
  </div>
</template>
