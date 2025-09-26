export interface ApiErrorResponse {
  errors?: Array<{ field: string; message: string }>
  message: string
}

export interface ApiSuccessResponse<T> {
  data: T
  message: string
}

export class ApiError extends Error {
  public readonly status?: number

  constructor(message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}
