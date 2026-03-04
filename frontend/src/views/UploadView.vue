<template>
  <div class="upload-view">
    <h1>Upload Playback Sessions</h1>
    <p>Select an <strong>.xlsx</strong> file exported from your analytics system.</p>

    <div class="upload-view__form">
      <input
        ref="fileInput"
        type="file"
        accept=".xlsx"
        class="upload-view__input"
        @change="onFileChange"
      />
      <button
        class="upload-view__btn"
        :disabled="!selectedFile || uploading"
        @click="onUpload"
      >
        {{ uploading ? 'Uploading…' : 'Upload' }}
      </button>
    </div>

    <div v-if="error" class="upload-view__error">{{ error }}</div>

    <div v-if="result" class="upload-view__result">
      <p>✅ Inserted: <strong>{{ result.rows_inserted }}</strong> rows</p>
      <p v-if="result.rows_skipped > 0">⚠ Skipped: {{ result.rows_skipped }} rows</p>
      <ul v-if="result.errors.length > 0">
        <li v-for="(e, i) in result.errors" :key="i">{{ e }}</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { uploadFile } from '@/services/api'
import type { UploadResult } from '@/services/api'

const router = useRouter()
const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const uploading = ref(false)
const error = ref<string | null>(null)
const result = ref<UploadResult | null>(null)

function onFileChange(event: Event): void {
  const input = event.target as HTMLInputElement
  selectedFile.value = input.files?.[0] ?? null
  result.value = null
  error.value = null
}

async function onUpload(): Promise<void> {
  if (!selectedFile.value) return
  uploading.value = true
  error.value = null
  try {
    result.value = await uploadFile(selectedFile.value)
    await router.push('/dashboard')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Upload failed'
  } finally {
    uploading.value = false
  }
}
</script>

<style lang="scss" scoped>
.upload-view {
  max-width: 560px;

  h1 {
    margin-bottom: 0.5rem;
  }

  &__form {
    display: flex;
    gap: 1rem;
    align-items: center;
    margin: 1.5rem 0;
  }

  &__btn {
    padding: 0.5rem 1.25rem;
    background: #1976d2;
    color: #fff;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.95rem;

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  }

  &__error {
    color: #c62828;
    margin-bottom: 1rem;
  }

  &__result {
    padding: 1rem;
    background: #f1f8e9;
    border-radius: 8px;
  }
}
</style>

