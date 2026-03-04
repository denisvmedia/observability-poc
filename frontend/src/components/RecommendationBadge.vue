<template>
  <div class="rec-badge" :class="recommendation.winner ? 'rec-badge--winner' : 'rec-badge--tie'">
    <div class="rec-badge__title">
      <template v-if="recommendation.winner">
        🏆 Winner: <strong>{{ recommendation.winner }}</strong>
      </template>
      <template v-else> 🤝 Tie </template>
    </div>
    <div class="rec-badge__reason">{{ recommendation.reason }}</div>
    <div class="rec-badge__score">
      A: {{ recommendation.wins_a }} wins &nbsp;|&nbsp; B: {{ recommendation.wins_b }} wins
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Recommendation } from '@/services/api'

defineProps<{
  recommendation: Recommendation
}>()
</script>

<style lang="scss" scoped>
.rec-badge {
  border-radius: 10px;
  padding: 1.25rem 1.5rem;
  margin-bottom: 1.5rem;

  &--winner {
    background: #e6f4ea;
    border: 1px solid #81c784;
  }

  &--tie {
    background: #f5f5f5;
    border: 1px solid #bdbdbd;
  }

  &__title {
    font-size: 1.25rem;
    font-weight: 700;
    margin-bottom: 0.5rem;
  }

  &__reason {
    font-size: 0.95rem;
    color: #444;
    margin-bottom: 0.25rem;
  }

  &__score {
    font-size: 0.85rem;
    color: #777;
  }
}
</style>

