import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import AppShell from '../layouts/AppShell.vue'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import DocumentsListView from '../views/DocumentsListView.vue'
import DocumentFormView from '../views/DocumentFormView.vue'
import UsersListView from '../views/UsersListView.vue'
import UserFormView from '../views/UserFormView.vue'
import SettingsView from '../views/SettingsView.vue'
import ProfileView from '../views/ProfileView.vue'
import AuditLogsView from '../views/AuditLogsView.vue'
import JobPositionsView from '../views/JobPositionsView.vue'
import TemplatesListView from '../views/TemplatesListView.vue'
import TemplateFormView from '../views/TemplateFormView.vue'
import ReportsView from '../views/ReportsView.vue'
import UpgradeView from '../views/UpgradeView.vue'
import PanduanView from '../views/PanduanView.vue'
import TentangView from '../views/TentangView.vue'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: LoginView,
  },
  {
    path: '/',
    component: AppShell,
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'dashboard', component: DashboardView },
      { path: 'documents', name: 'documents', component: DocumentsListView, props: { status: 'active', title: 'Dokumen Aktif' } },
      { path: 'documents/archived', name: 'documents-archived', component: DocumentsListView, props: { status: 'archived', title: 'Arsip Dokumen' } },
      { path: 'documents/new', name: 'documents-new', component: DocumentFormView },
      { path: 'documents/:id/edit', name: 'documents-edit', component: DocumentFormView },
      { path: 'users', name: 'users', component: UsersListView },
      { path: 'users/new', name: 'users-new', component: UserFormView },
      { path: 'users/:id/edit', name: 'users-edit', component: UserFormView },
      { path: 'jabatan', name: 'jabatan', component: JobPositionsView },
      { path: 'templates', name: 'templates', component: TemplatesListView },
      { path: 'templates/new', name: 'templates-new', component: TemplateFormView },
      { path: 'templates/:id/edit', name: 'templates-edit', component: TemplateFormView },
      { path: 'reports', name: 'reports', component: ReportsView },
      { path: 'audit-logs', name: 'audit', component: AuditLogsView },
      { path: 'settings', name: 'settings', component: SettingsView },
      { path: 'panduan', name: 'panduan', component: PanduanView },
      { path: 'tentang', name: 'tentang', component: TentangView },
      { path: 'upgrade', name: 'upgrade', component: UpgradeView },
      { path: 'profile', name: 'profile', component: ProfileView },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory('/app/'),
  routes,
})

let sessionChecked = false

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!sessionChecked) {
    await auth.fetchSession()
    sessionChecked = true
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login' }
  }

  if (to.name === 'login' && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }

  return true
})

export default router
