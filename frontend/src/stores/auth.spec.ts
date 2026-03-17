import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from './auth'
import { OpenAPI } from '@/lib/api/core/OpenAPI'
import { DefaultService } from '@/lib/api/services/DefaultService'

// ── helpers ───────────────────────────────────────────────────────────────────

/**
 * Build a minimal well-formed JWT string whose payload contains the given
 * fields. This does NOT produce a cryptographically valid signature — only the
 * structure matters for client-side decoding.
 */
function makeJwt(payload: Record<string, unknown>): string {
  const encode = (obj: object) =>
    btoa(JSON.stringify(obj)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
  return `${encode({ alg: 'HS256', typ: 'JWT' })}.${encode(payload)}.signature`
}

// ── setup / teardown ──────────────────────────────────────────────────────────

beforeEach(() => {
  // Fresh Pinia instance for each test to avoid state bleed.
  setActivePinia(createPinia())
  // Wipe localStorage so token restoration doesn't interfere.
  localStorage.clear()
  // Reset OpenAPI state.
  OpenAPI.TOKEN = undefined
})

afterEach(() => {
  vi.restoreAllMocks()
})

// ── handleCallback ────────────────────────────────────────────────────────────

describe('handleCallback', () => {
  it('decodes a valid JWT and populates token + user including email', () => {
    const jwt = makeJwt({ user_id: 42, user_name: 'alice', email: 'alice@example.com' })
    const store = useAuthStore()

    store.handleCallback(jwt)

    expect(store.token).toBe(jwt)
    expect(store.user).toEqual({ id: '42', username: 'alice', email: 'alice@example.com' })
    expect(store.isAuthenticated).toBe(true)
  })

  it('sets email to empty string when JWT has no email claim', () => {
    const jwt = makeJwt({ user_id: 7, user_name: 'gus' })
    const store = useAuthStore()

    store.handleCallback(jwt)

    expect(store.user?.email).toBe('')
  })

  it('persists the token to localStorage', () => {
    const jwt = makeJwt({ user_id: 1, user_name: 'bob' })
    const store = useAuthStore()

    store.handleCallback(jwt)

    expect(localStorage.getItem('token')).toBe(jwt)
  })

  it('sets OpenAPI.TOKEN so subsequent API calls are authenticated', () => {
    const jwt = makeJwt({ user_id: 1, user_name: 'charlie' })
    const store = useAuthStore()

    store.handleCallback(jwt)

    expect(OpenAPI.TOKEN).toBe(jwt)
  })

  it('does nothing for a malformed JWT', () => {
    const store = useAuthStore()

    store.handleCallback('not-a-jwt')

    expect(store.token).toBeNull()
    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })
})

// ── login ─────────────────────────────────────────────────────────────────────

describe('login', () => {
  it('redirects to the backend login URL with an encoded redirect_to param', () => {
    const locationSpy = vi.spyOn(window, 'location', 'get').mockReturnValue({
      ...window.location,
      href: '',
      origin: 'http://localhost:5173',
    } as Location)

    // Capture the href assignment.
    let assignedHref = ''
    Object.defineProperty(window, 'location', {
      value: { ...window.location, origin: 'http://localhost:5173', set href(v: string) { assignedHref = v } },
      writable: true,
    })

    const store = useAuthStore()
    store.login('google')

    expect(assignedHref).toContain('/auth/login/google')
    expect(assignedHref).toContain('redirect_to=')
    expect(assignedHref).toContain(encodeURIComponent('http://localhost:5173/auth/callback'))

    locationSpy.mockRestore()
  })
})

// ── logout ────────────────────────────────────────────────────────────────────

describe('logout', () => {
  it('calls DefaultService.postAuthLogout() when a token is present', async () => {
    const logoutSpy = vi.spyOn(DefaultService, 'postAuthLogout').mockResolvedValue({} as never)
    const jwt = makeJwt({ user_id: 9, user_name: 'dave' })
    const store = useAuthStore()
    store.handleCallback(jwt)

    await store.logout()

    expect(logoutSpy).toHaveBeenCalledOnce()
  })

  it('clears token, user, localStorage, and OpenAPI.TOKEN after logout', async () => {
    vi.spyOn(DefaultService, 'postAuthLogout').mockResolvedValue({} as never)
    const jwt = makeJwt({ user_id: 9, user_name: 'eve' })
    const store = useAuthStore()
    store.handleCallback(jwt)

    await store.logout()

    expect(store.token).toBeNull()
    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
    expect(localStorage.getItem('token')).toBeNull()
    expect(OpenAPI.TOKEN).toBeUndefined()
  })

  it('still clears local state even if the backend call fails', async () => {
    vi.spyOn(DefaultService, 'postAuthLogout').mockRejectedValue(new Error('network error'))
    vi.spyOn(console, 'warn').mockImplementation(() => {})
    const jwt = makeJwt({ user_id: 9, user_name: 'frank' })
    const store = useAuthStore()
    store.handleCallback(jwt)

    await store.logout()

    expect(store.token).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })

  it('skips the backend call when there is no token', async () => {
    const logoutSpy = vi.spyOn(DefaultService, 'postAuthLogout')
    const store = useAuthStore()

    await store.logout()

    expect(logoutSpy).not.toHaveBeenCalled()
  })
})

// ── deleteAccount ─────────────────────────────────────────────────────────────

describe('deleteAccount', () => {
  it('calls DefaultService.deleteAccount() and clears auth state on success', async () => {
    vi.spyOn(DefaultService, 'deleteAccount').mockResolvedValue({} as never)
    const jwt = makeJwt({ user_id: 5, user_name: 'hank', email: 'hank@example.com' })
    const store = useAuthStore()
    store.handleCallback(jwt)

    await store.deleteAccount()

    expect(store.token).toBeNull()
    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
    expect(localStorage.getItem('token')).toBeNull()
    expect(OpenAPI.TOKEN).toBeUndefined()
    expect(store.deleteError).toBeNull()
    expect(store.deleting).toBe(false)
  })

  it('sets deleteError and keeps auth state when backend returns an error', async () => {
    vi.spyOn(DefaultService, 'deleteAccount').mockRejectedValue(new Error('forbidden'))
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const jwt = makeJwt({ user_id: 6, user_name: 'ivy', email: 'ivy@example.com' })
    const store = useAuthStore()
    store.handleCallback(jwt)

    await store.deleteAccount()

    expect(store.deleteError).toBe('forbidden')
    expect(store.token).not.toBeNull()
    expect(store.isAuthenticated).toBe(true)
    expect(store.deleting).toBe(false)
  })

  it('sets deleting to true during the call and false after', async () => {
    let resolveFn!: () => void
    vi.spyOn(DefaultService, 'deleteAccount').mockImplementation(
      () => new Promise<never>((resolve) => { resolveFn = resolve as () => void }),
    )
    const jwt = makeJwt({ user_id: 7, user_name: 'jan', email: 'jan@example.com' })
    const store = useAuthStore()
    store.handleCallback(jwt)

    const promise = store.deleteAccount()
    expect(store.deleting).toBe(true)

    resolveFn()
    await promise
    expect(store.deleting).toBe(false)
  })
})
