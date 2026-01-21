import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'

const originalErrorHandler = window.onerror
window.onerror = function(message, source, lineno, colno, error) {
  if (message && message.includes('ResizeObserver loop')) {
    return true
  }
  if (originalErrorHandler) {
    return originalErrorHandler(message, source, lineno, colno, error)
  }
  return false
}

const originalUnhandledRejection = window.onunhandledrejection
window.onunhandledrejection = function(event) {
  if (event.reason && event.reason.message && event.reason.message.includes('ResizeObserver loop')) {
    event.preventDefault()
    return true
  }
  if (originalUnhandledRejection) {
    return originalUnhandledRejection(event)
  }
  return false
}

const app = createApp(App)
app.config.errorHandler = (err, instance, info) => {
  if (err.message && err.message.includes('ResizeObserver loop')) {
    return
  }
  console.error('Vue error:', err, info)
}

app.use(ElementPlus)
app.use(router)
app.mount('#app')
