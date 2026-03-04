<template>
  <div v-if="alerts.length > 0" class="alert-banner">
    <div class="alert-banner__header">⚠ {{ versionLabel }}</div>
    <div v-for="alert in alerts" :key="alert.code" class="alert-banner__item">
      <span class="alert-banner__code">{{ alert.code }}</span>
      <span class="alert-banner__message">{{ alert.message }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Alert } from '@/services/api'

defineProps<{
  alerts: Alert[]
  versionLabel: string
}>()
</script>

<style lang="scss" scoped>
.alert-banner {
  border: 1px solid #f0ad4e;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  background: #fff8e1;

  &__header {
    font-weight: 700;
    margin-bottom: 0.5rem;
    color: #7a5200;
  }

  &__item {
    display: flex;
    gap: 0.5rem;
    padding: 0.25rem 0;
    font-size: 0.875rem;
    color: #555;
  }

  &__code {
    font-weight: 600;
    color: #b45309;
    white-space: nowrap;
  }
}
</style>

