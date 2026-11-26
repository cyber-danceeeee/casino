import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { isLoggedIn } from '../mock/session'
import { getSafeRedirect } from '../utils/authRedirect'
import ActivityDetail from '../views/ActivityDetail.vue'
import AuthPage from '../views/AuthPage.vue'
import GameCategory from '../views/GameCategory.vue'
import GamePlay from '../views/GamePlay.vue'
import Home from '../views/Home.vue'
import Offers from '../views/Offers.vue'
import Profile from '../views/Profile.vue'
import ProfileSubPage from '../views/ProfileSubPage.vue'
import Share from '../views/Share.vue'

const requiresAuth = { requiresAuth: true }

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: Home,
  },
  {
    path: '/offers',
    name: 'offers',
    component: Offers,
  },
  {
    path: '/activity/:slug(.*)*',
    name: 'activity-detail',
    component: ActivityDetail,
  },
  {
    path: '/share',
    name: 'share',
    component: Share,
  },
  {
    path: '/vip',
    name: 'vip',
    component: Offers,
  },
  {
    path: '/profile',
    name: 'profile',
    component: Profile,
    meta: requiresAuth,
  },
  {
    path: '/wallet/records',
    name: 'wallet-records',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/wallet/deposit',
    name: 'wallet-deposit',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/cash-center',
    name: 'cash-center',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/bonus-center',
    name: 'bonus-center',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/bet-records',
    name: 'bet-records',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/security',
    name: 'security',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/security/bank-account',
    name: 'bank-account',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/security/phone',
    name: 'security-phone',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/security/password',
    name: 'security-password',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/support',
    name: 'support',
    component: ProfileSubPage,
  },
  {
    path: '/about',
    name: 'about',
    component: ProfileSubPage,
  },
  {
    path: '/auth/agreement',
    name: 'user-agreement',
    component: ProfileSubPage,
  },
  {
    path: '/settings',
    name: 'settings',
    component: ProfileSubPage,
    meta: requiresAuth,
  },
  {
    path: '/game/play',
    name: 'game-play',
    component: GamePlay,
  },
  {
    path: '/games/:category',
    name: 'game-category',
    component: GameCategory,
  },
  {
    path: '/auth/login',
    name: 'auth-login',
    component: AuthPage,
  },
  {
    path: '/auth/register',
    name: 'auth-register',
    component: AuthPage,
  },
  {
    path: '/popup-preview/:componentCode',
    name: 'popup-component-preview',
    component: () => import('../views/PopupComponentPreview.vue'),
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    void from
    if (to.name === 'game-category') return { left: 0, top: 0 }
    if (savedPosition) return savedPosition
    return { left: 0, top: 0 }
  },
})

NProgress.configure({
  minimum: 0.14,
  showSpinner: false,
  trickleSpeed: 180,
})

router.beforeEach((to) => {
  NProgress.start()

  if (to.meta.requiresAuth && !isLoggedIn.value) {
    return {
      name: 'auth-login',
      query: {
        redirect: to.fullPath,
      },
    }
  }

  if ((to.name === 'auth-login' || to.name === 'auth-register') && isLoggedIn.value) {
    const redirect = getSafeRedirect(to.query.redirect)
    return redirect || { name: 'home' }
  }
})

router.afterEach(() => {
  NProgress.done()
})

router.onError(() => {
  NProgress.done()
})

export default router
