<script setup lang="ts">
import { computed, ref, watch } from "vue";

const props = defineProps<{ h5Domain?: string; siteID?: number }>();
const loading = ref(true);

const previewURL = computed(() => {
  const raw = String(props.h5Domain || "").trim();
  if (!raw) return "";
  try {
    const url = new URL(/^https?:\/\//i.test(raw) ? raw : `https://${raw}`);
    url.searchParams.set("preview", "1");
    if (props.siteID) url.searchParams.set("siteId", String(props.siteID));
    url.searchParams.set("_previewAt", String(Date.now()));
    return url.toString();
  } catch {
    return "";
  }
});

watch(previewURL, () => { loading.value = true; });
</script>

<template>
  <div v-if="previewURL" v-loading="loading" class="h5-preview-frame">
    <iframe :src="previewURL" title="H5 Preview" @load="loading = false" />
  </div>
  <el-empty v-else description="H5 URL is not configured" />
</template>

<style scoped>
.h5-preview-frame{width:min(430px,100%);height:min(720px,68vh);margin:auto;border:1px solid var(--el-border-color);background:#101828;overflow:hidden}.h5-preview-frame iframe{display:block;width:100%;height:100%;border:0;background:#fff}
</style>
