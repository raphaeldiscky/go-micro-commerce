import { env } from '@/env'
import { PATH_AUTH } from '../../constants'
import type { ApiErrorResponse } from './types'
import { ApiError } from './types'

// Store access token in memory only (not localStorage)
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
        window.location.href = PATH_AUTH.login
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

/**
 * Get access token from memory
 * Note: Access token is NOT persisted to localStorage for security
 */
export function getAccessToken(): null | string {
  return accessToken
}

/**
 * Set access token in memory only
 * Note: Refresh token is stored in HTTP-only cookie by the server
 */
export function setAccessToken(token: null | string): void {
  accessToken = token
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
