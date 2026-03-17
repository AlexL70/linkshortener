import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import DashboardView from './DashboardView.vue'
import { useUrlsStore } from '@/stores/urls'
import type { UrlItem } from '@/lib/api/models/UrlItem'

// ── mocks ─────────────────────────────────────────────────────────────────────

const mockCopy = vi.fn()

vi.mock('@vueuse/core', async (importOriginal) => {
  const original = await importOriginal<typeof import('@vueuse/core')>()
  return {
    ...original,
    useClipboard: () => ({ copy: mockCopy }),
  }
})

vi.mock('@/stores/urls', () => ({
  useUrlsStore: vi.fn(),
}))

// ── test data ─────────────────────────────────────────────────────────────────

const testItems: UrlItem[] = [
  {
    id: 1,
    shortcode: 'abc123',
    long_url: 'https://example.com/very/long/path',
    short_url: 'https://short.example.com/r/abc123',
    last_updated: '2026-01-01T00:00:00Z',
  },
  {
    id: 2,
    shortcode: 'def456',
    long_url: 'https://another.example.com/path',
    short_url: 'https://short.example.com/r/def456',
    last_updated: '2026-01-02T00:00:00Z',
  },
]

function makeStoreMock() {
  return {
    items: testItems,
    total: 2,
    page: 1,
    pageSize: 20,
    loading: false,
    error: null,
    creating: false,
    createError: null,
    updating: false,
    updateError: null,
    deleting: false,
    deleteError: null,
    fetchUrls: vi.fn(),
    createUrl: vi.fn(),
    updateUrl: vi.fn(),
    deleteUrl: vi.fn(),
    refreshItems: vi.fn(),
    clearCreateError: vi.fn(),
    clearUpdateError: vi.fn(),
    clearDeleteError: vi.fn(),
  }
}

// ── setup / teardown ──────────────────────────────────────────────────────────

let mountTarget: HTMLDivElement

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(useUrlsStore).mockReturnValue(
    makeStoreMock() as unknown as ReturnType<typeof useUrlsStore>,
  )
  mountTarget = document.createElement('div')
  document.body.appendChild(mountTarget)
})

afterEach(() => {
  document.body.innerHTML = ''
  vi.restoreAllMocks()
})

async function mountView() {
  const wrapper = mount(DashboardView, { attachTo: mountTarget })
  await flushPromises()
  return wrapper
}

// ── copy button — desktop table ───────────────────────────────────────────────

describe('copy button in desktop table', () => {
  it('renders a Copy button for each URL item', async () => {
    const wrapper = await mountView()
    const table = wrapper.find('.hidden')
    const copyButtons = table.findAll('button').filter((btn) => btn.text() === 'Copy')
    expect(copyButtons).toHaveLength(testItems.length)
  })

  it('calls clipboard copy with the correct short_url on click', async () => {
    const wrapper = await mountView()
    const table = wrapper.find('.hidden')
    const copyButtons = table.findAll('button').filter((btn) => btn.text() === 'Copy')
    await copyButtons[0].trigger('click')
    expect(mockCopy).toHaveBeenCalledWith(testItems[0].short_url)
  })

  it('shows Copied! text on the clicked button after click', async () => {
    const wrapper = await mountView()
    const table = wrapper.find('.hidden')
    const copyButtons = table.findAll('button').filter((btn) => btn.text() === 'Copy')
    await copyButtons[0].trigger('click')
    await wrapper.vm.$nextTick()
    const copiedButtons = table.findAll('button').filter((btn) => btn.text() === 'Copied!')
    expect(copiedButtons).toHaveLength(1)
  })

  it('does not change Copy text on the second button after clicking the first', async () => {
    const wrapper = await mountView()
    const table = wrapper.find('.hidden')
    const copyButtons = table.findAll('button').filter((btn) => btn.text() === 'Copy')
    await copyButtons[0].trigger('click')
    await wrapper.vm.$nextTick()
    const stillCopyButtons = table.findAll('button').filter((btn) => btn.text() === 'Copy')
    expect(stillCopyButtons).toHaveLength(1)
  })
})

// ── copy button — mobile cards ────────────────────────────────────────────────

describe('copy button in mobile cards', () => {
  it('renders a copy icon button for each URL item', async () => {
    const wrapper = await mountView()
    const cards = wrapper.find('.md\\:hidden')
    const copyButtons = cards.findAll('button[aria-label="Copy short link"]')
    expect(copyButtons).toHaveLength(testItems.length)
  })

  it('calls clipboard copy with the correct short_url on click', async () => {
    const wrapper = await mountView()
    const cards = wrapper.find('.md\\:hidden')
    const copyButtons = cards.findAll('button[aria-label="Copy short link"]')
    await copyButtons[1].trigger('click')
    expect(mockCopy).toHaveBeenCalledWith(testItems[1].short_url)
  })
})
