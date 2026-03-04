<template>
  <div class="app">
    <nav class="app__nav">
      <RouterLink to="/">Upload</RouterLink>
      <RouterLink to="/dashboard">Dashboard</RouterLink>
    </nav>
    <main class="app__main">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<style lang="scss">
.app {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  font-family: sans-serif;

  &__nav {
    display: flex;
    gap: 1.5rem;
    margin-bottom: 2rem;
    border-bottom: 1px solid #e0e0e0;
    padding-bottom: 1rem;

    a {
      text-decoration: none;
      color: #1976d2;
      font-weight: 500;

      &.router-link-active {
        color: #0d47a1;
        text-decoration: underline;
      }
    }
  }
}
</style>

