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

  // LINKSHORTENER_ENV is always taken from the OS only — .env files must never override it.
  const lsEnvRaw = process.env['LINKSHORTENER_ENV'] ?? ''

  // During Vitest test runs (process.env.VITEST is set) allow an unset LINKSHORTENER_ENV,
  // defaulting to 'dev', so tests can run without requiring the variable to be exported.
  const isVitest = !!process.env['VITEST']
  if (!isVitest && lsEnvRaw !== 'dev' && lsEnvRaw !== 'prod') {
    throw new Error(`LINKSHORTENER_ENV must be "dev" or "prod", got "${lsEnvRaw}"`)
  }
  const lsEnv: 'dev' | 'prod' =
    lsEnvRaw === 'dev' || lsEnvRaw === 'prod' ? lsEnvRaw : 'dev'

  // prod mode: skip all file reading; use only OS-level environment variables.
  // dev mode : loadEnv with mode='dev' reads ../.env then ../.env.dev; process.env
  //            entries are merged in by Vite so OS values always win.
  const bundleEnv: Record<string, string> =
    lsEnv === 'prod'
      ? (Object.fromEntries(
          Object.entries(process.env).filter((e): e is [string, string] => e[1] !== undefined),
        ) as Record<string, string>)
      : loadEnv('dev', projectRoot, '')

  // Expose every VITE_* and APP_* variable to client code via import.meta.env.
  // Also expose LINKSHORTENER_ENV itself so frontend code can detect the mode.
  const define: Record<string, string> = {}
  for (const [key, value] of Object.entries(bundleEnv)) {
    if (key.startsWith('VITE_') || key.startsWith('APP_')) {
      define[`import.meta.env.${key}`] = JSON.stringify(value)
    }
  }
  define['import.meta.env.LINKSHORTENER_ENV'] = JSON.stringify(lsEnv)

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
