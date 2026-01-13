<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../lib/api'

const route = useRoute()
const router = useRouter()
const docId = computed(() => route.params.id)
const isEdit = computed(() => Boolean(docId.value))

const form = ref({
  nama_lengkap: '',
  tempat_lahir: '',
  tanggal_lahir: '',
  jenis_kelamin: '',
  agama: '',
  pekerjaan: '',
  alamat: '',
  lokasi_hilang: '',
  petugas_pelapor_id: '',
  pejabat_persetuju_id: '',
})

const items = ref([])
const itemTemplates = ref([])
const selectedTemplate = ref('')
const lainnya = ref({ nama_barang: '', deskripsi: '' })
const dynamicValues = ref({})
const operators = ref([])
const loading = ref(false)
const errorMessage = ref('')

const fetchOperators = async () => {
  try {
    const { data } = await api.get('/users/operators')
    operators.value = Array.isArray(data) ? data : []
  } catch {
    operators.value = []
  }
}

const fetchTemplates = async () => {
  try {
    const { data } = await api.get('/item-templates/active')
    itemTemplates.value = Array.isArray(data) ? data : []
  } catch {
    itemTemplates.value = []
  }
}

const selectedTemplateConfig = computed(() => {
  if (selectedTemplate.value === 'LAINNYA') return null
  return itemTemplates.value.find((t) => t.nama_barang === selectedTemplate.value) || null
})

const fetchDocument = async () => {
  if (!isEdit.value) return
  try {
    const { data } = await api.get(`/documents/${docId.value}`)
    form.value.nama_lengkap = data?.resident?.nama_lengkap || ''
    form.value.tempat_lahir = data?.resident?.tempat_lahir || ''
    form.value.tanggal_lahir = data?.resident?.tanggal_lahir?.split('T')[0] || ''
    form.value.jenis_kelamin = data?.resident?.jenis_kelamin || ''
    form.value.agama = data?.resident?.agama || ''
    form.value.pekerjaan = data?.resident?.pekerjaan || ''
    form.value.alamat = data?.resident?.alamat || ''
    form.value.lokasi_hilang = data?.lokasi_hilang || ''
    form.value.petugas_pelapor_id = data?.petugas_pelapor_id || ''
    form.value.pejabat_persetuju_id = data?.pejabat_persetuju_id || ''
    items.value = (data?.lost_items || []).map((item) => ({
      nama_barang: item.nama_barang,
      deskripsi: item.deskripsi,
    }))
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal memuat data dokumen.'
  }
}

const addItem = () => {
  if (!selectedTemplate.value) {
    errorMessage.value = 'Pilih jenis barang terlebih dahulu.'
    return
  }
  if (selectedTemplate.value === 'LAINNYA') {
    if (!lainnya.value.nama_barang.trim()) {
      errorMessage.value = 'Nama barang lainnya wajib diisi.'
      return
    }
    items.value.push({
      nama_barang: lainnya.value.nama_barang,
      deskripsi: lainnya.value.deskripsi,
    })
    lainnya.value = { nama_barang: '', deskripsi: '' }
    selectedTemplate.value = ''
    return
  }

  const template = selectedTemplateConfig.value
  if (!template) return
  const parts = []
  let invalid = false
  template.fields_config.forEach((field) => {
    const value = dynamicValues.value[field.data_label] || ''
    if (field.required_length && !value) {
      invalid = true
    }
    if (value) {
      parts.push(`${field.data_label}: ${value}`)
    }
  })
  if (invalid) {
    errorMessage.value = 'Lengkapi field yang wajib diisi.'
    return
  }
  items.value.push({
    nama_barang: template.nama_barang,
    deskripsi: parts.join(', '),
  })
  dynamicValues.value = {}
  selectedTemplate.value = ''
}

const removeItem = (index) => {
  items.value.splice(index, 1)
}

const submit = async () => {
  errorMessage.value = ''
  if (items.value.length === 0) {
    errorMessage.value = 'Minimal satu barang hilang harus diisi.'
    return
  }
  if (!form.value.petugas_pelapor_id || !form.value.pejabat_persetuju_id) {
    errorMessage.value = 'Petugas pelapor dan pejabat persetuju wajib dipilih.'
    return
  }
  loading.value = true
  try {
    const payload = {
      ...form.value,
      petugas_pelapor_id: Number(form.value.petugas_pelapor_id),
      pejabat_persetuju_id: Number(form.value.pejabat_persetuju_id),
      items: items.value,
    }
    if (isEdit.value) {
      await api.put(`/documents/${docId.value}`, payload)
    } else {
      await api.post('/documents', payload)
    }
    router.push({ name: 'documents' })
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal menyimpan dokumen.'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await fetchOperators()
  await fetchTemplates()
  await fetchDocument()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">{{ isEdit ? 'Edit Surat' : 'Buat Surat Baru' }}</h1>
      <p class="text-sm text-slate-500">Isi data pemohon dan detail kehilangan.</p>
    </div>

    <div v-if="errorMessage" class="rounded-xl bg-red-50 px-4 py-2 text-sm text-red-600">
      {{ errorMessage }}
    </div>

    <form class="space-y-6" @submit.prevent="submit">
      <div class="grid gap-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm md:grid-cols-2">
        <div>
          <label class="text-sm font-medium text-slate-700">Nama Lengkap</label>
          <input v-model="form.nama_lengkap" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Tempat Lahir</label>
          <input v-model="form.tempat_lahir" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Tanggal Lahir</label>
          <input v-model="form.tanggal_lahir" type="date" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Jenis Kelamin</label>
          <select v-model="form.jenis_kelamin" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required>
            <option value="">Pilih</option>
            <option value="Laki-laki">Laki-laki</option>
            <option value="Perempuan">Perempuan</option>
          </select>
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Agama</label>
          <input v-model="form.agama" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Pekerjaan</label>
          <input v-model="form.pekerjaan" type="text" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required />
        </div>
        <div class="md:col-span-2">
          <label class="text-sm font-medium text-slate-700">Alamat</label>
          <textarea v-model="form.alamat" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" rows="3" required></textarea>
        </div>
      </div>

      <div class="grid gap-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm md:grid-cols-2">
        <div>
          <label class="text-sm font-medium text-slate-700">Petugas Pelapor</label>
          <select v-model="form.petugas_pelapor_id" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required>
            <option value="">Pilih Petugas</option>
            <option v-for="user in operators" :key="user.id" :value="user.id">
              {{ user.pangkat }} {{ user.nama_lengkap }} ({{ user.nrp }})
            </option>
          </select>
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700">Pejabat Persetuju</label>
          <select v-model="form.pejabat_persetuju_id" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" required>
            <option value="">Pilih Pejabat</option>
            <option v-for="user in operators" :key="user.id" :value="user.id">
              {{ user.pangkat }} {{ user.nama_lengkap }} ({{ user.nrp }})
            </option>
          </select>
        </div>
        <div class="md:col-span-2">
          <label class="text-sm font-medium text-slate-700">Lokasi Hilang</label>
          <textarea v-model="form.lokasi_hilang" class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" rows="2" required></textarea>
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-800">Barang Hilang</h2>
        <div class="mt-4 grid gap-3 md:grid-cols-3">
          <select v-model="selectedTemplate" class="rounded-xl border border-slate-200 px-3 py-2 text-sm">
            <option value="">Pilih Jenis Barang</option>
            <option v-for="template in itemTemplates" :key="template.id" :value="template.nama_barang">
              {{ template.nama_barang }}
            </option>
            <option value="LAINNYA">Lainnya...</option>
          </select>
          <button type="button" class="rounded-xl bg-primary-600 px-3 py-2 text-sm text-white" @click="addItem">Tambah</button>
        </div>
        <div v-if="selectedTemplate === 'LAINNYA'" class="mt-3 grid gap-3 md:grid-cols-2">
          <input v-model="lainnya.nama_barang" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Nama barang" />
          <input v-model="lainnya.deskripsi" type="text" class="rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Deskripsi" />
        </div>
        <div v-if="selectedTemplateConfig" class="mt-4 grid gap-3 md:grid-cols-2">
          <div v-for="field in selectedTemplateConfig.fields_config" :key="field.data_label">
            <label class="text-xs text-slate-500">{{ field.label }}</label>
            <input
              v-if="field.type === 'text' || field.type === 'number'"
              v-model="dynamicValues[field.data_label]"
              :type="field.type === 'number' ? 'number' : 'text'"
              class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm"
              :placeholder="field.placeholder || ''"
            />
            <textarea
              v-else-if="field.type === 'textarea'"
              v-model="dynamicValues[field.data_label]"
              class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm"
              rows="2"
              :placeholder="field.placeholder || ''"
            ></textarea>
            <input
              v-else-if="field.type === 'date'"
              v-model="dynamicValues[field.data_label]"
              type="date"
              class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm"
            />
            <select
              v-else-if="field.type === 'select'"
              v-model="dynamicValues[field.data_label]"
              class="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm"
            >
              <option value="">Pilih</option>
              <option v-for="opt in field.options || []" :key="opt" :value="opt">{{ opt }}</option>
            </select>
          </div>
        </div>
        <div class="mt-4">
          <table class="min-w-full text-sm">
            <thead class="text-left text-xs uppercase text-slate-500">
              <tr>
                <th class="px-2 py-2">Nama</th>
                <th class="px-2 py-2">Deskripsi</th>
                <th class="px-2 py-2">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="items.length === 0">
                <td colspan="3" class="px-2 py-3 text-center text-slate-500">Belum ada barang.</td>
              </tr>
              <tr v-for="(item, index) in items" :key="index" class="border-t border-slate-100">
                <td class="px-2 py-2">{{ item.nama_barang }}</td>
                <td class="px-2 py-2">{{ item.deskripsi }}</td>
                <td class="px-2 py-2">
                  <button type="button" class="text-xs text-red-600" @click="removeItem(index)">Hapus</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="flex justify-end gap-3">
        <RouterLink class="rounded-xl border border-slate-200 px-4 py-2 text-sm" to="/documents">Kembali</RouterLink>
        <button type="submit" class="rounded-xl bg-slate-900 px-4 py-2 text-sm text-white" :disabled="loading">
          {{ loading ? 'Menyimpan...' : 'Simpan Surat' }}
        </button>
      </div>
    </form>
  </div>
</template>
