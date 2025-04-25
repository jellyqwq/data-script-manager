import axios from 'axios'

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 5000
})

// 自动添加 Authorization 头
instance.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export default instance
