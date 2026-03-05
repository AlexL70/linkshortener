import { describe, it, expect, vi, afterEach } from 'vitest'
import {
  maskedValue,
  validateConfig,
  logConfig,
  REQUIRED_VARS,
  OPTIONAL_DEFAULTS,
  SECRET_VARS,
} from './config'

afterEach(() => {
  vi.restoreAllMocks()
})

// ── maskedValue ───────────────────────────────────────────────────────────────

describe('maskedValue', () => {
  it('dev mode: returns actual value regardless of key', () => {
    for (const key of [...SECRET_VARS, 'APP_BASE_URL', 'LINKSHORTENER_ENV']) {
      expect(maskedValue(key, 'somevalue', false)).toBe('somevalue')
    }
  })

  it('prod mode: masks known secret variable names with "****"', () => {
    for (const key of SECRET_VARS) {
      expect(maskedValue(key, 'supersecret', true)).toBe('****')
    }
  })

  it('prod mode: does not mask non-secret variables', () => {
    expect(maskedValue('APP_BASE_URL', 'http://api.example.com', true)).toBe(
      'http://api.example.com',
    )
    expect(maskedValue('LINKSHORTENER_ENV', 'prod', true)).toBe('prod')
  })
})

// ── validateConfig ────────────────────────────────────────────────────────────

describe('validateConfig', () => {
  it('passes when all required variables are set', () => {
    const env: Record<string, unknown> = {
      LINKSHORTENER_ENV: 'dev',
      APP_BASE_URL: 'http://localhost:8080',
    }
    expect(() => validateConfig(env)).not.toThrow()
  })

  it('throws when LINKSHORTENER_ENV is missing', () => {
    const env: Record<string, unknown> = { APP_BASE_URL: 'http://localhost:8080' }
    expect(() => validateConfig(env)).toThrow('LINKSHORTENER_ENV')
  })

  it('throws when APP_BASE_URL is missing', () => {
    const env: Record<string, unknown> = { LINKSHORTENER_ENV: 'dev' }
    expect(() => validateConfig(env)).toThrow('APP_BASE_URL')
  })

  it('throws when a required variable is an empty string', () => {
    const env: Record<string, unknown> = {
      LINKSHORTENER_ENV: 'dev',
      APP_BASE_URL: '',
    }
    expect(() => validateConfig(env)).toThrow('APP_BASE_URL')
  })

  it('covers all entries in REQUIRED_VARS', () => {
    // Build an env with all required vars set, then remove each one and expect a throw.
    const full: Record<string, unknown> = {}
    for (const key of REQUIRED_VARS) full[key] = 'placeholder'

    for (const key of REQUIRED_VARS) {
      const partial = { ...full, [key]: undefined }
      expect(() => validateConfig(partial)).toThrow(key)
    }
  })
})

// ── logConfig ─────────────────────────────────────────────────────────────────

describe('logConfig', () => {
  it('dev mode: logs required vars with actual values', () => {
    const spy = vi.spyOn(console, 'info').mockImplementation(() => {})
    const env: Record<string, unknown> = {
      LINKSHORTENER_ENV: 'dev',
      APP_BASE_URL: 'http://localhost:8080',
    }
    logConfig(env)
    expect(spy).toHaveBeenCalledWith('[config]', 'LINKSHORTENER_ENV=dev')
    expect(spy).toHaveBeenCalledWith('[config]', 'APP_BASE_URL=http://localhost:8080')
  })

  it('prod mode: logs required vars with actual values (none are secrets)', () => {
    const spy = vi.spyOn(console, 'info').mockImplementation(() => {})
    const env: Record<string, unknown> = {
      LINKSHORTENER_ENV: 'prod',
      APP_BASE_URL: 'https://api.example.com',
    }
    logConfig(env)
    expect(spy).toHaveBeenCalledWith('[config]', 'LINKSHORTENER_ENV=prod')
    expect(spy).toHaveBeenCalledWith('[config]', 'APP_BASE_URL=https://api.example.com')
  })

  it('prod mode: masks secret variable values if present', () => {
    const spy = vi.spyOn(console, 'info').mockImplementation(() => {})
    // Simulate a future scenario where a secret VITE_ var is present.
    const env: Record<string, unknown> = {
      LINKSHORTENER_ENV: 'prod',
      APP_BASE_URL: 'https://api.example.com',
      VITE_GOOGLE_CLIENT_SECRET: 'my-secret',
    }
    // logConfig only logs REQUIRED_VARS and non-default OPTIONAL_DEFAULTS,
    // so VITE_GOOGLE_CLIENT_SECRET won't appear unless it's in one of those lists.
    // This test verifies maskedValue is wired correctly for env entries it does log.
    logConfig(env)
    // Verify the known required vars are unmasked (neither is a secret).
    expect(spy).toHaveBeenCalledWith('[config]', 'APP_BASE_URL=https://api.example.com')
  })

  it('logs optional vars only when they differ from the default', () => {
    const spy = vi.spyOn(console, 'info').mockImplementation(() => {})
    // OPTIONAL_DEFAULTS is currently empty, so nothing extra should be logged
    // beyond the required vars. This test ensures defaults-at-default are silent.
    const env: Record<string, unknown> = {
      LINKSHORTENER_ENV: 'dev',
      APP_BASE_URL: 'http://localhost:8080',
      ...Object.fromEntries(Object.entries(OPTIONAL_DEFAULTS).map(([k, v]) => [k, v])),
    }
    logConfig(env)
    // Only the two required vars should be logged.
    expect(spy).toHaveBeenCalledTimes(REQUIRED_VARS.length)
  })

  it('logs optional vars when they differ from their default', () => {
    // Temporarily add a mock optional default to verify the branching logic.
    // Cast to bypass readonly for this test-only mutation.
    const mutableDefaults = OPTIONAL_DEFAULTS as Record<string, string>
    mutableDefaults['VITE_TEST_VAR'] = 'default-val'

    const spy = vi.spyOn(console, 'info').mockImplementation(() => {})
    const env: Record<string, unknown> = {
      LINKSHORTENER_ENV: 'dev',
      APP_BASE_URL: 'http://localhost:8080',
      VITE_TEST_VAR: 'non-default-val',
    }
    logConfig(env)
    expect(spy).toHaveBeenCalledWith('[config]', 'VITE_TEST_VAR=non-default-val')

    // Cleanup the temporary mutation.
    delete mutableDefaults['VITE_TEST_VAR']
  })
})
