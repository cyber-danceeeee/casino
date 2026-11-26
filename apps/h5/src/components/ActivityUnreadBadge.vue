<template>
  <span class="activity-unread-badge" aria-hidden="true">
    <span class="activity-unread-badge__core">
      <span class="activity-unread-badge__ping" />
      <span class="activity-unread-badge__dot" />
    </span>
  </span>
</template>

<style scoped>
.activity-unread-badge {
  pointer-events: none;
  position: absolute;
  top: -7px;
  right: -3px;
  z-index: 4;
  display: flex;
  width: 14px;
  height: 14px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #090f1f;
}

.activity-unread-badge__core {
  position: relative;
  display: inline-flex;
  width: 8px;
  height: 8px;
}

.activity-unread-badge__ping,
.activity-unread-badge__dot {
  position: absolute;
  inset: 0;
  border-radius: 999px;
  background: #ef263d;
}

.activity-unread-badge__ping {
  animation: activityUnreadPing 1.35s cubic-bezier(0, 0, 0.2, 1) infinite;
}

@keyframes activityUnreadPing {
  75%,
  100% {
    opacity: 0;
    transform: scale(2.1);
  }
}
</style>
