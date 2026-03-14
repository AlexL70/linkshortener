import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ThemeMode = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'theme'

function getSystemMedia() {
  return window.matchMedia('(prefers-color-scheme: dark)')
}

function resolvedDark(mode: ThemeMode): boolean {
  return mode === 'system' ? getSystemMedia().matches : mode === 'dark'
}

function applyMode(mode: ThemeMode) {
  document.documentElement.classList.toggle('dark', resolvedDark(mode))
}

export const useThemeStore = defineStore('theme', () => {
  const stored = localStorage.getItem(STORAGE_KEY) as ThemeMode | null
  const initial: ThemeMode = stored ?? 'system'

  const mode = ref<ThemeMode>(initial)
  applyMode(initial)

  // Keep the DOM in sync when the user is in 'system' mode and the OS theme changes.
  getSystemMedia().addEventListener('change', () => {
    if (mode.value === 'system') applyMode('system')
  })

  function setMode(newMode: ThemeMode) {
    mode.value = newMode
    if (newMode === 'system') {
      localStorage.removeItem(STORAGE_KEY)
    } else {
      localStorage.setItem(STORAGE_KEY, newMode)
    }
    applyMode(newMode)
  }

  return { mode, setMode }
})
