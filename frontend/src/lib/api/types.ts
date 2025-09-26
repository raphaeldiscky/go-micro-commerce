export interface ApiErrorResponse {
  errors?: Array<{ field: string; message: string }>
  message: string
}

// Universal Links structure matching backend dto.Links
export interface PaginationLinks {
  self: string
  first: string
  prev: string
  next: string
  last: string
}

// Universal Pagination metadata matching backend dto.PageMetaData
export interface PaginationInfo {
  page: number
  size: number
  total_item: number
  total_page: number
  links: PaginationLinks
}

// Universal WebResponse matching backend dto.WebResponse[T]
export interface WebResponse<T> {
  message: string
  data: T
  pagination?: PaginationInfo | null
}

// Paginated response type (when pagination is always present)
export interface PaginatedWebResponse<T> {
  message: string
  data: Array<T>
  pagination: PaginationInfo
}

// Backward compatibility aliases
export interface ApiPaginatedResponse<T> extends PaginatedWebResponse<T> {}
export type ApiSuccessResponse<T> = WebResponse<T>

export class ApiError extends Error {
  public readonly status?: number

  constructor(message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}
