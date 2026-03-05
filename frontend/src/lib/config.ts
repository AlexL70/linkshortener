/**
 * Frontend environment configuration: validation and startup logging.
 * Call validateConfig() then logConfig() in main.ts before app.mount().
 */

/** Required frontend variables — always included in startup logs. */
export const REQUIRED_VARS: ReadonlyArray<string> = ['LINKSHORTENER_ENV', 'APP_BASE_URL']

/**
 * Optional frontend variable defaults.
 * Currently no optional VITE_* variables exist; all optional variables in the
 * catalog are backend-only. Extend this map when frontend optional vars are added.
 */
export const OPTIONAL_DEFAULTS: Readonly<Record<string, string>> = {}

/**
 * Variable names whose values must be masked with "****" in prod-mode logs.
 * Matches the secret variables listed in ai-instructions/app-environment.md §4.1.
 */
export const SECRET_VARS: ReadonlySet<string> = new Set([
  'JWT_SECRET',
  'SESSION_SECRET',
  'VITE_GOOGLE_CLIENT_SECRET',
  'VITE_MICROSOFT_CLIENT_SECRET',
  'VITE_FACEBOOK_CLIENT_SECRET',
])

/**
 * Returns the log-safe representation of a configuration value.
 * In dev mode no masking is applied. In prod mode secret vars are replaced with "****".
 */
export function maskedValue(key: string, value: string, isProd: boolean): string {
  if (!isProd) return value
  if (SECRET_VARS.has(key)) return '****'
  return value
}

/**
 * Checks that all required variables are present. Throws if any is absent.
 * Must be called before app.mount() so the app refuses to start on misconfiguration.
 *
 * @param env - Environment object to validate (defaults to import.meta.env).
 */
export function validateConfig(
  env: Record<string, unknown> = import.meta.env as Record<string, unknown>,
): void {
  for (const key of REQUIRED_VARS) {
    if (!env[key]) {
      throw new Error(`Required environment variable ${key} is not set`)
    }
  }
}

/**
 * Logs a startup configuration summary to the browser console (INFO level).
 * Required variables are always logged. Optional variables are logged only when
 * their value differs from the documented default.
 * In prod mode secret variable values are masked; in dev mode all values are printed as-is.
 *
 * @param env - Environment object to log (defaults to import.meta.env).
 */
export function logConfig(
  env: Record<string, unknown> = import.meta.env as Record<string, unknown>,
): void {
  const isProd = String(env['LINKSHORTENER_ENV']) === 'prod'

  for (const key of REQUIRED_VARS) {
    const value = String(env[key] ?? '')
    console.info('[config]', `${key}=${maskedValue(key, value, isProd)}`)
  }

  for (const [key, defaultVal] of Object.entries(OPTIONAL_DEFAULTS)) {
    const value = String(env[key] ?? '')
    if (value !== defaultVal) {
      console.info('[config]', `${key}=${maskedValue(key, value, isProd)}`)
    }
  }
}
