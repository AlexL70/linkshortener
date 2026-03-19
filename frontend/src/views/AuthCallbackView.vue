<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Loader2 } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { DefaultService } from '@/lib/api/services/DefaultService'
import { ApiError } from '@/lib/api/core/ApiError'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

const router = useRouter()
const auth = useAuthStore()

// ── state ───────────────────────────────────────────────────────────────────
const loading = ref(true)
const registrationOpen = ref(false)
const preRegistrationToken = ref('')
const userName = ref('')
const registrationError = ref('')
const registrationLoading = ref(false)
const callbackError = ref('')

// ── helpers ─────────────────────────────────────────────────────────────────
function parseHash(): URLSearchParams {
  return new URLSearchParams(window.location.hash.slice(1))
}

async function submitRegistration() {
  if (!userName.value.trim()) return

  registrationError.value = ''
  registrationLoading.value = true
  try {
    await DefaultService.registerUser({
      requestBody: {
        pre_registration_token: preRegistrationToken.value,
        user_name: userName.value.trim(),
      },
    })
    // Backend sets the session cookie on 204; fetch user state now.
    await auth.fetchMe()
    router.replace('/dashboard')
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      registrationError.value = 'Username already taken. Please choose another.'
    } else {
      registrationError.value = 'Registration failed. Please try again.'
    }
  } finally {
    registrationLoading.value = false
  }
}

// ── lifecycle ────────────────────────────────────────────────────────────────
onMounted(async () => {
  const params = parseHash()

  const preRegToken = params.get('pre_registration_token')
  if (preRegToken) {
    preRegistrationToken.value = preRegToken
    userName.value = params.get('suggested_user_name') ?? ''
    registrationOpen.value = true
    loading.value = false
    return
  }

  // No pre-registration token — the backend redirected here after setting the
  // session cookie. Fetch the user to confirm the session is active.
  await auth.fetchMe()
  if (auth.isAuthenticated) {
    router.replace('/dashboard')
    return
  }

  const error = params.get('error')
  callbackError.value = error ?? 'Authentication failed.'
  loading.value = false
})
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <!-- Spinner while processing a direct-token redirect -->
    <Loader2 v-if="loading" class="size-8 animate-spin text-muted-foreground" />

    <!-- Authentication error from backend -->
    <div v-else-if="callbackError" class="text-center space-y-4">
      <p class="text-destructive">{{ callbackError }}</p>
      <Button variant="outline" @click="router.replace('/')">Go home</Button>
    </div>

    <!-- Registration dialog for new OAuth users -->
    <Dialog :open="registrationOpen">
      <DialogContent class="sm:max-w-md" :hide-close-button="true">
        <DialogHeader>
          <DialogTitle>Create your account</DialogTitle>
          <DialogDescription>
            Choose a username to complete registration.
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-4" @submit.prevent="submitRegistration">
          <Input v-model="userName" placeholder="Username" :disabled="registrationLoading" autocomplete="username" />
          <p v-if="registrationError" class="text-sm text-destructive">
            {{ registrationError }}
          </p>

          <DialogFooter>
            <Button type="submit" :disabled="registrationLoading || !userName.trim()">
              <Loader2 v-if="registrationLoading" class="mr-2 size-4 animate-spin" />
              Create account
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>
