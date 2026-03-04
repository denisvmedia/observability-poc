<template>
  <div class="kpi-card">
    <div class="kpi-card__name">{{ name }}</div>
    <div class="kpi-card__values">
      <div
        class="kpi-card__value"
        :class="{
          'kpi-card__value--winner': winner === 'A',
          'kpi-card__value--loser': winner === 'B',
        }"
      >
        <span class="kpi-card__label">A</span>
        <span class="kpi-card__number">{{ formatted(valueA) }}{{ unit }}</span>
      </div>
      <div
        class="kpi-card__value"
        :class="{
          'kpi-card__value--winner': winner === 'B',
          'kpi-card__value--loser': winner === 'A',
        }"
      >
        <span class="kpi-card__label">B</span>
        <span class="kpi-card__number">{{ formatted(valueB) }}{{ unit }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    name: string
    valueA: number
    valueB: number
    winner: 'A' | 'B' | 'tie'
    unit?: string
  }>(),
  { unit: '' },
)

function formatted(v: number): string {
  return v.toFixed(3)
}
</script>

<style lang="scss" scoped>
.kpi-card {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 1rem;

  &__name {
    font-weight: 600;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    color: #555;
  }

  &__values {
    display: flex;
    gap: 1rem;
  }

  &__value {
    flex: 1;
    text-align: center;
    padding: 0.5rem;
    border-radius: 6px;
    background: #f5f5f5;

    &--winner {
      background: #e6f4ea;
      color: #1e7e34;
    }

    &--loser {
      background: #fdecea;
      color: #c62828;
    }
  }

  &__label {
    display: block;
    font-size: 0.75rem;
    font-weight: 700;
    margin-bottom: 0.25rem;
  }

  &__number {
    font-size: 1rem;
    font-weight: 500;
  }
}
</style>

