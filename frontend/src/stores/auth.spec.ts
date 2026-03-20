import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from './auth'
import { DefaultService } from '@/lib/api/services/DefaultService'
import type { MeBody } from '@/lib/api/models/MeBody'

// ── helpers ───────────────────────────────────────────────────────────────────

function makeMeBody(overrides: Partial<MeBody> = {}): MeBody {
  return {
    user_id: 1,
    user_name: 'alice',
    provider_email: 'alice@example.com',
    ...overrides,
  }
}

// ── setup / teardown ──────────────────────────────────────────────────────────

beforeEach(() => {
  // Fresh Pinia instance for each test to avoid state bleed.
  setActivePinia(createPinia())
})

afterEach(() => {
  vi.restoreAllMocks()
})

// ── fetchMe ───────────────────────────────────────────────────────────────────

describe('fetchMe', () => {
  it('populates user when the backend returns user data', async () => {
    vi.spyOn(DefaultService, 'getMe').mockResolvedValue(makeMeBody())
    const store = useAuthStore()

    await store.fetchMe()

    expect(store.user).toEqual({ id: '1', username: 'alice', email: 'alice@example.com' })
    expect(store.isAuthenticated).toBe(true)
  })

  it('converts user_id to a string', async () => {
    vi.spyOn(DefaultService, 'getMe').mockResolvedValue(makeMeBody({ user_id: 42 }))
    const store = useAuthStore()

    await store.fetchMe()

    expect(store.user?.id).toBe('42')
  })

  it('clears user when the request fails (not authenticated)', async () => {
    vi.spyOn(DefaultService, 'getMe').mockRejectedValue(new Error('401'))
    const store = useAuthStore()

    await store.fetchMe()

    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })

  it('clears user when the response lacks user_id (ErrorModel shape)', async () => {
    // Simulate a response that looks like an ErrorModel (no user_id).
    vi.spyOn(DefaultService, 'getMe').mockResolvedValue({ status: 401, title: 'Unauthorized' } as never)
    const store = useAuthStore()

    await store.fetchMe()

    expect(store.user).toBeNull()
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
    expect(assignedHref).toContain(encodeURIComponent('http://localhost:5173/callback'))

    locationSpy.mockRestore()
  })
})

// ── logout ────────────────────────────────────────────────────────────────────

describe('logout', () => {
  it('always calls DefaultService.postAuthLogout()', async () => {
    const logoutSpy = vi.spyOn(DefaultService, 'postAuthLogout').mockResolvedValue({} as never)
    const store = useAuthStore()

    await store.logout()

    expect(logoutSpy).toHaveBeenCalledOnce()
  })

  it('clears user after logout', async () => {
    vi.spyOn(DefaultService, 'getMe').mockResolvedValue(makeMeBody())
    vi.spyOn(DefaultService, 'postAuthLogout').mockResolvedValue({} as never)
    const store = useAuthStore()
    await store.fetchMe()
    expect(store.isAuthenticated).toBe(true)

    await store.logout()

    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })

  it('still clears local state even if the backend call fails', async () => {
    vi.spyOn(DefaultService, 'getMe').mockResolvedValue(makeMeBody())
    vi.spyOn(DefaultService, 'postAuthLogout').mockRejectedValue(new Error('network error'))
    vi.spyOn(console, 'warn').mockImplementation(() => {})
    const store = useAuthStore()
    await store.fetchMe()

    await store.logout()

    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })
})

// ── deleteAccount ─────────────────────────────────────────────────────────────

describe('deleteAccount', () => {
  it('calls DefaultService.deleteAccount() and clears auth state on success', async () => {
    vi.spyOn(DefaultService, 'getMe').mockResolvedValue(makeMeBody({ user_id: 5, user_name: 'hank', provider_email: 'hank@example.com' }))
    vi.spyOn(DefaultService, 'deleteAccount').mockResolvedValue({} as never)
    const store = useAuthStore()
    await store.fetchMe()

    await store.deleteAccount()

    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
    expect(store.deleteError).toBeNull()
    expect(store.deleting).toBe(false)
  })

  it('sets deleteError and keeps auth state when backend returns an error', async () => {
    vi.spyOn(DefaultService, 'getMe').mockResolvedValue(makeMeBody({ user_id: 6, user_name: 'ivy', provider_email: 'ivy@example.com' }))
    vi.spyOn(DefaultService, 'deleteAccount').mockRejectedValue(new Error('forbidden'))
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const store = useAuthStore()
    await store.fetchMe()

    await store.deleteAccount()

    expect(store.deleteError).toBe('forbidden')
    expect(store.user).not.toBeNull()
    expect(store.isAuthenticated).toBe(true)
    expect(store.deleting).toBe(false)
  })

  it('sets deleting to true during the call and false after', async () => {
    vi.spyOn(DefaultService, 'getMe').mockResolvedValue(makeMeBody({ user_id: 7, user_name: 'jan', provider_email: 'jan@example.com' }))
    let resolveFn!: () => void
    vi.spyOn(DefaultService, 'deleteAccount').mockImplementation(
      () => new Promise<never>((resolve) => { resolveFn = resolve as () => void }) as never,
    )
    const store = useAuthStore()
    await store.fetchMe()

    const promise = store.deleteAccount()
    expect(store.deleting).toBe(true)

    resolveFn()
    await promise
    expect(store.deleting).toBe(false)
  })
})
