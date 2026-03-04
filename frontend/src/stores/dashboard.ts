import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getDashboard, listVersions } from '@/services/api'
import type { DashboardResponse } from '@/services/api'

export const useDashboardStore = defineStore('dashboard', () => {
  const versions = ref<string[]>([])
  const selectedV1 = ref<string>('')
  const selectedV2 = ref<string>('')
  const dashboard = ref<DashboardResponse | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchVersions(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      versions.value = await listVersions()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load versions'
    } finally {
      loading.value = false
    }
  }

  async function fetchDashboard(): Promise<void> {
    if (!selectedV1.value || !selectedV2.value) return
    loading.value = true
    error.value = null
    try {
      dashboard.value = await getDashboard(selectedV1.value, selectedV2.value)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load dashboard'
    } finally {
      loading.value = false
    }
  }

  return { versions, selectedV1, selectedV2, dashboard, loading, error, fetchVersions, fetchDashboard }
})

