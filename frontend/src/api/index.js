import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

api.interceptors.response.use(
  response => response,
  error => {
    if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时，请检查后端服务是否启动')
    } else if (error.response) {
      console.error('API错误:', error.response.data)
    } else if (error.request) {
      console.error('网络错误:', error.request)
      ElMessage.error('无法连接到后端服务，请确保后端已启动在 http://localhost:8080')
    } else {
      console.error('错误:', error.message)
    }
    return Promise.reject(error)
  }
)

export const getWebPages = () => api.get('/web-pages')
export const createWebPage = (data) => api.post('/web-pages', data)
export const updateWebPage = (id, data) => api.put(`/web-pages/${id}`, data)
export const deleteWebPage = (id) => api.delete(`/web-pages/${id}`)

export const getKeywords = () => api.get('/keywords')
export const createKeyword = (data) => api.post('/keywords', data)
export const deleteKeyword = (id) => api.delete(`/keywords/${id}`)

export const getMonitorConfig = () => api.get('/monitor-config')
export const createMonitorConfig = (data) => api.post('/monitor-config', data)
export const updateMonitorConfig = (id, data) => api.put(`/monitor-config/${id}`, data)
export const deleteMonitorConfig = (id) => api.delete(`/monitor-config/${id}`)

export const getSubscribeConfig = () => api.get('/subscribe-config')
export const createSubscribeConfig = (data) => api.post('/subscribe-config', data)
export const updateSubscribeConfig = (id, data) => api.put(`/subscribe-config/${id}`, data)
export const deleteSubscribeConfig = (id) => api.delete(`/subscribe-config/${id}`)

export const getAnnouncements = (params) => api.get('/announcements', { params })

export const getPushConfig = () => api.get('/push-config')
export const updatePushConfig = (data) => api.put('/push-config', data)
