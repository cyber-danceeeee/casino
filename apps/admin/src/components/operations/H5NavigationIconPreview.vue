<script setup lang="ts">
import { computed } from "vue";
import { Grid, Present, Share, User, Trophy } from "@element-plus/icons-vue";

const props = withDefaults(
  defineProps<{
    menuCode?: string;
    iconUrl?: string;
    size?: number;
  }>(),
  { menuCode: "", iconUrl: "", size: 38 },
);

const fallbackIcons = { home: Grid, offers: Present, world: Trophy, share: Share, profile: User };
const fallback = computed(
  () => fallbackIcons[props.menuCode as keyof typeof fallbackIcons] || Grid,
);
</script>

<template>
  <div
    class="h5-navigation-icon"
    :style="{ width: `${size}px`, height: `${size}px` }"
  >
    <el-image
      v-if="iconUrl"
      class="h5-navigation-icon__image"
      :src="iconUrl"
      fit="contain"
      :preview-src-list="[iconUrl]"
      preview-teleported
    />
    <el-icon v-else :size="Math.round(size * 0.62)">
      <component :is="fallback" />
    </el-icon>
  </div>
</template>

<style scoped>
.h5-navigation-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex: 0 0 auto;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  color: var(--el-color-primary);
  background: var(--el-fill-color-light);
}
.h5-navigation-icon__image {
  width: 100%;
  height: 100%;
}
</style>
