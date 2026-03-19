import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue'),
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('@/views/DashboardView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('@/views/ProfileView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/profile/settings',
    name: 'profile-settings',
    component: () => import('@/views/ProfileSettingsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    // Public route: receives the OAuth redirect from the backend and processes
    // the hash fragment (#token=…, #pre_registration_token=…, or #error=…).
    path: '/auth/callback',
    name: 'auth-callback',
    component: () => import('@/views/AuthCallbackView.vue'),
  },
  {
    path: '/terms-of-service',
    name: 'terms-of-service',
    component: () => import('@/views/TermsOfServiceView.vue'),
  },
  {
    path: '/privacy-policy',
    name: 'privacy-policy',
    component: () => import('@/views/PrivacyPolicyView.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()

  // Protected route — unauthenticated users are sent to the home page.
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'home' }
  }

  // Home page — authenticated users are sent directly to the dashboard.
  if (to.name === 'home' && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }
})

export default router
