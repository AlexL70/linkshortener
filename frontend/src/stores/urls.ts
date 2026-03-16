import { defineStore } from 'pinia'
import { ref } from 'vue'
import { DefaultService } from '@/lib/api/services/DefaultService'
import type { UrlItem } from '@/lib/api/models/UrlItem'
import type { CreateUrlResponseBody } from '@/lib/api/models/CreateUrlResponseBody'
import type { UpdateUrlResponseBody } from '@/lib/api/models/UpdateUrlResponseBody'

export interface UrlsState {
  items: UrlItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  error: string | null
}

export const useUrlsStore = defineStore('urls', () => {
  const items = ref<UrlItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const creating = ref(false)
  const createError = ref<string | null>(null)
  const updating = ref(false)
  const updateError = ref<string | null>(null)

  async function fetchUrls(requestedPage = 1, requestedPageSize = 0) {
    loading.value = true
    error.value = null
    try {
      const response = await DefaultService.listUserUrls({
        page: requestedPage,
        pageSize: requestedPageSize > 0 ? requestedPageSize : undefined,
      })
      if ('status' in response && typeof response.status === 'number') {
        // ErrorModel response
        error.value = (response as { title?: string }).title ?? 'Failed to load URLs'
        return
      }
      const data = response as { items: UrlItem[] | null; total: number; page: number; page_size: number }
      items.value = data.items ?? []
      total.value = data.total
      page.value = data.page
      pageSize.value = data.page_size
    } catch (err) {
      console.error('fetchUrls: request failed', err)
      error.value = 'An unexpected error occurred while loading your URLs.'
    } finally {
      loading.value = false
    }
  }

  function reset() {
    items.value = []
    total.value = 0
    page.value = 1
    pageSize.value = 0
    loading.value = false
    error.value = null
  }

  async function createUrl(params: {
    longUrl: string
    shortcode?: string
    expiresAt?: string
  }): Promise<CreateUrlResponseBody | null> {
    creating.value = true
    createError.value = null
    try {
      const response = await DefaultService.createUrl({
        requestBody: {
          long_url: params.longUrl,
          shortcode: params.shortcode,
          expires_at: params.expiresAt,
        },
      })
      if ('status' in response && typeof response.status === 'number') {
        createError.value = (response as { title?: string }).title ?? 'Failed to create URL'
        return null
      }
      await fetchUrls(1, pageSize.value > 0 ? pageSize.value : undefined)
      return response as CreateUrlResponseBody
    } catch (err) {
      console.error('createUrl: request failed', err)
      createError.value = 'An unexpected error occurred while creating your URL.'
      return null
    } finally {
      creating.value = false
    }
  }

  function clearCreateError() {
    createError.value = null
  }

  async function updateUrl(params: {
    id: number
    longUrl: string
    shortcode?: string
    expiresAt?: string
    lastUpdated: string
  }): Promise<UpdateUrlResponseBody | null> {
    updating.value = true
    updateError.value = null
    try {
      const response = await DefaultService.updateUrl({
        id: params.id,
        requestBody: {
          long_url: params.longUrl,
          shortcode: params.shortcode,
          expires_at: params.expiresAt,
          last_updated: params.lastUpdated,
        },
      })
      if ('status' in response && typeof response.status === 'number') {
        const errResponse = response as { status: number; title?: string }
        if (errResponse.status === 409) {
          updateError.value = 'This item was recently changed by someone else. Please refresh and try again.'
          await fetchUrls(page.value, pageSize.value > 0 ? pageSize.value : undefined)
        } else {
          updateError.value = errResponse.title ?? 'Failed to update URL'
        }
        return null
      }
      await fetchUrls(page.value, pageSize.value > 0 ? pageSize.value : undefined)
      return response as UpdateUrlResponseBody
    } catch (err) {
      console.error('updateUrl: request failed', err)
      updateError.value = 'An unexpected error occurred while updating your URL.'
      return null
    } finally {
      updating.value = false
    }
  }

  function clearUpdateError() {
    updateError.value = null
  }

  return { items, total, page, pageSize, loading, error, creating, createError, updating, updateError, fetchUrls, reset, createUrl, clearCreateError, updateUrl, clearUpdateError }
})
