import { env } from '@/env'
import type { ApiErrorResponse, ApiSuccessResponse } from './types'
import { ApiError } from './types'

let accessToken: null | string = null

export async function apiRequest<T>(
  url: string,
  options: RequestInit = {},
): Promise<T> {
  const apiGatewayUrl = env.VITE_API_GATEWAY_URL
  const fullUrl = `${apiGatewayUrl}${url}`

  const requestOptions: RequestInit = {
    ...options,
    credentials: 'include',
    headers: {
      ...createHeaders(),
      ...options.headers,
    },
  }

  try {
    let response = await fetch(fullUrl, requestOptions)

    if (response.status === 401 && getAccessToken()) {
      const refreshed = await refreshAccessToken()
      if (refreshed) {
        requestOptions.headers = {
          ...createHeaders(),
          ...options.headers,
        }
        response = await fetch(fullUrl, requestOptions)
      }
    }

    if (!response.ok) {
      if (response.status === 401) {
        setAccessToken(null)
        window.location.href = '/auth/login'
        throw new ApiError('Authentication required', 401)
      }

      let errorData: ApiErrorResponse
      try {
        errorData = await response.json()
      } catch {
        throw new ApiError(
          `Request failed: ${response.status} ${response.statusText}`,
          response.status,
        )
      }

      throw new ApiError(errorData.message, response.status)
    }

    return await response.json()
  } catch (error) {
    if (error instanceof ApiError) {
      throw error
    }

    throw new ApiError(error instanceof Error ? error.message : 'Network error')
  }
}

export function getAccessToken(): null | string {
  return accessToken || localStorage.getItem('access_token')
}

export function setAccessToken(token: null | string): void {
  accessToken = token
  if (token) {
    localStorage.setItem('access_token', token)
  } else {
    localStorage.removeItem('access_token')
  }
}

function createHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  const token = getAccessToken()
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  return headers
}

async function refreshAccessToken(): Promise<boolean> {
  try {
    const response = await fetch(
      `${env.VITE_API_GATEWAY_URL}/auth/v1/refresh-token`,
      {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        method: 'POST',
      },
    )

    if (!response.ok) {
      return false
    }

    const data: ApiSuccessResponse<{ access_token: string }> =
      await response.json()
    setAccessToken(data.data.access_token)
    return true
  } catch {
    return false
  }
}
