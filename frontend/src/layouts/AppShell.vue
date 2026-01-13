<script setup>
import { computed, ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()
const isNavOpen = ref(false)

const userName = computed(() => auth.user?.nama_lengkap || 'Pengguna')
const userRole = computed(() => auth.user?.peran || '-')
const isAdmin = computed(() => userRole.value === 'SUPER_ADMIN')

const navItems = computed(() => [
  { label: 'Dashboard', to: '/', admin: false },
  { label: 'Dokumen Aktif', to: '/documents', admin: false },
  { label: 'Arsip Dokumen', to: '/documents/archived', admin: false },
  { label: 'Buat Surat Baru', to: '/documents/new', admin: false },
  { label: 'Manajemen Pengguna', to: '/users', admin: true },
  { label: 'Master Jabatan', to: '/jabatan', admin: true },
  { label: 'Template Barang', to: '/templates', admin: true },
  { label: 'Laporan Agregat', to: '/reports', admin: true },
  { label: 'Log Audit', to: '/audit-logs', admin: true },
  { label: 'Pengaturan Sistem', to: '/settings', admin: true },
  { label: 'Panduan', to: '/panduan', admin: false },
  { label: 'Tentang', to: '/tentang', admin: false },
  { label: 'Upgrade', to: '/upgrade', admin: false },
])

const visibleNavItems = computed(() =>
  navItems.value.filter((item) => !item.admin || isAdmin.value),
)

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
          <RouterLink
            v-for="item in visibleNavItems"
            :key="item.to"
            class="nav-link"
            :to="item.to"
          >
            {{ item.label }}
          </RouterLink>
        </nav>

        <div class="rounded-2xl bg-white/10 p-4 text-xs">
          <p class="font-semibold">{{ userName }}</p>
          <p class="text-white/70">{{ userRole }}</p>
        </div>
      </aside>

      <div class="lg:hidden">
        <div
          v-if="isNavOpen"
          class="fixed inset-0 z-40 bg-black/40"
          @click="isNavOpen = false"
        ></div>
        <aside
          class="fixed left-0 top-0 z-50 flex h-full w-64 flex-col gap-6 bg-slate-900 px-6 py-6 text-white transition-transform duration-200"
          :class="isNavOpen ? 'translate-x-0' : '-translate-x-full'"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-white/10">S</div>
              <div>
                <p class="text-lg font-semibold">SIMDOKPOL</p>
                <p class="text-xs text-white/60">SPKT System</p>
              </div>
            </div>
            <button class="text-white/70" @click="isNavOpen = false">Tutup</button>
          </div>

          <nav class="flex-1 space-y-2 text-sm">
            <RouterLink
              v-for="item in visibleNavItems"
              :key="item.to"
              class="nav-link"
              :to="item.to"
              @click="isNavOpen = false"
            >
              {{ item.label }}
            </RouterLink>
          </nav>

          <div class="rounded-2xl bg-white/10 p-4 text-xs">
            <p class="font-semibold">{{ userName }}</p>
            <p class="text-white/70">{{ userRole }}</p>
          </div>
        </aside>
      </div>

      <main class="flex min-h-screen flex-1 flex-col">
        <header class="flex items-center justify-between bg-white px-4 py-3 shadow-sm lg:hidden">
          <button class="rounded-xl border border-slate-200 px-3 py-2 text-sm" @click="isNavOpen = true">
            Menu
          </button>
          <span class="text-sm font-semibold text-slate-700">SIMDOKPOL</span>
        </header>

        <header class="hidden items-center justify-between bg-white px-6 py-4 shadow-sm lg:flex">
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
