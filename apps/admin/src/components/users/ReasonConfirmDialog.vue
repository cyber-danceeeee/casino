<script setup lang="ts">
import { ref, watch } from "vue";
import { useAdminCenterText } from "@/composables/useAdminCenterText";

const props = defineProps<{
  modelValue: boolean;
  title: string;
  message: string;
  loading?: boolean;
}>();
const emit = defineEmits<{
  "update:modelValue": [value: boolean];
  confirm: [reason: string];
}>();
const reason = ref("");
const tr = useAdminCenterText("userCenter");

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) reason.value = "";
  },
);
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="460px"
    @close="emit('update:modelValue', false)"
  >
    <p class="reason-message">{{ message }}</p>
    <el-form label-position="left" label-width="112px">
      <el-form-item :label="tr('Operator reason')" required>
        <el-input
          v-model="reason"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          :placeholder="tr('Enter a reason for the audit log')"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">{{
        tr("Cancel")
      }}</el-button>
      <el-button
        type="primary"
        :loading="loading"
        :disabled="!reason.trim()"
        @click="emit('confirm', reason.trim())"
        >{{ tr("Confirm") }}</el-button
      >
    </template>
  </el-dialog>
</template>

<style scoped>
.reason-message {
  margin: 0 0 16px;
  color: var(--admin-text-secondary);
  line-height: 1.6;
}
</style>
