<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import { useUrlsStore } from '@/stores/urls'
import {
  Table,
  TableHeader,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
  TableEmpty,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ChevronDown, ChevronUp } from 'lucide-vue-next'

const urlsStore = useUrlsStore()
const expandedId = ref<string | null>(null)

function toggleCard(id: string) {
  expandedId.value = expandedId.value === id ? null : id
}

const totalPages = computed(() =>
  urlsStore.pageSize > 0 ? Math.ceil(urlsStore.total / urlsStore.pageSize) : 1,
)

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function isExpired(expiresAt?: string): boolean {
  if (!expiresAt) return false
  return new Date(expiresAt) < new Date()
}

function goToPage(p: number) {
  urlsStore.fetchUrls(p, urlsStore.pageSize)
}

onMounted(() => {
  urlsStore.fetchUrls()
})
</script>

<template>
  <div class="flex min-h-screen flex-col gap-6 p-4 sm:p-8">
    <h1 class="text-3xl font-bold">Dashboard</h1>

    <!-- Error state -->
    <div v-if="urlsStore.error"
      class="rounded-md bg-destructive/10 border border-destructive/30 px-4 py-3 text-destructive text-sm">
      {{ urlsStore.error }}
    </div>

    <!-- Loading skeleton -->
    <div v-if="urlsStore.loading" class="flex items-center justify-center py-16 text-muted-foreground text-sm">
      Loading your URLs…
    </div>

    <!-- Table (desktop) + Cards (mobile) -->
    <template v-else>
      <!-- Table: md and above -->
      <div class="hidden md:block rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Short URL</TableHead>
              <TableHead>Long URL</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Expires</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="urlsStore.items.length === 0">
              You haven't created any short URLs yet.
            </TableEmpty>
            <TableRow v-for="url in urlsStore.items" :key="url.id">
              <TableCell class="font-mono font-medium">{{ url.shortcode }}</TableCell>
              <TableCell class="max-w-0 w-full text-muted-foreground">
                <a :href="url.long_url" target="_blank" rel="noopener noreferrer"
                  class="block truncate hover:underline">
                  {{ url.long_url }}
                </a>
              </TableCell>
              <TableCell>
                <Badge v-if="isExpired(url.expires_at)" variant="destructive">Expired</Badge>
                <Badge v-else variant="secondary">Active</Badge>
              </TableCell>
              <TableCell class="whitespace-nowrap text-sm">{{ formatDate(url.created_at) }}</TableCell>
              <TableCell class="whitespace-nowrap text-sm text-muted-foreground">
                {{ url.expires_at ? formatDate(url.expires_at) : '—' }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <!-- Cards: below md -->
      <div class="md:hidden flex flex-col gap-3">
        <p v-if="urlsStore.items.length === 0" class="py-8 text-center text-sm text-muted-foreground">
          You haven't created any short URLs yet.
        </p>
        <div v-for="url in urlsStore.items" :key="url.id" class="cursor-pointer rounded-md border p-4"
          @click="toggleCard(String(url.id))">
          <!-- Row 1: shortcode + badge + chevron -->
          <div class="flex items-center gap-2">
            <span class="flex-1 font-mono font-medium">{{ url.shortcode }}</span>
            <Badge v-if="isExpired(url.expires_at)" variant="destructive">Expired</Badge>
            <Badge v-else variant="secondary">Active</Badge>
            <ChevronUp v-if="expandedId === String(url.id)" class="h-4 w-4 text-muted-foreground" />
            <ChevronDown v-else class="h-4 w-4 text-muted-foreground" />
          </div>
          <!-- Row 2: truncated URL preview (always visible) -->
          <p class="mt-1 truncate text-sm text-muted-foreground">{{ url.long_url }}</p>
          <!-- Expanded details -->
          <template v-if="expandedId === String(url.id)">
            <a :href="url.long_url" target="_blank" rel="noopener noreferrer"
              class="mt-2 block break-all text-sm text-muted-foreground hover:underline" @click.stop>
              {{ url.long_url }}
            </a>
            <div class="mt-2 flex gap-4 text-xs text-muted-foreground">
              <span>Created {{ formatDate(url.created_at) }}</span>
              <span v-if="url.expires_at">Expires {{ formatDate(url.expires_at) }}</span>
              <span v-else>No expiry</span>
            </div>
          </template>
        </div>
      </div>
    </template>

    <!-- Pagination -->
    <div v-if="!urlsStore.loading && totalPages > 1"
      class="flex items-center justify-between text-sm text-muted-foreground">
      <span>Page {{ urlsStore.page }} of {{ totalPages }} ({{ urlsStore.total }} total)</span>
      <div class="flex gap-2">
        <Button variant="outline" size="sm" :disabled="urlsStore.page <= 1" @click="goToPage(urlsStore.page - 1)">
          Previous
        </Button>
        <Button variant="outline" size="sm" :disabled="urlsStore.page >= totalPages"
          @click="goToPage(urlsStore.page + 1)">
          Next
        </Button>
      </div>
    </div>
  </div>
</template>
