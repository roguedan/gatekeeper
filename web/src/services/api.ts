import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig } from 'axios'
import { env } from '@/config'
import { storage } from './storage'
import { ErrorResponse } from '@/types'

class APIClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: env.apiUrl,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.setupInterceptors()
  }

  private setupInterceptors() {
    // Request interceptor - add auth token
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = storage.getToken()
        if (token && config.headers) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    // Response interceptor - handle errors
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ErrorResponse>) => {
        if (error.response?.status === 401) {
          // Clear auth on 401
          storage.clear()
          window.dispatchEvent(new Event('auth:logout'))
        }

        const errorMessage = error.response?.data?.message || error.message || 'An error occurred'
        return Promise.reject(new Error(errorMessage))
      }
    )
  }

  get<T>(url: string, config?: any) {
    return this.client.get<T>(url, config)
  }

  post<T>(url: string, data?: any, config?: any) {
    return this.client.post<T>(url, data, config)
  }

  put<T>(url: string, data?: any, config?: any) {
    return this.client.put<T>(url, data, config)
  }

  delete<T>(url: string, config?: any) {
    return this.client.delete<T>(url, config)
  }

  patch<T>(url: string, data?: any, config?: any) {
    return this.client.patch<T>(url, data, config)
  }
}

export const apiClient = new APIClient()
