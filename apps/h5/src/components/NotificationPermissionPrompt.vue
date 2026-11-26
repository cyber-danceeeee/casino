<script setup lang="ts">
import { computed, ref } from 'vue'
import notificationBell from '../assets/images/notification-bell.png'
import { runtimeConfig } from '../services/runtimeConfig'
import {
  markNotificationPromptSeen,
  requestNotificationPermission,
} from '../services/notificationPermission'

const emit = defineEmits<{
  complete: []
}>()

const submitting = ref(false)
const isChinese = computed(() => {
  const language = runtimeConfig.value.language.current || runtimeConfig.value.language.default
  return language.toLowerCase().startsWith('zh')
})
const content = computed(() => isChinese.value
  ? {
      title: '开启通知，不错过免费奖励！',
      cancel: '暂不开启',
      confirm: '好的',
      label: '开启通知',
    }
  : {
      title: 'Enable Alerts and Never Miss Your Free Rewards!',
      cancel: 'Not now',
      confirm: 'Sounds good',
      label: 'Enable notifications',
    })

const cancel = () => {
  markNotificationPromptSeen()
  emit('complete')
}

const confirm = async () => {
  if (submitting.value) return
  submitting.value = true
  try {
    await requestNotificationPermission()
  } finally {
    submitting.value = false
    emit('complete')
  }
}
</script>

<template>
  <div class="notification-tip-backdrop" role="presentation">
    <section
      class="notic-user-tip-warp"
      role="dialog"
      aria-modal="true"
      :aria-label="content.label"
    >
      <div class="notic-user-tip-top">
        <img :src="notificationBell" alt="" width="48" height="48" />
        <div class="notic-user-tip-data">{{ content.title }}</div>
      </div>
      <div class="button-warp">
        <button type="button" class="web-push-tip-cancel" @click="cancel">
          {{ content.cancel }}
        </button>
        <button type="button" class="confirm" :disabled="submitting" @click="confirm">
          {{ content.confirm }}
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped lang="scss">
.notification-tip-backdrop {
  position: fixed;
  inset: 0;
  z-index: 50020;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  max-width: 430px;
  margin: 0 auto;
  padding: 18px;
  background: rgb(0 6 20 / 72%);
  backdrop-filter: blur(1px);
}

.notic-user-tip-warp {
  width: min(100%, 340px);
  overflow: hidden;
  border: 1px solid rgb(255 255 255 / 8%);
  border-radius: 12px;
  background: #1e2438;
  box-shadow: 0 16px 42px rgb(0 0 0 / 42%);
  animation: notification-tip-enter 220ms ease-out both;
}

.notic-user-tip-top {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  padding: 28px 24px 24px;
  text-align: center;
}

.notic-user-tip-top img {
  display: block;
  width: 48px;
  height: 48px;
  object-fit: contain;
}

.notic-user-tip-data {
  color: rgb(79 131 255);
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
}

.button-warp {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 48px;
  border-top: 1px solid rgb(255 255 255 / 9%);
}

.button-warp button {
  min-width: 0;
  border: 0;
  border-radius: 0;
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.web-push-tip-cancel {
  color: rgb(224 224 224);
  background: transparent;
}

.confirm {
  color: #fff;
  background: rgb(79 131 255);
}

.confirm:disabled {
  cursor: default;
  opacity: 0.65;
}

@keyframes notification-tip-enter {
  from {
    opacity: 0;
    transform: translateY(10px) scale(0.96);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>
