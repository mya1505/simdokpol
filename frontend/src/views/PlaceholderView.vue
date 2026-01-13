<script setup>
const props = defineProps({
  title: { type: String, required: true },
})
</script>

<template>
  <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
    <h1 class="text-2xl font-semibold text-slate-800">{{ props.title }}</h1>
    <p class="mt-2 text-sm text-slate-500">
      Halaman ini sedang dimigrasi ke Vue. Fitur lengkap akan menyusul.
    </p>
  </div>
</template>
