import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'
import App from './App.vue'
import router from './router'
import { validateConfig, logConfig } from './lib/config'
import { useAuthStore } from './stores/auth'

// Validate required env vars before mounting — throws if any required var is absent.
validateConfig()
// Emit startup config summary to the browser console before the app starts serving.
logConfig()

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)
app.use(router)

// Restore user session from the HttpOnly cookie before mounting.
await useAuthStore().fetchMe()

app.mount('#app')
