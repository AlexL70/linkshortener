import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'
import App from './App.vue'
import router from './router'
import { validateConfig, logConfig } from './lib/config'

// Validate required env vars before mounting — throws if any required var is absent.
validateConfig()
// Emit startup config summary to the browser console before the app starts serving.
logConfig()

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
