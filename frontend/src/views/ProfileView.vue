<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'

const auth = useAuthStore()
</script>

<template>
    <div class="container mx-auto px-6 py-8">
        <h1 class="text-2xl font-bold mb-6">Profile</h1>
        <div class="max-w-md space-y-6">
            <div class="space-y-1">
                <p class="text-sm text-muted-foreground">Username</p>
                <p class="text-base font-medium">{{ auth.user?.username }}</p>
            </div>
            <Button as-child variant="outline">
                <RouterLink to="/profile/settings">Settings</RouterLink>
            </Button>
        </div>
    </div>
</template>
