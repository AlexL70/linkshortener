<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const dialogOpen = ref(false)
const confirmEmail = ref('')

const isConfirmed = computed(
    () => confirmEmail.value.toLowerCase() === (auth.user?.email ?? '').toLowerCase(),
)

async function handleDeleteAccount() {
    if (!isConfirmed.value) return
    await auth.deleteAccount()
    if (!auth.deleteError) {
        dialogOpen.value = false
        router.push({ name: 'home' })
    }
}

function openDialog() {
    confirmEmail.value = ''
    dialogOpen.value = true
}
</script>

<template>
    <div class="container mx-auto px-6 py-8">
        <h1 class="text-2xl font-bold mb-6">Settings</h1>
        <section class="max-w-md space-y-4">
            <div>
                <h2 class="text-lg font-semibold text-destructive">Danger Zone</h2>
                <p class="text-sm text-muted-foreground mt-1">
                    Once you delete your account, there is no going back. Please be certain.
                </p>
            </div>

            <Dialog v-model:open="dialogOpen">
                <DialogTrigger as-child>
                    <Button variant="destructive" @click="openDialog">
                        Delete Account Permanently
                    </Button>
                </DialogTrigger>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Delete Account</DialogTitle>
                        <DialogDescription>
                            This action is permanent and cannot be undone. All your shortened URLs and account
                            data will be deleted.
                        </DialogDescription>
                    </DialogHeader>

                    <div class="space-y-3 py-2">
                        <p class="text-sm">
                            To confirm, type your account email address:
                            <span class="font-semibold">{{ auth.user?.email }}</span>
                        </p>
                        <Input v-model="confirmEmail" type="email" placeholder="Enter your email to confirm"
                            autocomplete="off" @paste.prevent @drop.prevent />
                        <p v-if="auth.deleteError" class="text-sm text-destructive">
                            {{ auth.deleteError }}
                        </p>
                    </div>

                    <DialogFooter>
                        <Button variant="outline" :disabled="auth.deleting" @click="dialogOpen = false">
                            Cancel
                        </Button>
                        <Button variant="destructive" :disabled="!isConfirmed || auth.deleting"
                            @click="handleDeleteAccount">
                            {{ auth.deleting ? 'Deleting…' : 'Delete My Account' }}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </section>
    </div>
</template>
