import { defineStore } from 'pinia'
import api from '../lib/api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    loading: false,
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.user),
  },
  actions: {
    async fetchSession() {
      try {
        this.loading = true
        const { data } = await api.get('/auth/me')
        this.user = data?.data?.user || null
      } catch (error) {
        this.user = null
      } finally {
        this.loading = false
      }
    },
    async login(payload) {
      await api.post('/login', payload)
      await this.fetchSession()
    },
    async logout() {
      await api.post('/logout')
      this.user = null
    },
  },
})
