/// <reference types="vitest" />
import { defineConfig } from 'vitest/config'
import { loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'
import { resolve } from 'node:path'

// https://vite.dev/config/
export default defineConfig(() => {
  const projectRoot = resolve(fileURLToPath(new URL('.', import.meta.url)), '..')

  // Step 1 — load ../.env (base values, no mode suffix)
  const baseEnv = loadEnv('', projectRoot, '')

  // Step 2 — resolve LINKSHORTENER_ENV: OS env takes precedence over .env file value
  const lsEnv: string = process.env['LINKSHORTENER_ENV'] ?? baseEnv['LINKSHORTENER_ENV'] ?? ''

  // Step 3 — load ../.env.<lsEnv> on top of ../.env; loadEnv also merges process.env so OS vars
  //           always win regardless of what any file declares
  const mergedEnv: Record<string, string> = lsEnv ? loadEnv(lsEnv, projectRoot, '') : baseEnv

  // Expose every VITE_* variable from the merged env to client code via import.meta.env
  const define: Record<string, string> = {}
  for (const [key, value] of Object.entries(mergedEnv)) {
    if (key.startsWith('VITE_')) {
      define[`import.meta.env.${key}`] = JSON.stringify(value)
    }
  }

  return {
    plugins: [vue(), tailwindcss()],
    define,
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    test: {
      environment: 'jsdom',
      globals: true,
      coverage: {
        provider: 'v8' as const,
        thresholds: {
          lines: 80,
          functions: 80,
          branches: 80,
          statements: 80,
        },
      },
    },
  }
})
