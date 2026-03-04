<template>
  <div class="dashboard-view">
    <h1>Dashboard</h1>

    <div class="dashboard-view__controls">
      <select v-model="store.selectedV1" class="dashboard-view__select">
        <option value="" disabled>Version A</option>
        <option v-for="v in store.versions" :key="v" :value="v">{{ v }}</option>
      </select>
      <select v-model="store.selectedV2" class="dashboard-view__select">
        <option value="" disabled>Version B</option>
        <option v-for="v in store.versions" :key="v" :value="v">{{ v }}</option>
      </select>
      <button
        class="dashboard-view__btn"
        :disabled="!canCompare || store.loading"
        @click="store.fetchDashboard()"
      >
        {{ store.loading ? 'Loading…' : 'Compare' }}
      </button>
    </div>

    <div v-if="store.error" class="dashboard-view__error">{{ store.error }}</div>

    <template v-if="store.dashboard">
      <RecommendationBadge :recommendation="store.dashboard.recommendation" />

      <RadarChart
        :version-a="store.dashboard.versions.a"
        :version-b="store.dashboard.versions.b"
      />

      <div class="dashboard-view__alerts">
        <AlertBanner :alerts="store.dashboard.alerts.a" :version-label="`Version A: ${store.selectedV1}`" />
        <AlertBanner :alerts="store.dashboard.alerts.b" :version-label="`Version B: ${store.selectedV2}`" />
      </div>

      <div class="dashboard-view__grid">
        <KPICard
          v-for="dim in store.dashboard.recommendation.dimensions"
          :key="dim.name"
          :name="dim.name"
          :value-a="dim.version_a"
          :value-b="dim.version_b"
          :winner="dim.winner"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useDashboardStore } from '@/stores/dashboard'
import KPICard from '@/components/KPICard.vue'
import AlertBanner from '@/components/AlertBanner.vue'
import RecommendationBadge from '@/components/RecommendationBadge.vue'
import RadarChart from '@/components/RadarChart.vue'

const store = useDashboardStore()

onMounted(() => {
  store.fetchVersions()
})

const canCompare = computed(
  () => store.selectedV1 && store.selectedV2 && store.selectedV1 !== store.selectedV2,
)

</script>

<style lang="scss" scoped>
.dashboard-view {
  &__controls {
    display: flex;
    gap: 1rem;
    align-items: center;
    margin: 1.5rem 0;
    flex-wrap: wrap;
  }

  &__select {
    padding: 0.5rem 0.75rem;
    border: 1px solid #ccc;
    border-radius: 6px;
    font-size: 0.95rem;
    min-width: 160px;
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

  &__alerts {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
    flex-wrap: wrap;

    > * {
      flex: 1;
      min-width: 260px;
    }
  }

  &__grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 1rem;
  }
}
</style>

