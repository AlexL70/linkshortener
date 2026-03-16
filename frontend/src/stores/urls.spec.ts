import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useUrlsStore } from './urls'
import { DefaultService } from '@/lib/api/services/DefaultService'
import type { UrlItem } from '@/lib/api/models/UrlItem'
import type { CreateUrlResponseBody } from '@/lib/api/models/CreateUrlResponseBody'

// ── helpers ───────────────────────────────────────────────────────────────────

function makeUrlItem(overrides: Partial<UrlItem> = {}): UrlItem {
  return {
    id: 1,
    shortcode: 'abc123',
    long_url: 'https://example.com',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    ...overrides,
  }
}

function makeListResponse(items: UrlItem[], page = 1, pageSize = 20, total?: number) {
  return {
    items,
    total: total ?? items.length,
    page,
    page_size: pageSize,
  }
}

// ── setup / teardown ──────────────────────────────────────────────────────────

beforeEach(() => {
  setActivePinia(createPinia())
})

afterEach(() => {
  vi.restoreAllMocks()
})

// ── initial state ─────────────────────────────────────────────────────────────

describe('initial state', () => {
  it('starts with empty items and sensible defaults', () => {
    const store = useUrlsStore()

    expect(store.items).toEqual([])
    expect(store.total).toBe(0)
    expect(store.page).toBe(1)
    expect(store.pageSize).toBe(0)
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })
})

// ── fetchUrls — success ───────────────────────────────────────────────────────

describe('fetchUrls — success', () => {
  it('populates items and pagination from a successful API response', async () => {
    const url1 = makeUrlItem({ id: 1, shortcode: 'aaa111' })
    const url2 = makeUrlItem({ id: 2, shortcode: 'bbb222' })
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      makeListResponse([url1, url2], 1, 20, 2) as never,
    )

    const store = useUrlsStore()
    await store.fetchUrls()

    expect(store.items).toEqual([url1, url2])
    expect(store.total).toBe(2)
    expect(store.page).toBe(1)
    expect(store.pageSize).toBe(20)
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('passes page and pageSize query parameters to the API', async () => {
    const spy = vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      makeListResponse([], 3, 10, 0) as never,
    )

    const store = useUrlsStore()
    await store.fetchUrls(3, 10)

    expect(spy).toHaveBeenCalledWith({ page: 3, pageSize: 10 })
  })

  it('omits pageSize when called with no explicit page size, deferring to server default', async () => {
    const spy = vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      makeListResponse([], 1, 5, 5) as never,
    )

    const store = useUrlsStore()
    await store.fetchUrls()

    expect(spy).toHaveBeenCalledWith({ page: 1, pageSize: undefined })
  })

  it('treats null items as an empty array', async () => {
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      { items: null, total: 0, page: 1, page_size: 20 } as never,
    )

    const store = useUrlsStore()
    await store.fetchUrls()

    expect(store.items).toEqual([])
  })

  it('sets loading to true during the request and false after', async () => {
    let resolveFn!: (v: unknown) => void
    vi.spyOn(DefaultService, 'listUserUrls').mockReturnValue(
      new Promise((resolve) => { resolveFn = resolve }) as never,
    )

    const store = useUrlsStore()
    const fetchPromise = store.fetchUrls()

    expect(store.loading).toBe(true)
    resolveFn(makeListResponse([]))
    await fetchPromise
    expect(store.loading).toBe(false)
  })
})

// ── fetchUrls — API error response ────────────────────────────────────────────

describe('fetchUrls — API error response', () => {
  it('sets error when the API returns an ErrorModel with a title', async () => {
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      { status: 401, title: 'Unauthorized' } as never,
    )

    const store = useUrlsStore()
    await store.fetchUrls()

    expect(store.error).toBe('Unauthorized')
    expect(store.items).toEqual([])
  })

  it('falls back to a generic message when ErrorModel has no title', async () => {
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      { status: 500 } as never,
    )

    const store = useUrlsStore()
    await store.fetchUrls()

    expect(store.error).toBe('Failed to load URLs')
  })
})

// ── fetchUrls — network error ─────────────────────────────────────────────────

describe('fetchUrls — network error', () => {
  it('sets a generic error message when the request throws', async () => {
    vi.spyOn(DefaultService, 'listUserUrls').mockRejectedValue(new Error('Network Error'))
    vi.spyOn(console, 'error').mockImplementation(() => {})

    const store = useUrlsStore()
    await store.fetchUrls()

    expect(store.error).toBe('An unexpected error occurred while loading your URLs.')
    expect(store.loading).toBe(false)
  })

  it('logs the error to the console', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    vi.spyOn(DefaultService, 'listUserUrls').mockRejectedValue(new Error('timeout'))

    const store = useUrlsStore()
    await store.fetchUrls()

    expect(consoleSpy).toHaveBeenCalledWith('fetchUrls: request failed', expect.any(Error))
  })
})

// ── reset ─────────────────────────────────────────────────────────────────────

describe('reset', () => {
  it('clears all state back to defaults', async () => {
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      makeListResponse([makeUrlItem()], 2, 10, 1) as never,
    )

    const store = useUrlsStore()
    await store.fetchUrls(2, 10)

    store.reset()

    expect(store.items).toEqual([])
    expect(store.total).toBe(0)
    expect(store.page).toBe(1)
    expect(store.pageSize).toBe(0)
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })
})

// ── createUrl helpers ─────────────────────────────────────────────────────────

function makeCreateUrlResponse(overrides: Partial<CreateUrlResponseBody> = {}): CreateUrlResponseBody {
  return {
    id: 99,
    shortcode: 'abc123',
    long_url: 'https://example.com',
    short_url: 'https://s.example.com/r/abc123',
    created_at: '2024-06-01T00:00:00Z',
    ...overrides,
  }
}

// ── createUrl — success ───────────────────────────────────────────────────────

describe('createUrl — success', () => {
  it('returns the created URL and refreshes the list', async () => {
    const created = makeCreateUrlResponse()
    vi.spyOn(DefaultService, 'createUrl').mockResolvedValue(created as never)
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      makeListResponse([makeUrlItem({ id: 99, shortcode: 'abc123' })]) as never,
    )

    const store = useUrlsStore()
    const result = await store.createUrl({ longUrl: 'https://example.com' })

    expect(result).toEqual(created)
    expect(store.creating).toBe(false)
    expect(store.createError).toBeNull()
    expect(store.items).toHaveLength(1)
  })

  it('passes all parameters to the API', async () => {
    const spy = vi.spyOn(DefaultService, 'createUrl').mockResolvedValue(
      makeCreateUrlResponse() as never,
    )
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      makeListResponse([]) as never,
    )

    const store = useUrlsStore()
    await store.createUrl({ longUrl: 'https://example.com', shortcode: 'abc123', expiresAt: '2025-01-01T00:00:00Z' })

    expect(spy).toHaveBeenCalledWith({
      requestBody: {
        long_url: 'https://example.com',
        shortcode: 'abc123',
        expires_at: '2025-01-01T00:00:00Z',
      },
    })
  })

  it('sets creating to true during the request and false after', async () => {
    let resolveFn!: (v: unknown) => void
    vi.spyOn(DefaultService, 'createUrl').mockReturnValue(
      new Promise((resolve) => { resolveFn = resolve }) as never,
    )

    const store = useUrlsStore()
    const createPromise = store.createUrl({ longUrl: 'https://example.com' })

    expect(store.creating).toBe(true)
    resolveFn(makeCreateUrlResponse())
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(makeListResponse([]) as never)
    await createPromise
    expect(store.creating).toBe(false)
  })
})

// ── createUrl — API error response ────────────────────────────────────────────

describe('createUrl — API error response', () => {
  it('sets createError and returns null when the API returns an ErrorModel', async () => {
    vi.spyOn(DefaultService, 'createUrl').mockResolvedValue(
      { status: 400, title: 'Invalid URL' } as never,
    )

    const store = useUrlsStore()
    const result = await store.createUrl({ longUrl: 'not-a-url' })

    expect(result).toBeNull()
    expect(store.createError).toBe('Invalid URL')
    expect(store.creating).toBe(false)
  })

  it('falls back to a generic message when ErrorModel has no title', async () => {
    vi.spyOn(DefaultService, 'createUrl').mockResolvedValue(
      { status: 500 } as never,
    )

    const store = useUrlsStore()
    await store.createUrl({ longUrl: 'https://example.com' })

    expect(store.createError).toBe('Failed to create URL')
  })
})

// ── createUrl — network error ─────────────────────────────────────────────────

describe('createUrl — network error', () => {
  it('sets a generic createError message when the request throws', async () => {
    vi.spyOn(DefaultService, 'createUrl').mockRejectedValue(new Error('Network Error'))
    vi.spyOn(console, 'error').mockImplementation(() => {})

    const store = useUrlsStore()
    const result = await store.createUrl({ longUrl: 'https://example.com' })

    expect(result).toBeNull()
    expect(store.createError).toBe('An unexpected error occurred while creating your URL.')
    expect(store.creating).toBe(false)
  })
})

// ── clearCreateError ──────────────────────────────────────────────────────────

describe('clearCreateError', () => {
  it('resets createError to null', async () => {
    vi.spyOn(DefaultService, 'createUrl').mockResolvedValue({ status: 400, title: 'Bad' } as never)

    const store = useUrlsStore()
    await store.createUrl({ longUrl: 'bad' })
    expect(store.createError).not.toBeNull()

    store.clearCreateError()
    expect(store.createError).toBeNull()
  })
})

// ── updateUrl helpers ─────────────────────────────────────────────────────────

function makeUpdateUrlResponse() {
  return {
    id: 99,
    shortcode: 'abc123',
    long_url: 'https://updated.com',
    created_at: '2024-06-01T00:00:00Z',
    updated_at: '2024-06-02T00:00:00Z',
  }
}

// ── updateUrl — success ───────────────────────────────────────────────────────

describe('updateUrl — success', () => {
  it('returns the updated URL and refreshes the list', async () => {
    const updated = makeUpdateUrlResponse()
    vi.spyOn(DefaultService, 'updateUrl').mockResolvedValue(updated as never)
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      makeListResponse([makeUrlItem({ id: 99, long_url: 'https://updated.com' })]) as never,
    )

    const store = useUrlsStore()
    const result = await store.updateUrl({ id: 99, longUrl: 'https://updated.com' })

    expect(result).toEqual(updated)
    expect(store.updating).toBe(false)
    expect(store.updateError).toBeNull()
    expect(store.items).toHaveLength(1)
  })

  it('passes all parameters to the API', async () => {
    const spy = vi.spyOn(DefaultService, 'updateUrl').mockResolvedValue(
      makeUpdateUrlResponse() as never,
    )
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(
      makeListResponse([]) as never,
    )

    const store = useUrlsStore()
    await store.updateUrl({ id: 5, longUrl: 'https://updated.com', shortcode: 'new-sc', expiresAt: '2025-01-01T00:00:00Z' })

    expect(spy).toHaveBeenCalledWith({
      id: 5,
      requestBody: {
        long_url: 'https://updated.com',
        shortcode: 'new-sc',
        expires_at: '2025-01-01T00:00:00Z',
      },
    })
  })

  it('sets updating to true during the request and false after', async () => {
    let resolveFn!: (v: unknown) => void
    vi.spyOn(DefaultService, 'updateUrl').mockReturnValue(
      new Promise((resolve) => { resolveFn = resolve }) as never,
    )

    const store = useUrlsStore()
    const updatePromise = store.updateUrl({ id: 1, longUrl: 'https://example.com' })

    expect(store.updating).toBe(true)
    resolveFn(makeUpdateUrlResponse())
    vi.spyOn(DefaultService, 'listUserUrls').mockResolvedValue(makeListResponse([]) as never)
    await updatePromise
    expect(store.updating).toBe(false)
  })
})

// ── updateUrl — API error response ────────────────────────────────────────────

describe('updateUrl — API error response', () => {
  it('sets updateError and returns null when the API returns an ErrorModel', async () => {
    vi.spyOn(DefaultService, 'updateUrl').mockResolvedValue(
      { status: 404, title: 'Not Found' } as never,
    )

    const store = useUrlsStore()
    const result = await store.updateUrl({ id: 99, longUrl: 'https://example.com' })

    expect(result).toBeNull()
    expect(store.updateError).toBe('Not Found')
    expect(store.updating).toBe(false)
  })

  it('falls back to a generic message when ErrorModel has no title', async () => {
    vi.spyOn(DefaultService, 'updateUrl').mockResolvedValue(
      { status: 500 } as never,
    )

    const store = useUrlsStore()
    await store.updateUrl({ id: 99, longUrl: 'https://example.com' })

    expect(store.updateError).toBe('Failed to update URL')
  })
})

// ── updateUrl — network error ─────────────────────────────────────────────────

describe('updateUrl — network error', () => {
  it('sets a generic updateError message when the request throws', async () => {
    vi.spyOn(DefaultService, 'updateUrl').mockRejectedValue(new Error('Network Error'))
    vi.spyOn(console, 'error').mockImplementation(() => {})

    const store = useUrlsStore()
    const result = await store.updateUrl({ id: 1, longUrl: 'https://example.com' })

    expect(result).toBeNull()
    expect(store.updateError).toBe('An unexpected error occurred while updating your URL.')
    expect(store.updating).toBe(false)
  })
})

// ── clearUpdateError ──────────────────────────────────────────────────────────

describe('clearUpdateError', () => {
  it('resets updateError to null', async () => {
    vi.spyOn(DefaultService, 'updateUrl').mockResolvedValue({ status: 404, title: 'Not Found' } as never)

    const store = useUrlsStore()
    await store.updateUrl({ id: 1, longUrl: 'bad' })
    expect(store.updateError).not.toBeNull()

    store.clearUpdateError()
    expect(store.updateError).toBeNull()
  })
})
