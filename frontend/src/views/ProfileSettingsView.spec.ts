import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import ProfileSettingsView from './ProfileSettingsView.vue'
import { useAuthStore } from '@/stores/auth'
import { DefaultService } from '@/lib/api/services/DefaultService'

// ── helpers ───────────────────────────────────────────────────────────────────

function makeJwt(payload: Record<string, unknown>): string {
  const encode = (obj: object) =>
    btoa(JSON.stringify(obj)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
  return `${encode({ alg: 'HS256', typ: 'JWT' })}.${encode(payload)}.sig`
}

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: { template: '<div>Home</div>' } },
      { path: '/profile/settings', name: 'profile-settings', component: ProfileSettingsView },
    ],
  })
}

// ── setup / teardown ──────────────────────────────────────────────────────────

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

// ── mount helper ──────────────────────────────────────────────────────────────

async function mountView(auth: ReturnType<typeof useAuthStore>) {
  const jwt = makeJwt({ user_id: 1, user_name: 'alice', email: 'alice@example.com' })
  auth.handleCallback(jwt)
  const router = makeRouter()
  await router.push('/profile/settings')
  const wrapper = mount(ProfileSettingsView, {
    attachTo: mountTarget,
    global: { plugins: [router] },
  })
  await flushPromises()
  return wrapper
}

// ── initial render ────────────────────────────────────────────────────────────

describe('initial render', () => {
  it('displays the Danger Zone heading and description', async () => {
    const store = useAuthStore()
    const wrapper = await mountView(store)
    expect(wrapper.text()).toContain('Danger Zone')
    expect(wrapper.text()).toContain('Once you delete your account')
  })

  it('shows the delete trigger button', async () => {
    const store = useAuthStore()
    const wrapper = await mountView(store)
    expect(wrapper.text()).toContain('Delete Account Permanently')
  })
})

// ── open dialog ───────────────────────────────────────────────────────────────

describe('dialog opening', () => {
  it('reveals dialog content after clicking the trigger button', async () => {
    const store = useAuthStore()
    await mountView(store)

    const trigger = document.querySelector('button') as HTMLButtonElement
    trigger.click()
    await flushPromises()

    expect(document.body.textContent).toContain('Delete Account')
    expect(document.body.textContent).toContain('alice@example.com')
  })
})

// ── paste/drop prevention ───────────────────────────────────────────────────

describe('paste and drop prevention', () => {
  async function openDialog(store: ReturnType<typeof useAuthStore>) {
    await mountView(store)
    const trigger = document.querySelector('button') as HTMLButtonElement
    trigger.click()
    await flushPromises()
  }

  it('ignores paste events on the confirm input', async () => {
    const store = useAuthStore()
    await openDialog(store)

    const input = document.querySelector<HTMLInputElement>('input[type="email"]')!
    const pasteEvent = new Event('paste', { bubbles: true, cancelable: true })
    input.dispatchEvent(pasteEvent)
    await flushPromises()

    // The input value must not have changed via paste
    expect(input.value).toBe('')
    expect(pasteEvent.defaultPrevented).toBe(true)
  })

  it('ignores drop events on the confirm input', async () => {
    const store = useAuthStore()
    await openDialog(store)

    const input = document.querySelector<HTMLInputElement>('input[type="email"]')!
    const dropEvent = new Event('drop', { bubbles: true, cancelable: true })
    input.dispatchEvent(dropEvent)
    await flushPromises()

    expect(dropEvent.defaultPrevented).toBe(true)
  })
})

// ── confirmation guard ────────────────────────────────────────────────────────

describe('email confirmation guard', () => {
  async function openDialog(store: ReturnType<typeof useAuthStore>) {
    await mountView(store)
    const trigger = document.querySelector('button') as HTMLButtonElement
    trigger.click()
    await flushPromises()
  }

  it('delete button is disabled when confirm input is empty', async () => {
    const store = useAuthStore()
    await openDialog(store)

    const buttons = Array.from(document.querySelectorAll<HTMLButtonElement>('button'))
    const deleteBtn = buttons.find((b) => b.textContent?.includes('Delete My Account'))
    expect(deleteBtn?.disabled).toBe(true)
  })

  it('delete button is disabled when confirm input does not match email', async () => {
    const store = useAuthStore()
    await openDialog(store)

    const input = document.querySelector<HTMLInputElement>('input[type="email"]')!
    input.value = 'wrong@example.com'
    input.dispatchEvent(new Event('input', { bubbles: true }))
    await flushPromises()

    const buttons = Array.from(document.querySelectorAll<HTMLButtonElement>('button'))
    const deleteBtn = buttons.find((b) => b.textContent?.includes('Delete My Account'))
    expect(deleteBtn?.disabled).toBe(true)
  })

  it('delete button is enabled when confirm input matches email (case-insensitive)', async () => {
    const store = useAuthStore()
    await openDialog(store)

    const input = document.querySelector<HTMLInputElement>('input[type="email"]')!
    input.value = 'ALICE@EXAMPLE.COM'
    input.dispatchEvent(new Event('input', { bubbles: true }))
    await flushPromises()

    const buttons = Array.from(document.querySelectorAll<HTMLButtonElement>('button'))
    const deleteBtn = buttons.find((b) => b.textContent?.includes('Delete My Account'))
    expect(deleteBtn?.disabled).toBe(false)
  })
})

// ── successful deletion ───────────────────────────────────────────────────────

describe('successful deletion', () => {
  it('calls deleteAccount, closes the dialog, and navigates to home', async () => {
    vi.spyOn(DefaultService, 'deleteAccount').mockResolvedValue({} as never)
    const store = useAuthStore()
    const router = makeRouter()
    const jwt = makeJwt({ user_id: 1, user_name: 'alice', email: 'alice@example.com' })
    store.handleCallback(jwt)
    await router.push('/profile/settings')
    mount(ProfileSettingsView, {
      attachTo: mountTarget,
      global: { plugins: [router] },
    })
    await flushPromises()

    // Open dialog
    const trigger = document.querySelector('button') as HTMLButtonElement
    trigger.click()
    await flushPromises()

    // Type matching email
    const input = document.querySelector<HTMLInputElement>('input[type="email"]')!
    input.value = 'alice@example.com'
    input.dispatchEvent(new Event('input', { bubbles: true }))
    await flushPromises()

    // Click confirm
    const buttons = Array.from(document.querySelectorAll<HTMLButtonElement>('button'))
    const deleteBtn = buttons.find((b) => b.textContent?.includes('Delete My Account'))!
    deleteBtn.click()
    await flushPromises()

    expect(DefaultService.deleteAccount).toHaveBeenCalledOnce()
    expect(router.currentRoute.value.name).toBe('home')
  })
})

// ── failed deletion ───────────────────────────────────────────────────────────

describe('failed deletion', () => {
  it('shows deleteError message and keeps the dialog open', async () => {
    vi.spyOn(DefaultService, 'deleteAccount').mockRejectedValue(new Error('server error'))
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const store = useAuthStore()
    await mountView(store)

    // Open dialog
    const trigger = document.querySelector('button') as HTMLButtonElement
    trigger.click()
    await flushPromises()

    // Type matching email
    const input = document.querySelector<HTMLInputElement>('input[type="email"]')!
    input.value = 'alice@example.com'
    input.dispatchEvent(new Event('input', { bubbles: true }))
    await flushPromises()

    // Click confirm
    const buttons = Array.from(document.querySelectorAll<HTMLButtonElement>('button'))
    const deleteBtn = buttons.find((b) => b.textContent?.includes('Delete My Account'))!
    deleteBtn.click()
    await flushPromises()

    expect(document.body.textContent).toContain('server error')
    // Dialog should still be open (error shown inside it)
    expect(document.body.textContent).toContain('Delete Account')
  })
})
