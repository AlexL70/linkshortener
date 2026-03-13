<script setup lang="ts">
import { onMounted, computed } from 'vue'
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

const urlsStore = useUrlsStore()

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
  <div class="flex min-h-screen flex-col gap-6 p-8">
    <h1 class="text-3xl font-bold">Dashboard</h1>

    <!-- Error state -->
    <div v-if="urlsStore.error" class="rounded-md bg-destructive/10 border border-destructive/30 px-4 py-3 text-destructive text-sm">
      {{ urlsStore.error }}
    </div>

    <!-- Loading skeleton -->
    <div v-if="urlsStore.loading" class="flex items-center justify-center py-16 text-muted-foreground text-sm">
      Loading your URLs…
    </div>

    <!-- Table -->
    <div v-else class="rounded-md border">
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
            <TableCell class="max-w-xs truncate text-muted-foreground">
              <a :href="url.long_url" target="_blank" rel="noopener noreferrer" class="hover:underline">
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

    <!-- Pagination -->
    <div
      v-if="!urlsStore.loading && totalPages > 1"
      class="flex items-center justify-between text-sm text-muted-foreground"
    >
      <span>Page {{ urlsStore.page }} of {{ totalPages }} ({{ urlsStore.total }} total)</span>
      <div class="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          :disabled="urlsStore.page <= 1"
          @click="goToPage(urlsStore.page - 1)"
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          :disabled="urlsStore.page >= totalPages"
          @click="goToPage(urlsStore.page + 1)"
        >
          Next
        </Button>
      </div>
    </div>
  </div>
</template>

