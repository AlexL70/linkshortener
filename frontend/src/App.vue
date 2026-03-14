<script setup lang="ts">
import { RouterView, useRouter } from 'vue-router'
import { ref } from 'vue'
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
const signInOpen = ref(false)

const providers = [
  { id: 'google', label: 'Sign in with Google' },
  { id: 'microsoft', label: 'Sign in with Microsoft' },
  { id: 'facebook', label: 'Sign in with Facebook' },
] as const

function signInWith(provider: string) {
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
    <span class="text-lg font-semibold tracking-tight">Link Shortener</span>
    <div class="flex items-center gap-2">
      <TooltipProvider :delay-duration="300">
        <div class="flex items-center">
          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                variant="ghost" size="icon"
                :class="themeStore.mode === 'light' ? 'bg-accent' : ''"
                aria-label="Light mode"
                @click="themeStore.setMode('light')">
                <Sun />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Light mode</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                variant="ghost" size="icon"
                :class="themeStore.mode === 'dark' ? 'bg-accent' : ''"
                aria-label="Dark mode"
                @click="themeStore.setMode('dark')">
                <Moon />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Dark mode</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                variant="ghost" size="icon"
                :class="themeStore.mode === 'system' ? 'bg-accent' : ''"
                aria-label="Follow system settings"
                @click="themeStore.setMode('system')">
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
    </DialogContent>
  </Dialog>

  <main class="pt-14">
    <RouterView />
  </main>
</template>
