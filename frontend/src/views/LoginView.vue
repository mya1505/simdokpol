<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const nrp = ref('')
const password = ref('')
const errorMessage = ref('')
const loading = ref(false)

const submit = async () => {
  errorMessage.value = ''
  loading.value = true
  try {
    await auth.login({ nrp: nrp.value, password: password.value })
    await router.push({ name: 'dashboard' })
  } catch (error) {
    const message = error?.response?.data?.error || 'Gagal login. Coba lagi.'
    errorMessage.value = message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-950">
    <div class="mx-auto flex min-h-screen max-w-6xl items-center justify-center px-6 py-12">
      <div class="w-full max-w-md rounded-3xl bg-white/95 p-8 shadow-2xl">
        <div class="mb-6 text-center">
          <div class="mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-600 text-white">
            <span class="text-2xl font-bold">S</span>
          </div>
          <h1 class="text-2xl font-semibold text-slate-900">SIMDOKPOL</h1>
          <p class="text-sm text-slate-500">Sistem Informasi Manajemen Dokumen Kepolisian</p>
        </div>

        <form class="space-y-4" @submit.prevent="submit">
          <div>
            <label class="text-sm font-medium text-slate-700">NRP</label>
            <input v-model="nrp" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-4 py-3 text-sm focus:border-primary-500 focus:outline-none" placeholder="Masukkan NRP" required />
          </div>
          <div>
            <label class="text-sm font-medium text-slate-700">Kata Sandi</label>
            <input v-model="password" type="password" class="mt-1 w-full rounded-xl border border-slate-200 px-4 py-3 text-sm focus:border-primary-500 focus:outline-none" placeholder="Masukkan kata sandi" required />
          </div>

          <p v-if="errorMessage" class="rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">
            {{ errorMessage }}
          </p>

          <button type="submit" class="flex w-full items-center justify-center rounded-xl bg-primary-600 px-4 py-3 text-sm font-semibold text-white transition hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-60" :disabled="loading">
            <span v-if="loading">Memproses...</span>
            <span v-else>Masuk</span>
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
