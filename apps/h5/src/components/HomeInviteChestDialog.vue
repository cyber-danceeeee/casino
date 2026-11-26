<script setup lang="ts">
import { localDialogImage } from '../utils/localAssets'

const chestMain = localDialogImage('offers-invite-chest-detail-1.webp')
const chestSide = localDialogImage('offers-invite-chest-detail-2.webp')
const chestGold1 = localDialogImage('offers-invite-chest-gold-1.webp')
const chestGold2 = localDialogImage('offers-invite-chest-gold-2.webp')
const chestGold3 = localDialogImage('offers-invite-chest-gold-3.webp')

const socialItems = [
  { label: 'Facebook', image: localDialogImage('social-facebook.webp') },
  { label: 'Messenger', image: localDialogImage('social-messenger.webp') },
  { label: 'WhatsApp', image: localDialogImage('social-whatsapp.webp') },
  { label: 'Telegram', image: localDialogImage('social-telegram.webp') },
  { label: 'Instagram', image: localDialogImage('social-instagram.webp') },
  { label: 'TikTok', image: localDialogImage('social-tiktok.webp') },
  { label: 'X', image: localDialogImage('social-x.webp') },
]

defineEmits<{
  close: []
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="invite-overlay">
      <section class="invite-chest-dialog" aria-label="Invite Chest promotion">
        <article class="invite-chest-dialog__card">
          <button type="button" class="invite-chest-dialog__close" aria-label="Close" @click="$emit('close')">
            <span />
          </button>

          <div class="invite-chest-dialog__hero" aria-hidden="true">
            <img class="invite-chest-dialog__side" :src="chestSide" alt="" />
            <img class="invite-chest-dialog__main" :src="chestMain" alt="" />
            <img class="invite-chest-dialog__gold invite-chest-dialog__gold--1" :src="chestGold1" alt="" />
            <img class="invite-chest-dialog__gold invite-chest-dialog__gold--2" :src="chestGold2" alt="" />
            <img class="invite-chest-dialog__gold invite-chest-dialog__gold--3" :src="chestGold3" alt="" />
          </div>

          <div class="invite-chest-dialog__ribbon">
            <span>Invite Chest</span>
          </div>

          <div class="invite-chest-dialog__panel">
            <h2>Invite Friends To Earn Commission, Earn Unlimited Rewards!</h2>
            <div class="invite-chest-dialog__socials">
              <button v-for="item in socialItems" :key="item.label" type="button" :aria-label="item.label">
                <img :src="item.image" :alt="item.label" />
              </button>
              <button type="button" aria-label="Share">
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M14 5 21 12l-7 7v-4.5C8.9 14.5 5.5 16.2 3 20c.8-6.2 4.3-10.5 11-10.5V5Z" />
                </svg>
              </button>
            </div>
            <p>Tap An Icon Below<br />To Share The Link.</p>
          </div>
        </article>
      </section>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.invite-chest-dialog {
  position: fixed;
  inset: 0;
  z-index: 880;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: rgba(4, 8, 20, 0.72);
  backdrop-filter: blur(2px);
}

.invite-chest-dialog__card {
  position: relative;
  width: min(calc(100vw - 54px), 318px);
  padding-top: 128px;
  filter: drop-shadow(0 18px 24px rgba(20, 34, 86, 0.5));
  animation: invite-card-in 420ms cubic-bezier(0.2, 1, 0.3, 1);
}

.invite-chest-dialog__close {
  position: absolute;
  top: 48px;
  right: -14px;
  z-index: 8;
  display: grid;
  width: 52px;
  height: 52px;
  place-items: center;
  border: 2px solid rgba(255, 255, 255, 0.9);
  border-radius: 999px;
  background: rgba(48, 57, 83, 0.52);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.18);

  span,
  &::after {
    position: absolute;
    width: 26px;
    height: 4px;
    border-radius: 999px;
    background: #ffffff;
    content: "";
  }

  span {
    transform: rotate(45deg);
  }

  &::after {
    transform: rotate(-45deg);
  }
}

.invite-chest-dialog__hero {
  pointer-events: none;
  position: absolute;
  top: 0;
  left: 50%;
  z-index: 4;
  width: 260px;
  height: 178px;
  transform: translateX(-50%);

  &::before {
    position: absolute;
    inset: -34px -38px 22px;
    z-index: -1;
    border-radius: 50%;
    background: radial-gradient(circle, rgba(255, 245, 176, 0.7), rgba(78, 122, 255, 0.18) 48%, transparent 70%);
    animation: invite-halo 1.9s ease-in-out infinite;
    content: "";
  }
}

.invite-chest-dialog__side,
.invite-chest-dialog__main,
.invite-chest-dialog__gold {
  position: absolute;
  object-fit: contain;
}

.invite-chest-dialog__side {
  top: 24px;
  left: 50%;
  width: 214px;
  transform: translateX(-50%);
  animation: invite-float 2.3s ease-in-out infinite;
}

.invite-chest-dialog__main {
  top: 6px;
  left: 50%;
  width: 238px;
  transform: translateX(-50%);
  animation: invite-float 2.1s ease-in-out infinite reverse;
}

.invite-chest-dialog__gold--1 {
  left: 26px;
  top: 69px;
  width: 54px;
}

.invite-chest-dialog__gold--2 {
  right: 32px;
  top: 66px;
  width: 45px;
}

.invite-chest-dialog__gold--3 {
  right: 82px;
  top: 96px;
  width: 48px;
}

.invite-chest-dialog__gold {
  animation: invite-gold 1.4s ease-in-out infinite;
}

.invite-chest-dialog__ribbon {
  position: relative;
  z-index: 6;
  display: grid;
  height: 66px;
  place-items: center;
  margin: 0 10px -18px;
  color: #ffffff;
  text-align: center;
  filter: drop-shadow(0 5px 0 rgba(44, 54, 120, 0.45));

  &::before,
  &::after {
    position: absolute;
    top: 18px;
    width: 62px;
    height: 34px;
    background: linear-gradient(180deg, #7f91ff 0%, #5264d6 100%);
    clip-path: polygon(0 0, 100% 0, 82% 50%, 100% 100%, 0 100%, 18% 50%);
    content: "";
  }

  &::before {
    left: -6px;
    transform: rotate(-9deg);
  }

  &::after {
    right: -6px;
    transform: scaleX(-1) rotate(-9deg);
  }

  span {
    position: relative;
    z-index: 7;
    min-width: 210px;
    border-radius: 50% / 42%;
    padding: 12px 22px 13px;
    background: linear-gradient(180deg, #7d91ff 0%, #435ed8 100%);
    font-family: "Poetsen One", Montserrat, sans-serif;
    font-size: 28px;
    font-weight: 900;
    line-height: 1;
    text-shadow: 0 2px 0 #263b9d;
  }
}

.invite-chest-dialog__panel {
  position: relative;
  z-index: 2;
  overflow: hidden;
  border-radius: 0 0 28px 28px;
  padding: 42px 27px 26px;
  text-align: center;
  background: linear-gradient(180deg, #ffffff 0%, #edf7ff 100%);
  box-shadow:
    inset 0 0 0 4px rgba(42, 74, 140, 0.12),
    0 0 0 3px rgba(255, 255, 255, 0.24);

  h2 {
    margin: 0;
    color: #2a56a9;
    font-size: 17px;
    font-weight: 900;
    line-height: 1.28;
  }

  p {
    margin: 18px 0 0;
    color: #2f4f91;
    font-size: 15px;
    font-weight: 500;
    line-height: 1.2;
  }
}

.invite-chest-dialog__socials {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px 16px;
  margin-top: 24px;

  button {
    display: grid;
    width: 51px;
    height: 51px;
    place-items: center;
    justify-self: center;
    overflow: hidden;
    border-radius: 12px;
    padding: 0;
    background: linear-gradient(135deg, #16b9ff 0%, #4b6dff 100%);
    box-shadow: 0 5px 10px rgba(25, 65, 142, 0.22);
    transition: transform 180ms ease;

    &:active {
      transform: scale(0.94);
    }
  }

  img,
  svg {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  svg {
    width: 27px;
    height: 27px;
    fill: #ffffff;
  }
}

.invite-overlay-enter-active,
.invite-overlay-leave-active {
  transition: opacity 220ms ease;
}

.invite-overlay-enter-from,
.invite-overlay-leave-to {
  opacity: 0;
}

@keyframes invite-card-in {
  from {
    opacity: 0;
    transform: translateY(24px) scale(0.88);
  }
}

@keyframes invite-halo {
  50% {
    opacity: 0.45;
    transform: scale(1.06);
  }
}

@keyframes invite-float {
  50% {
    transform: translateX(-50%) translateY(-6px);
  }
}

@keyframes invite-gold {
  50% {
    transform: translateY(-8px) rotate(8deg);
  }
}

@media (max-width: 374px) {
  .invite-chest-dialog__card {
    width: min(calc(100vw - 44px), 300px);
  }

  .invite-chest-dialog__panel {
    padding-right: 22px;
    padding-left: 22px;
  }

  .invite-chest-dialog__socials {
    gap: 10px 12px;
  }
}
</style>
