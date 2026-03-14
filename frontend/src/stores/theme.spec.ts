import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useThemeStore } from './theme'

// jsdom does not implement matchMedia — provide a controllable stub.
function mockMatchMedia(prefersDark: boolean) {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn((query: string) => ({
      matches: query === '(prefers-color-scheme: dark)' ? prefersDark : false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  })
}

beforeEach(() => {
  setActivePinia(createPinia())
  localStorage.clear()
  document.documentElement.classList.remove('dark')
  mockMatchMedia(false)
})

afterEach(() => {
  vi.restoreAllMocks()
})

// ── initialization ────────────────────────────────────────────────────────────

describe('useThemeStore - initialization', () => {
  it('defaults to system mode when nothing is stored', () => {
    const store = useThemeStore()
    expect(store.mode).toBe('system')
  })

  it('applies light class when system prefers light and mode is system', () => {
    mockMatchMedia(false)
    useThemeStore()
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('applies dark class when system prefers dark and mode is system', () => {
    mockMatchMedia(true)
    useThemeStore()
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('restores light mode from localStorage regardless of system preference', () => {
    localStorage.setItem('theme', 'light')
    mockMatchMedia(true)
    const store = useThemeStore()
    expect(store.mode).toBe('light')
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('restores dark mode from localStorage regardless of system preference', () => {
    localStorage.setItem('theme', 'dark')
    mockMatchMedia(false)
    const store = useThemeStore()
    expect(store.mode).toBe('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('restores system mode from localStorage', () => {
    localStorage.setItem('theme', 'system')
    const store = useThemeStore()
    expect(store.mode).toBe('system')
  })
})

// ── setMode ───────────────────────────────────────────────────────────────────

describe('useThemeStore - setMode', () => {
  it('sets light mode and removes dark class', () => {
    mockMatchMedia(true)
    const store = useThemeStore()
    store.setMode('light')
    expect(store.mode).toBe('light')
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('sets dark mode and adds dark class', () => {
    const store = useThemeStore()
    store.setMode('dark')
    expect(store.mode).toBe('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('sets system mode and follows system preference', () => {
    mockMatchMedia(true)
    const store = useThemeStore()
    store.setMode('light') // override first
    store.setMode('system')
    expect(store.mode).toBe('system')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('persists light and dark modes to localStorage', () => {
    const store = useThemeStore()
    store.setMode('dark')
    expect(localStorage.getItem('theme')).toBe('dark')
    store.setMode('light')
    expect(localStorage.getItem('theme')).toBe('light')
  })

  it('removes localStorage entry when switching to system mode', () => {
    localStorage.setItem('theme', 'dark')
    const store = useThemeStore()
    store.setMode('system')
    expect(localStorage.getItem('theme')).toBeNull()
  })
})
