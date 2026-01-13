<script setup>
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

const userName = computed(() => auth.user?.nama_lengkap || 'Pengguna')
const userRole = computed(() => auth.user?.peran || '-')
const isAdmin = computed(() => userRole.value === 'SUPER_ADMIN')

const handleLogout = async () => {
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen bg-slate-100">
    <div class="flex">
      <aside class="hidden min-h-screen w-64 flex-col gap-6 bg-slate-900 px-6 py-6 text-white lg:flex">
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-white/10">S</div>
          <div>
            <p class="text-lg font-semibold">SIMDOKPOL</p>
            <p class="text-xs text-white/60">SPKT System</p>
          </div>
        </div>

        <nav class="flex-1 space-y-2 text-sm">
          <RouterLink class="nav-link" to="/">Dashboard</RouterLink>
          <RouterLink class="nav-link" to="/documents">Dokumen Aktif</RouterLink>
          <RouterLink class="nav-link" to="/documents/archived">Arsip Dokumen</RouterLink>
          <RouterLink class="nav-link" to="/documents/new">Buat Surat Baru</RouterLink>
          <RouterLink v-if="isAdmin" class="nav-link" to="/users">Manajemen Pengguna</RouterLink>
          <RouterLink v-if="isAdmin" class="nav-link" to="/jabatan">Master Jabatan</RouterLink>
          <RouterLink v-if="isAdmin" class="nav-link" to="/templates">Template Barang</RouterLink>
          <RouterLink v-if="isAdmin" class="nav-link" to="/reports">Laporan Agregat</RouterLink>
          <RouterLink v-if="isAdmin" class="nav-link" to="/audit-logs">Log Audit</RouterLink>
          <RouterLink v-if="isAdmin" class="nav-link" to="/settings">Pengaturan Sistem</RouterLink>
          <RouterLink class="nav-link" to="/panduan">Panduan</RouterLink>
          <RouterLink class="nav-link" to="/tentang">Tentang</RouterLink>
          <RouterLink class="nav-link" to="/upgrade">Upgrade</RouterLink>
        </nav>

        <div class="rounded-2xl bg-white/10 p-4 text-xs">
          <p class="font-semibold">{{ userName }}</p>
          <p class="text-white/70">{{ userRole }}</p>
        </div>
      </aside>

      <main class="flex min-h-screen flex-1 flex-col">
        <header class="flex items-center justify-between bg-white px-6 py-4 shadow-sm">
          <div>
            <p class="text-sm text-slate-500">SIMDOKPOL</p>
            <h2 class="text-lg font-semibold text-slate-800">Panel Operasional</h2>
          </div>
          <div class="flex items-center gap-3">
            <RouterLink class="rounded-xl border border-slate-200 px-3 py-2 text-sm" to="/profile">Profil</RouterLink>
            <button class="rounded-xl bg-slate-900 px-3 py-2 text-sm text-white" @click="handleLogout">Keluar</button>
          </div>
        </header>

        <div class="flex-1 px-6 py-6">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.nav-link {
  display: block;
  border-radius: 12px;
  padding: 8px 12px;
  color: rgba(255, 255, 255, 0.75);
  transition: all 0.2s ease;
}
.nav-link.router-link-active {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}
.nav-link:hover {
  color: #fff;
}
</style>
