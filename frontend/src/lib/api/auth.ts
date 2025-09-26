import { apiRequest, setAccessToken } from './client'
import type { ApiSuccessResponse } from './types'

export interface AuthResponse {
  access_token: string
  expires_in: number
  token_type: string
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  first_name: string
  last_name: string
  password: string
  username: string
}

export interface User {
  created_at: string
  email: string
  email_verified_at?: string
  first_name: string
  id: string
  is_active: boolean
  is_email_verified: boolean
  last_login_at?: string
  last_name: string
  roles: Array<string>
  updated_at: string
  username: string
}

export async function getCurrentUser(): Promise<User> {
  const response = await apiRequest<ApiSuccessResponse<User>>(
    '/auth/v1/users/whoami',
  )
  return response.data
}

export async function login(credentials: LoginRequest): Promise<AuthResponse> {
  const response = await apiRequest<ApiSuccessResponse<AuthResponse>>(
    '/auth/v1/login',
    {
      body: JSON.stringify(credentials),
      method: 'POST',
    },
  )

  setAccessToken(response.data.access_token)
  return response.data
}

export async function logout(): Promise<void> {
  try {
    await apiRequest('/auth/v1/logout', {
      method: 'POST',
    })
  } finally {
    setAccessToken(null)
  }
}

export async function register(
  userData: RegisterRequest,
): Promise<AuthResponse> {
  const response = await apiRequest<ApiSuccessResponse<AuthResponse>>(
    '/auth/v1/register',
    {
      body: JSON.stringify(userData),
      method: 'POST',
    },
  )

  setAccessToken(response.data.access_token)
  return response.data
}
