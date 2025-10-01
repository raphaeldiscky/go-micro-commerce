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
    const response = await fetch(fullUrl, requestOptions)

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
