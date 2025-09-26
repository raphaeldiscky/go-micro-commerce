export interface ApiErrorResponse {
  errors?: Array<{ field: string; message: string }>
  message: string
}

export interface ApiSuccessResponse<T> {
  data: T
  message: string
}

export interface PaginationInfo {
  page: number
  size: number
  total_item: number
  total_page: number
  links: {
    self: string
    first: string
    prev: string
    next: string
    last: string
  }
}

export interface ApiPaginatedResponse<T> {
  data: Array<T>
  message: string
  pagination: PaginationInfo
}

export class ApiError extends Error {
  public readonly status?: number

  constructor(message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}
