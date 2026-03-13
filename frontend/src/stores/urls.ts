import { defineStore } from 'pinia'
import { ref } from 'vue'
import { DefaultService } from '@/lib/api/services/DefaultService'
import type { UrlItem } from '@/lib/api/models/UrlItem'

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
  const pageSize = ref(20)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchUrls(requestedPage = 1, requestedPageSize = 20) {
    loading.value = true
    error.value = null
    try {
      const response = await DefaultService.listUserUrls({
        page: requestedPage,
        pageSize: requestedPageSize,
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
    pageSize.value = 20
    loading.value = false
    error.value = null
  }

  return { items, total, page, pageSize, loading, error, fetchUrls, reset }
})
