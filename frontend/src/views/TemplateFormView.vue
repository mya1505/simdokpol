<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../lib/api'

const route = useRoute()
const router = useRouter()
const templateId = computed(() => route.params.id)
const isEdit = computed(() => Boolean(templateId.value))

const form = ref({
  nama_barang: '',
  urutan: 10,
  is_active: true,
})

const fields = ref([])
const errorMessage = ref('')
const loading = ref(false)

const addField = () => {
  fields.value.push({
    label: '',
    type: 'text',
    data_label: '',
    placeholder: '',
    options: [],
    options_input: '',
    is_numeric: false,
    is_uppercase: false,
    is_titlecase: false,
    required_length: false,
  })
}

const removeField = (index) => {
  fields.value.splice(index, 1)
}

const loadTemplate = async () => {
  if (!isEdit.value) return
  try {
    const { data } = await api.get(`/item-templates/${templateId.value}`)
    form.value.nama_barang = data?.nama_barang || ''
    form.value.urutan = data?.urutan || 0
    form.value.is_active = data?.is_active ?? true
    fields.value = Array.isArray(data?.fields_config) ? data.fields_config.map((field) => ({
      label: field.label || '',
      type: field.type || 'text',
      data_label: field.data_label || '',
      placeholder: field.placeholder || '',
      options: field.options || [],
      options_input: Array.isArray(field.options) ? field.options.join(', ') : '',
      is_numeric: Boolean(field.is_numeric),
      is_uppercase: Boolean(field.is_uppercase),
      is_titlecase: Boolean(field.is_titlecase),
      required_length: Boolean(field.required_length),
    })) : []
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal memuat template.'
  }
}

const submit = async () => {
  errorMessage.value = ''
  if (!form.value.nama_barang) {
    errorMessage.value = 'Nama barang wajib diisi.'
    return
  }
  if (fields.value.length === 0) {
    errorMessage.value = 'Minimal satu field harus ditambahkan.'
    return
  }
  loading.value = true
  try {
    const payload = {
      nama_barang: form.value.nama_barang,
      urutan: Number(form.value.urutan || 0),
      is_active: form.value.is_active,
      fields_config: fields.value.map((field) => ({
        label: field.label,
        type: field.type,
        data_label: field.data_label,
        placeholder: field.placeholder,
        options: field.type === 'select'
          ? (field.options_input || '')
              .split(',')
              .map((value) => value.trim())
              .filter(Boolean)
          : [],
        is_numeric: field.is_numeric,
        is_uppercase: field.is_uppercase,
        is_titlecase: field.is_titlecase,
        required_length: field.required_length ? 1 : 0,
      })),
    }
    if (isEdit.value) {
      await api.put(`/item-templates/${templateId.value}`, payload)
    } else {
      await api.post('/item-templates', payload)
    }
    router.push({ name: 'templates' })
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal menyimpan template.'
  } finally {
    loading.value = false
  }
}

onMounted(loadTemplate)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">{{ isEdit ? 'Edit Template Barang' : 'Tambah Template Barang' }}</h1>
      <p class="text-sm text-slate-500">Buat konfigurasi field untuk formulir barang hilang.</p>
    </div>

    <div v-if="errorMessage" class="rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">
      {{ errorMessage }}
    </div>

    <form class="space-y-6" @submit.prevent="submit">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="grid gap-4 md:grid-cols-3">
          <div>
            <label class="text-sm font-medium text-slate-700">Nama Barang</label>
            <input v-model="form.nama_barang" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="text-sm font-medium text-slate-700">Urutan</label>
            <input v-model="form.urutan" type="number" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" />
          </div>
          <div class="flex items-center gap-2 pt-6 text-sm">
            <input v-model="form.is_active" type="checkbox" class="h-4 w-4" />
            Aktif
          </div>
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-lg font-semibold text-slate-800">Konfigurasi Field</h2>
          <button type="button" class="rounded-xl bg-primary-600 px-3 py-2 text-sm text-white" @click="addField">Tambah Field</button>
        </div>

        <div v-if="fields.length === 0" class="mt-4 rounded-xl border border-dashed border-slate-200 p-6 text-center text-sm text-slate-500">
          Belum ada field. Klik "Tambah Field" untuk mulai.
        </div>

        <div v-for="(field, index) in fields" :key="index" class="mt-4 rounded-xl border border-slate-200 p-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-semibold text-slate-700">Field {{ index + 1 }}</h3>
            <button type="button" class="text-xs text-red-600" @click="removeField(index)">Hapus</button>
          </div>
          <div class="mt-3 grid gap-3 md:grid-cols-3">
            <input v-model="field.label" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Label" />
            <select v-model="field.type" class="rounded-xl border border-slate-200 px-3 py-2 text-sm">
              <option value="text">Teks</option>
              <option value="number">Angka</option>
              <option value="textarea">Area Teks</option>
              <option value="date">Tanggal</option>
              <option value="select">Dropdown</option>
            </select>
            <input v-model="field.data_label" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Data Label" />
            <input v-model="field.placeholder" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Placeholder" />
            <div class="flex flex-wrap items-center gap-3 text-xs text-slate-600">
              <label class="flex items-center gap-2"><input v-model="field.required_length" type="checkbox" class="h-4 w-4" />Wajib</label>
              <label class="flex items-center gap-2"><input v-model="field.is_numeric" type="checkbox" class="h-4 w-4" />Angka</label>
              <label class="flex items-center gap-2"><input v-model="field.is_uppercase" type="checkbox" class="h-4 w-4" />Uppercase</label>
              <label class="flex items-center gap-2"><input v-model="field.is_titlecase" type="checkbox" class="h-4 w-4" />Titlecase</label>
            </div>
          </div>
          <div v-if="field.type === 'select'" class="mt-3">
            <label class="text-xs text-slate-500">Opsi (pisahkan dengan koma)</label>
            <input v-model="field.options_input" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="contoh: SIM A, SIM C, SIM B" />
            <div class="mt-2 text-xs text-slate-500">Saat ini: {{ field.options_input || '-' }}</div>
          </div>
        </div>
      </div>

      <div class="flex justify-end gap-3">
        <RouterLink class="rounded-xl border border-slate-200 px-4 py-2 text-sm" to="/templates">Kembali</RouterLink>
        <button type="submit" class="rounded-xl bg-slate-900 px-4 py-2 text-sm text-white" :disabled="loading">
          {{ loading ? 'Menyimpan...' : 'Simpan Template' }}
        </button>
      </div>
    </form>
  </div>
</template>
