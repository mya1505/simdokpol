<script setup>
import { onMounted, ref } from 'vue'
import api from '../lib/api'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const profile = ref({ nama_lengkap: '', nrp: '', pangkat: '' })
const passwordForm = ref({ old_password: '', new_password: '', confirm_password: '' })
const message = ref('')
const errorMessage = ref('')

const loadProfile = async () => {
  await auth.fetchSession()
  profile.value.nama_lengkap = auth.user?.nama_lengkap || ''
  profile.value.nrp = auth.user?.nrp || ''
  profile.value.pangkat = auth.user?.pangkat || ''
}

const saveProfile = async () => {
  message.value = ''
  errorMessage.value = ''
  try {
    const { data } = await api.put('/profile', profile.value)
    message.value = data?.message || 'Profil diperbarui.'
    await auth.fetchSession()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal memperbarui profil.'
  }
}

const changePassword = async () => {
  message.value = ''
  errorMessage.value = ''
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    errorMessage.value = 'Konfirmasi kata sandi tidak cocok.'
    return
  }
  try {
    const { data } = await api.put('/profile/password', passwordForm.value)
    message.value = data?.message || 'Kata sandi diperbarui.'
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal mengganti kata sandi.'
  }
}

onMounted(loadProfile)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">Profil Pengguna</h1>
      <p class="text-sm text-slate-500">Perbarui data diri dan kata sandi.</p>
    </div>

    <div v-if="message" class="rounded-xl bg-emerald-50 px-4 py-2 text-sm text-emerald-700">{{ message }}</div>
    <div v-if="errorMessage" class="rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">{{ errorMessage }}</div>

    <div class="grid gap-6 lg:grid-cols-2">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Data Profil</h2>
        <div class="mt-4 space-y-3">
          <input v-model="profile.nama_lengkap" type="text" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Nama Lengkap" />
          <input v-model="profile.nrp" type="text" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="NRP" />
          <input v-model="profile.pangkat" type="text" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Pangkat" />
        </div>
        <button class="mt-4 rounded-xl bg-slate-900 px-4 py-2 text-sm text-white" @click="saveProfile">Simpan Profil</button>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Ganti Kata Sandi</h2>
        <div class="mt-4 space-y-3">
          <input v-model="passwordForm.old_password" type="password" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Kata sandi saat ini" />
          <input v-model="passwordForm.new_password" type="password" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Kata sandi baru" />
          <input v-model="passwordForm.confirm_password" type="password" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Konfirmasi kata sandi" />
        </div>
        <button class="mt-4 rounded-xl bg-primary-600 px-4 py-2 text-sm text-white" @click="changePassword">Update Password</button>
      </div>
    </div>
  </div>
</template>
