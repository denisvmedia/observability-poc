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

    <section class="dashboard-view__legend">
      <h2 class="dashboard-view__legend-title">Metrics reference</h2>
      <table class="dashboard-view__legend-table">
        <thead>
          <tr>
            <th>Metric</th>
            <th>What it measures</th>
            <th>Better when</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in KPI_META" :key="m.name">
            <td class="dashboard-view__legend-name">{{ m.name }}</td>
            <td>{{ m.description }}</td>
            <td :class="m.lowerBetter ? 'dashboard-view__legend-dir--lower' : 'dashboard-view__legend-dir--higher'">
              {{ m.lowerBetter ? '↓ lower' : '↑ higher' }}
            </td>
          </tr>
        </tbody>
      </table>
    </section>
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


const KPI_META = [
  { name: 'VSF Rate',        description: 'Video Start Failure — share of attempts where playback never started',              lowerBetter: true  },
  { name: 'VPF Rate',        description: 'Video Playback Failure — share of plays that failed mid-stream',                   lowerBetter: true  },
  { name: 'CIRR',            description: 'Complete Inactivity Rebuffering Rate — share of plays with full rebuffering stops', lowerBetter: true  },
  { name: 'Avg VST',         description: 'Average Video Start Time — seconds from request to first frame',                   lowerBetter: true  },
  { name: 'Play Rate',       description: 'Share of attempts where playback actually started',                                lowerBetter: false },
  { name: 'Completion Rate', description: 'Share of started videos played to completion',                                     lowerBetter: false },
] as const
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

  &__legend {
    margin-top: 2.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid #e0e0e0;
  }

  &__legend-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: #555;
    margin-bottom: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  &__legend-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;

    th {
      text-align: left;
      padding: 0.4rem 0.75rem;
      border-bottom: 2px solid #e0e0e0;
      color: #777;
      font-weight: 600;
      white-space: nowrap;
    }

    td {
      padding: 0.4rem 0.75rem;
      border-bottom: 1px solid #f0f0f0;
      color: #444;
      vertical-align: top;
    }
  }

  &__legend-name {
    white-space: nowrap;
    font-weight: 500;
    color: #222;
  }

  &__legend-dir {
    &--lower {
      color: #1e7e34;
      font-weight: 600;
      white-space: nowrap;
    }

    &--higher {
      color: #1565c0;
      font-weight: 600;
      white-space: nowrap;
    }
  }
}
</style>

