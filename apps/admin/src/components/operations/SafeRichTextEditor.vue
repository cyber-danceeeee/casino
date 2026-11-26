<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from "vue";
import { Link, List } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";

const props = defineProps<{ modelValue: string }>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
const { t } = useI18n();
const editor = ref<HTMLElement | null>(null);

function sync() {
  emit('update:modelValue', editor.value?.innerHTML || '');
}

function command(name: string, value?: string) {
  editor.value?.focus();
  document.execCommand(name, false, value);
  sync();
}

function addLink() {
  const value = window.prompt(t('announcementLibrary.editor.linkPrompt'), 'https://');
  if (value) command('createLink', value);
}

function pastePlainText(event: ClipboardEvent) {
  event.preventDefault();
  command('insertText', event.clipboardData?.getData('text/plain') || '');
}

watch(() => props.modelValue, async (value) => {
  await nextTick();
  if (editor.value && document.activeElement !== editor.value && editor.value.innerHTML !== value) editor.value.innerHTML = value || '';
});

onMounted(() => { if (editor.value) editor.value.innerHTML = props.modelValue || ''; });
</script>

<template>
  <div class="rich-editor">
    <div class="toolbar">
      <el-tooltip :content="t('announcementLibrary.editor.bold')"><button type="button" class="text-tool" @click="command('bold')"><strong>B</strong></button></el-tooltip>
      <el-tooltip :content="t('announcementLibrary.editor.italic')"><button type="button" class="text-tool" @click="command('italic')"><em>I</em></button></el-tooltip>
      <el-tooltip :content="t('announcementLibrary.editor.bulletList')"><el-button text :icon="List" @click="command('insertUnorderedList')" /></el-tooltip>
      <el-tooltip :content="t('announcementLibrary.editor.numberedList')"><button type="button" class="text-tool" @click="command('insertOrderedList')">1.</button></el-tooltip>
      <el-tooltip :content="t('announcementLibrary.editor.link')"><el-button text :icon="Link" @click="addLink" /></el-tooltip>
    </div>
    <div ref="editor" class="editor-body" contenteditable="true" role="textbox" @input="sync" @blur="sync" @paste="pastePlainText" />
  </div>
</template>

<style scoped>
.rich-editor { border: 1px solid #d0d5dd; border-radius: 5px; overflow: hidden; background: #fff; }
.toolbar { display: flex; align-items: center; gap: 3px; min-height: 38px; padding: 3px 7px; border-bottom: 1px solid #e4e7ec; background: #f9fafb; }
.text-tool { width: 32px; height: 32px; border: 0; background: transparent; color: #344054; cursor: pointer; font-size: 14px; }
.text-tool:hover { background: #eaecf0; }
.editor-body { min-height: 180px; padding: 12px; outline: none; line-height: 1.6; color: #18212f; }
</style>
