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
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { ChevronDown, ChevronUp } from 'lucide-vue-next'

const urlsStore = useUrlsStore()
const expandedId = ref<string | null>(null)

// ── Create Link dialog ────────────────────────────────────────────────────────
const dialogOpen = ref(false)
const longUrl = ref('')
const shortcodeField = ref('')
const expiresAt = ref('')
const hasExpiry = ref(false)

function openDialog() {
  longUrl.value = ''
  shortcodeField.value = ''
  expiresAt.value = ''
  hasExpiry.value = false
  urlsStore.clearCreateError()
  dialogOpen.value = true
}

function onDialogOpenChange(open: boolean) {
  if (!open) {
    urlsStore.clearCreateError()
    longUrl.value = ''
    shortcodeField.value = ''
    expiresAt.value = ''
    hasExpiry.value = false
  }
  dialogOpen.value = open
}

async function submitCreate() {
  const result = await urlsStore.createUrl({
    longUrl: longUrl.value,
    shortcode: shortcodeField.value.trim() || undefined,
    expiresAt: expiresAt.value ? new Date(expiresAt.value).toISOString() : undefined,
  })
  if (result !== null) {
    dialogOpen.value = false
  }
}
// ─────────────────────────────────────────────────────────────────────────────

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
    <div class="flex items-center justify-between gap-4">
      <h1 class="text-3xl font-bold">Dashboard</h1>
      <Button @click="openDialog">Create Link</Button>
    </div>

    <!-- Create Link dialog -->
    <Dialog :open="dialogOpen" @update:open="onDialogOpenChange">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create a new short link</DialogTitle>
          <DialogDescription>
            Paste a long URL below. Optionally provide a 6-character shortcode and an expiry date.
          </DialogDescription>
        </DialogHeader>

        <form class="flex flex-col gap-4" @submit.prevent="submitCreate">
          <div class="flex flex-col gap-1.5">
            <label for="create-long-url" class="text-sm font-medium">Long URL <span
                class="text-destructive">*</span></label>
            <Input id="create-long-url" v-model="longUrl" type="url" placeholder="https://example.com/very/long/path"
              required maxlength="2048" :disabled="urlsStore.creating" />
          </div>

          <div class="flex flex-col gap-1.5">
            <label for="create-shortcode" class="text-sm font-medium">Shortcode <span
                class="text-muted-foreground font-normal">(optional, exactly 6 chars)</span></label>
            <Input id="create-shortcode" v-model="shortcodeField" placeholder="auto-generated" maxlength="6"
              :disabled="urlsStore.creating" />
          </div>

          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium">Expires at <span
                class="text-muted-foreground font-normal">(optional)</span></label>
            <template v-if="hasExpiry">
              <div class="flex gap-2">
                <Input id="create-expires-at" v-model="expiresAt" type="datetime-local" class="flex-1"
                  :disabled="urlsStore.creating" />
                <Button type="button" variant="ghost" size="sm" :disabled="urlsStore.creating"
                  @click="hasExpiry = false; expiresAt = ''">Remove</Button>
              </div>
            </template>
            <template v-else>
              <Button type="button" variant="outline" size="sm" class="self-start" :disabled="urlsStore.creating"
                @click="hasExpiry = true">Add expiry date</Button>
            </template>
          </div>

          <div v-if="urlsStore.createError"
            class="rounded-md bg-destructive/10 border border-destructive/30 px-3 py-2 text-destructive text-sm">
            {{ urlsStore.createError }}
          </div>

          <DialogFooter class="flex gap-2 pt-2">
            <DialogClose as-child>
              <Button type="button" variant="outline" :disabled="urlsStore.creating">Cancel</Button>
            </DialogClose>
            <Button type="submit" :disabled="urlsStore.creating || !longUrl.trim()">
              {{ urlsStore.creating ? 'Creating…' : 'Create' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

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
