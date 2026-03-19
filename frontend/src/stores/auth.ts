import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { OpenAPI } from '@/lib/api/core/OpenAPI'
import { DefaultService } from '@/lib/api/services/DefaultService'

interface User {
  id: string
  username: string
  email: string
}

const BACKEND_URL = import.meta.env.APP_BASE_URL as string ?? ''

// Configure the generated API client. WITH_CREDENTIALS must be true so the
// browser sends the HttpOnly session cookie on cross-origin API requests.
// This is set here (not in the generated OpenAPI.ts) so it survives re-generation.
OpenAPI.BASE = BACKEND_URL
OpenAPI.WITH_CREDENTIALS = true

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const deleting = ref(false)
  const deleteError = ref<string | null>(null)

  const isAuthenticated = computed(() => user.value !== null)

  /**
   * Fetch the current user from the backend using the HttpOnly session cookie.
   * Populates `user` on success; clears it if the request fails (not authenticated).
   */
  async function fetchMe() {
    try {
      const result = await DefaultService.getMe()
      if ('user_id' in result) {
        user.value = {
          id: String(result.user_id),
          username: result.user_name,
          email: result.provider_email,
        }
      } else {
        user.value = null
      }
    } catch {
      user.value = null
    }
  }

  /**
   * Redirect the browser to the backend OAuth login endpoint, passing the
   * current origin's /auth/callback as the post-auth redirect target.
   */
  function login(provider: string) {
    const redirectTo = encodeURIComponent(window.location.origin + '/auth/callback')
    window.location.href = `${BACKEND_URL}/auth/login/${provider}?redirect_to=${redirectTo}`
  }

  /**
   * Invalidate the session cookie on the backend, then clear local auth state.
   * Navigation to / is left to the caller (router guards handle it).
   */
  async function logout() {
    try {
      await DefaultService.postAuthLogout()
    } catch (err) {
      // Backend rejection is non-fatal — we still clear local state so the
      // user is signed out in the browser.
      console.warn('logout: backend call failed', err)
    }
    user.value = null
  }

  /**
   * Permanently delete the authenticated user's account via the backend API,
   * then clear all local auth state. Navigation to / is left to the caller.
   */
  async function deleteAccount() {
    deleting.value = true
    deleteError.value = null
    try {
      await DefaultService.deleteAccount()
      user.value = null
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'Failed to delete account. Please try again.'
      deleteError.value = message
      console.error('deleteAccount: backend call failed', err)
    } finally {
      deleting.value = false
    }
  }

  return { user, deleting, deleteError, isAuthenticated, fetchMe, login, logout, deleteAccount }
})
