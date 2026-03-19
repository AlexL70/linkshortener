<script setup lang="ts">
function closeTab() {
    window.close()
}
</script>

<template>
    <div class="min-h-screen bg-background text-foreground">
        <div class="container mx-auto max-w-2xl px-6 py-12">
            <div class="mb-8">
                <p class="text-sm text-muted-foreground mb-1">Link Shortener</p>
                <h1 class="text-3xl font-bold">Terms of Service</h1>
                <p class="text-sm text-muted-foreground mt-2">Last updated: March 19, 2026</p>
            </div>

            <div class="space-y-6 text-base leading-relaxed">
                <section>
                    <h2 class="text-xl font-semibold mb-2">Testing Mode Only</h2>
                    <p>
                        This application is currently operating in testing mode exclusively. It does not offer
                        any services of a commercial or production nature. Access to the application is provided
                        solely for evaluation and experimental purposes.
                    </p>
                </section>

                <section>
                    <h2 class="text-xl font-semibold mb-2">No Service Guarantee</h2>
                    <p>
                        The application is provided "as is" and "as available," with no warranty of any kind,
                        express or implied. The author and domain owner make no representations regarding
                        availability, reliability, or fitness for any purpose.
                    </p>
                </section>

                <section>
                    <h2 class="text-xl font-semibold mb-2">Limitation of Liability</h2>
                    <p>
                        The author and domain owner bear no responsibility whatsoever for any harm — direct,
                        indirect, incidental, or consequential — that users may experience, intentionally or
                        unintentionally, as a result of using this application. By using this application, you
                        accept full responsibility for any consequences arising from your use.
                    </p>
                </section>
            </div>

            <div class="mt-10">
                <button
                    class="rounded-md border border-input bg-background px-4 py-2 text-sm font-medium shadow-sm hover:bg-accent hover:text-accent-foreground transition-colors"
                    @click="closeTab">
                    Close this tab
                </button>
            </div>
        </div>
    </div>
</template>
