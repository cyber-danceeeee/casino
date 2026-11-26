<script setup lang="ts">
import { computed } from "vue";
import UserDetailPage from "@/views/users/UserDetailPage.vue";
import { useAdminCenterText } from "@/composables/useAdminCenterText";

const props = defineProps<{ modelValue: boolean; userId: number }>();
const emit = defineEmits<{ "update:modelValue": [value: boolean] }>();
const tr = useAdminCenterText("userCenter");
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="tr('User Detail')"
    width="min(1180px, 94vw)"
    destroy-on-close
    append-to-body
    class="user-detail-dialog"
  >
    <UserDetailPage v-if="userId" :user-id="userId" embedded />
  </el-dialog>
</template>

