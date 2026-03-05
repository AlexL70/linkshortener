/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Runtime mode selector. Injected by vite.config.ts from LINKSHORTENER_ENV. */
  readonly LINKSHORTENER_ENV: 'dev' | 'prod'
  /** Base URL of the backend API, e.g. http://localhost:8080. */
  readonly APP_BASE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
