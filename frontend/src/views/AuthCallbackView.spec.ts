import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import AuthCallbackView from './AuthCallbackView.vue'
import { useAuthStore } from '@/stores/auth'
import { DefaultService } from '@/lib/api/services/DefaultService'
import { ApiError } from '@/lib/api/core/ApiError'

// ── router stub ───────────────────────────────────────────────────────────────

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: { template: '<div>Home</div>' } },
      { path: '/dashboard', name: 'dashboard', component: { template: '<div>Dashboard</div>' } },
      { path: '/auth/callback', name: 'auth-callback', component: AuthCallbackView },
    ],
  })
}

// ── helpers ───────────────────────────────────────────────────────────────────

function makeJwt(payload: Record<string, unknown>): string {
  const encode = (obj: object) =>
    btoa(JSON.stringify(obj)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
  return `${encode({ alg: 'HS256', typ: 'JWT' })}.${encode(payload)}.sig`
}

function setHash(hash: string) {
  Object.defineProperty(window, 'location', {
    value: { ...window.location, hash },
    writable: true,
    configurable: true,
  })
}

async function mountView(router: ReturnType<typeof makeRouter>) {
  const wrapper = mount(AuthCallbackView, {
    attachTo: mountTarget,
    global: {
      plugins: [createPinia(), router],
    },
  })
  await flushPromises()
  return wrapper
}

// ── setup / teardown ──────────────────────────────────────────────────────────

// Attach mount target so portals (DialogPortal → Teleport) render into <body>.
let mountTarget: HTMLDivElement

beforeEach(() => {
  setActivePinia(createPinia())
  localStorage.clear()
  mountTarget = document.createElement('div')
  document.body.appendChild(mountTarget)
})

afterEach(() => {
  document.body.innerHTML = ''
  vi.restoreAllMocks()
})

// ── token in hash ─────────────────────────────────────────────────────────────

describe('when #token= is present', () => {
  it('calls handleCallback with the token', async () => {
    const jwt = makeJwt({ user_id: 1, user_name: 'alice' })
    setHash(`#token=${jwt}`)
    const router = makeRouter()
    await router.push('/auth/callback')

    await mount(AuthCallbackView, { global: { plugins: [createPinia(), router] } })
    await flushPromises()

    const store = useAuthStore()
    expect(store.token).toBe(jwt)
    expect(store.isAuthenticated).toBe(true)
  })

  it('redirects to /dashboard', async () => {
    const jwt = makeJwt({ user_id: 1, user_name: 'alice' })
    setHash(`#token=${jwt}`)
    const router = makeRouter()
    await router.push('/auth/callback')
    vi.spyOn(router, 'replace')

    await mount(AuthCallbackView, { global: { plugins: [createPinia(), router] } })
    await flushPromises()

    expect(router.replace).toHaveBeenCalledWith('/dashboard')
  })
})

// ── pre_registration_token in hash ────────────────────────────────────────────

describe('when #pre_registration_token= is present', () => {
  it('shows the registration dialog with pre-filled username', async () => {
    setHash('#pre_registration_token=abc&suggested_user_name=newuser')
    const router = makeRouter()
    await router.push('/auth/callback')

    await mountView(router)

    // DialogPortal renders into document.body via Teleport — check there.
    expect(document.body.textContent).toContain('Create your account')
    const input = document.querySelector<HTMLInputElement>('input')
    expect(input).not.toBeNull()
    expect(input!.value).toBe('newuser')
  })

  it('redirects to /dashboard after successful registration', async () => {
    setHash('#pre_registration_token=tok123&suggested_user_name=jane')
    const jwt = makeJwt({ user_id: 2, user_name: 'jane' })
    vi.spyOn(DefaultService, 'registerUser').mockResolvedValue({ token: jwt })
    const router = makeRouter()
    await router.push('/auth/callback')
    vi.spyOn(router, 'replace')

    await mountView(router)
    const form = document.querySelector('form')!
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }))
    await flushPromises()

    expect(router.replace).toHaveBeenCalledWith('/dashboard')
    const store = useAuthStore()
    expect(store.isAuthenticated).toBe(true)
  })

  it('shows a conflict error message on a 409 response', async () => {
    setHash('#pre_registration_token=tok123&suggested_user_name=taken')
    const apiError = new ApiError(
      { method: 'POST', url: '/auth/register' },
      { url: '/auth/register', ok: false, status: 409, statusText: 'Conflict', body: {} },
      'Conflict',
    )
    vi.spyOn(DefaultService, 'registerUser').mockRejectedValue(apiError)
    const router = makeRouter()
    await router.push('/auth/callback')

    await mountView(router)
    const form = document.querySelector('form')!
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }))
    await flushPromises()

    expect(document.body.textContent).toContain('Username already taken')
  })

  it('shows a generic error message on other failures', async () => {
    setHash('#pre_registration_token=tok123&suggested_user_name=err')
    vi.spyOn(DefaultService, 'registerUser').mockRejectedValue(new Error('Network error'))
    const router = makeRouter()
    await router.push('/auth/callback')

    await mountView(router)
    const form = document.querySelector('form')!
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }))
    await flushPromises()

    expect(document.body.textContent).toContain('Registration failed')
  })
})

// ── error in hash ─────────────────────────────────────────────────────────────

describe('when #error= is present', () => {
  it('shows the error message', async () => {
    setHash('#error=authentication_failed')
    const router = makeRouter()
    await router.push('/auth/callback')

    const wrapper = await mountView(router)

    expect(wrapper.text()).toContain('authentication_failed')
  })

  it('provides a "Go home" button that navigates to /', async () => {
    setHash('#error=authentication_failed')
    const router = makeRouter()
    await router.push('/auth/callback')
    vi.spyOn(router, 'replace')

    const wrapper = await mountView(router)
    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(router.replace).toHaveBeenCalledWith('/')
  })
})
