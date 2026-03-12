import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { OpenAPI } from '@/lib/api/core/OpenAPI'
import { DefaultService } from '@/lib/api/services/DefaultService'

interface User {
  id: string
  username: string
}

const BACKEND_URL = import.meta.env.APP_BASE_URL as string ?? ''

// Configure the generated API client with the backend base URL.
OpenAPI.BASE = BACKEND_URL

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('token', newToken)
    // Keep the generated API client in sync so authenticated calls include the
    // Bearer token automatically.
    OpenAPI.TOKEN = newToken
  }

  function setUser(newUser: User) {
    user.value = newUser
  }

  /**
   * Decode the JWT payload (base64url) and populate the store state.
   * Called after a successful OAuth callback or registration.
   */
  function handleCallback(jwtToken: string) {
    const parts = jwtToken.split('.')
    if (parts.length !== 3) return

    try {
      const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')))
      setToken(jwtToken)
      setUser({ id: String(payload.user_id ?? ''), username: payload.user_name ?? '' })
    } catch {
      // Malformed token — silently ignore; caller should check isAuthenticated.
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
   * Invalidate the JWT on the backend, then clear all local auth state.
   * Navigation to / is left to the caller (router guards handle it).
   */
  async function logout() {
    if (token.value) {
      OpenAPI.TOKEN = token.value
      try {
        await DefaultService.logout()
      } catch (err) {
        // Backend rejection (e.g. already-expired token) is non-fatal — we
        // still clear local state so the user is signed out in the browser.
        console.warn('logout: backend call failed', err)
      }
    }
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    OpenAPI.TOKEN = undefined
  }

  // Restore the API client token if we already have one in localStorage.
  if (token.value) {
    OpenAPI.TOKEN = token.value
  }

  return { token, user, isAuthenticated, setToken, setUser, handleCallback, login, logout }
})
