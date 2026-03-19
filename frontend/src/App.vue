<script setup lang="ts">
import { RouterView, RouterLink, useRouter, useRoute } from 'vue-router'
import { ref, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Moon, Sun, Monitor } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'

const auth = useAuthStore()
const themeStore = useThemeStore()
const router = useRouter()
const route = useRoute()
const signInOpen = ref(false)
const unavailableNotice = ref(false)
const appVersion = (import.meta.env.APP_VERSION as string | undefined) ?? 'dev'

const providers = [
  { id: 'google', label: 'Sign in with Google' },
  { id: 'microsoft', label: 'Sign in with Microsoft' },
  { id: 'facebook', label: 'Sign in with Facebook' },
] as const

// Clear the notice whenever the dialog is reopened.
watch(signInOpen, (open) => {
  if (!open) unavailableNotice.value = false
})

function signInWith(provider: string) {
  if (provider !== 'google') {
    unavailableNotice.value = true
    return
  }
  signInOpen.value = false
  auth.login(provider)
}

async function handleLogout() {
  await auth.logout()
  router.push('/')
}
</script>

<template>
  <header
    class="fixed top-0 left-0 right-0 z-50 flex h-14 items-center justify-between border-b bg-background px-6 shadow-sm">
    <div class="flex items-center gap-6">
      <span class="text-lg font-semibold tracking-tight">Link Shortener</span>
      <Badge variant="secondary" class="ml-1 text-xs">v{{ appVersion }}</Badge>
      <nav v-if="auth.isAuthenticated" class="flex items-center gap-1">
        <Button as-child variant="ghost" :class="route.name === 'dashboard' ? 'bg-accent' : ''">
          <RouterLink to="/dashboard">Dashboard</RouterLink>
        </Button>
        <Button as-child variant="ghost" :class="route.path.startsWith('/profile') ? 'bg-accent' : ''">
          <RouterLink to="/profile">Profile</RouterLink>
        </Button>
      </nav>
    </div>
    <div class="flex items-center gap-2">
      <TooltipProvider :delay-duration="300">
        <div class="flex items-center">
          <Tooltip>
            <TooltipTrigger as-child>
              <Button variant="ghost" size="icon" :class="themeStore.mode === 'light' ? 'bg-accent' : ''"
                aria-label="Light mode" @click="themeStore.setMode('light')">
                <Sun />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Light mode</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button variant="ghost" size="icon" :class="themeStore.mode === 'dark' ? 'bg-accent' : ''"
                aria-label="Dark mode" @click="themeStore.setMode('dark')">
                <Moon />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Dark mode</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button variant="ghost" size="icon" :class="themeStore.mode === 'system' ? 'bg-accent' : ''"
                aria-label="Follow system settings" @click="themeStore.setMode('system')">
                <Monitor />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Follow system settings</TooltipContent>
          </Tooltip>
        </div>
      </TooltipProvider>
      <Button v-if="!auth.isAuthenticated" variant="outline" @click="signInOpen = true">
        Sign In
      </Button>
      <Button v-else variant="outline" @click="handleLogout">
        Sign Out
      </Button>
    </div>
  </header>

  <Dialog v-model:open="signInOpen">
    <DialogContent class="max-w-sm">
      <DialogHeader>
        <DialogTitle>Sign In</DialogTitle>
        <DialogDescription>Choose a provider to sign in to your account.</DialogDescription>
      </DialogHeader>
      <div class="flex flex-col gap-3 pt-2">
        <Button v-for="provider in providers" :key="provider.id" variant="outline" class="w-full"
          @click="signInWith(provider.id)">
          {{ provider.label }}
        </Button>
      </div>
      <p v-if="unavailableNotice" role="status"
        class="rounded-md border bg-muted px-4 py-3 text-sm text-muted-foreground">
        Sorry, this sign-in method isn't available yet. Please use
        <strong class="text-foreground">Sign in with Google</strong> for now.
      </p>
      <p class="text-center text-xs text-muted-foreground pt-1">
        By signing in, you agree to our
        <a href="/terms-of-service" target="_blank" rel="noopener noreferrer"
          class="underline underline-offset-4 hover:text-foreground transition-colors">Terms of Service</a>
        and
        <a href="/privacy-policy" target="_blank" rel="noopener noreferrer"
          class="underline underline-offset-4 hover:text-foreground transition-colors">Privacy Policy</a>.
      </p>
    </DialogContent>
  </Dialog>

  <main class="pt-14">
    <RouterView />
  </main>
</template>
