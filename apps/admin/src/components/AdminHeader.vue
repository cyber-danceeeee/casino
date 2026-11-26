<script setup lang="ts">
defineProps<{
  title: string
}>()

defineEmits<{
  toggleSidebar: []
  toggleMobileMenu: []
}>()
</script>

<template>
  <header class="admin-header">
    <div class="admin-header__left">
      <button class="icon-button admin-header__mobile-menu" type="button" aria-label="Open menu" @click="$emit('toggleMobileMenu')">
        <span />
        <span />
        <span />
      </button>
      <button class="icon-button admin-header__collapse" type="button" aria-label="Toggle sidebar" @click="$emit('toggleSidebar')">
        <span />
        <span />
      </button>
      <div>
        <p class="admin-header__eyebrow">Casino Platform Back Office</p>
        <h1>{{ title }}</h1>
      </div>
    </div>

    <div class="admin-header__search">
      <span>Search</span>
      <input placeholder="玩家编号 / 订单号 / Provider交易号" />
      <kbd>⌘K</kbd>
    </div>

    <div class="admin-header__actions">
      <button class="ghost-button" type="button">同步游戏</button>
      <button class="ghost-button" type="button">导出报表</button>
      <button class="admin-header__user" type="button">
        <span>AD</span>
        <strong>Admin</strong>
      </button>
    </div>
  </header>
</template>
