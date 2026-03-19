<template>
  <div class="flex min-h-screen flex-col items-center justify-center px-6 py-16">
    <div class="w-full max-w-xl space-y-10">

      <section class="space-y-4 text-center">
        <h1 class="text-4xl font-bold tracking-tight">Shorten your links, simply.</h1>
        <p class="text-lg leading-relaxed text-muted-foreground">
          Paste a long URL, get a short link you can actually share. Sign in with your existing
          Google, Microsoft, or Facebook account — no password required, no new account to set up.
        </p>
      </section>

      <section class="rounded-lg border bg-muted px-6 py-5 space-y-2">
        <p class="text-sm font-semibold">Testing mode — not for production use</p>
        <p class="text-sm text-muted-foreground">
          This app exists solely for experimentation. No data retention guarantees are made:
          stored links and account data may be removed at any time without notice. Use it to
          play around, not to run anything critical.
        </p>
      </section>

      <section class="space-y-3 text-center text-sm text-muted-foreground">
        <p>
          Before signing up, take a moment to read our
          <a href="/privacy-policy" target="_blank" rel="noopener noreferrer"
            class="underline underline-offset-4 hover:text-foreground transition-colors">Privacy Policy</a>
          and
          <a href="/terms-of-service" target="_blank" rel="noopener noreferrer"
            class="underline underline-offset-4 hover:text-foreground transition-colors">Terms of Service</a>.
        </p>
        <p>
          If you like what you see, a
          <a href="https://github.com/AlexL70/linkshortener" target="_blank" rel="noopener noreferrer"
            class="underline underline-offset-4 hover:text-foreground transition-colors">GitHub star</a>
          is always appreciated.
        </p>
      </section>

    </div>
  </div>
</template>
