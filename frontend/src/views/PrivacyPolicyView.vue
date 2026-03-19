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
                <h1 class="text-3xl font-bold">Privacy Policy</h1>
                <p class="text-sm text-muted-foreground mt-2">Last updated: March 19, 2026</p>
            </div>

            <div class="space-y-6 text-base leading-relaxed">
                <section>
                    <h2 class="text-xl font-semibold mb-2">Data We Collect</h2>
                    <p>
                        This application does not collect any personal information beyond the minimum required
                        to identify your account. Specifically, we store only the <strong>email address</strong>
                        and <strong>user ID</strong> returned by the third-party authentication provider you
                        choose to sign in with (such as Google, Microsoft, or Facebook). No passwords are ever
                        stored.
                    </p>
                </section>

                <section>
                    <h2 class="text-xl font-semibold mb-2">Data Retention and Deletion</h2>
                    <p>
                        All data associated with your account — including your email address and user ID — is
                        permanently and irrecoverably removed from our database when you delete your account.
                        You may delete your account at any time from your profile settings.
                    </p>
                </section>

                <section>
                    <h2 class="text-xl font-semibold mb-2">Security</h2>
                    <p>
                        The author takes every reasonable measure to prevent unauthorized access to stored data
                        and to avoid security vulnerabilities. However, no system can be guaranteed to be
                        completely secure. The author bears no responsibility in the event of a data breach or
                        security incident.
                    </p>
                </section>

                <section>
                    <h2 class="text-xl font-semibold mb-2">Open Source</h2>
                    <p>
                        This application is open source. The full source code is publicly available at
                        <a href="https://github.com/AlexL70/linkshortener" target="_blank" rel="noopener noreferrer"
                            class="underline underline-offset-4 hover:text-foreground transition-colors">
                            github.com/AlexL70/linkshortener
                        </a>.
                        Anyone may download, inspect, and run their own instance of this service. Users who
                        choose to operate their own instance do so entirely at their own risk.
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
