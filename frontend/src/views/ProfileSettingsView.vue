<script setup lang="ts">
import { Button } from '@/components/ui/button'
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
            <Button variant="destructive">
                Delete Account Permanently
            </Button>
        </section>
    </div>
</template>
