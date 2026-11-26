<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { dashboardItem, menuGroups } from '@/data/adminMenu'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  navigate: []
}>()

const route = useRoute()
const { t } = useI18n()
const openGroups = ref<string[]>([])

const activeGroupKey = computed(() => {
  const matched = menuGroups.find((group) => group.children.some((item) => item.path === route.path))
  return matched?.key || ''
})

watch(activeGroupKey, (key) => {
  if (key && !openGroups.value.includes(key)) {
    openGroups.value = [...openGroups.value, key]
  }
}, { immediate: true })

const isGroupOpen = (key: string) => openGroups.value.includes(key)

const toggleGroup = (key: string) => {
  openGroups.value = isGroupOpen(key)
    ? openGroups.value.filter((item) => item !== key)
    : [...openGroups.value, key]
}
</script>

<template>
  <aside class="admin-sidebar">
    <RouterLink class="admin-sidebar__brand" :to="dashboardItem.path" @click="$emit('navigate')">
      <span class="admin-sidebar__mark">88</span>
      <strong>Casino Admin</strong>
    </RouterLink>

    <nav class="admin-sidebar__nav" aria-label="Admin navigation">
      <RouterLink
        class="admin-sidebar__dashboard"
        :class="{ 'router-link-active': route.path === dashboardItem.path }"
        :to="dashboardItem.path"
        @click="$emit('navigate')"
      >
        <span>DB</span>
        <em>{{ t(dashboardItem.i18nKey) }}</em>
      </RouterLink>

      <section v-for="group in menuGroups" :key="group.key" class="admin-menu-group" :class="{ 'is-open': isGroupOpen(group.key) }">
        <button type="button" class="admin-menu-group__button" @click="toggleGroup(group.key)">
          <span>{{ group.icon }}</span>
          <em>{{ t(group.i18nKey) }}</em>
          <i />
        </button>

        <div class="admin-menu-group__children">
          <RouterLink
            v-for="item in group.children"
            :key="item.key"
            :to="item.path"
            @click="$emit('navigate')"
          >
            {{ t(item.i18nKey) }}
          </RouterLink>
        </div>
      </section>
    </nav>
  </aside>
</template>
