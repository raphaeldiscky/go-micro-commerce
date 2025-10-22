export interface CursorPaginationResult<T> {
  data: Array<T>
  hasNextPage: boolean
  hasPreviousPage: boolean
  startCursor: null | string
  endCursor: null | string
}

export function paginateWithCursor<T extends { id: string }>(
  items: Array<T>,
  limit: number = 10,
  cursor?: null | string,
): CursorPaginationResult<T> {
  let startIndex = 0

  if (cursor) {
    const cursorIndex = items.findIndex((item) => item.id === cursor)
    if (cursorIndex !== -1) {
      startIndex = cursorIndex + 1
    }
  }

  const endIndex = startIndex + limit
  const paginatedItems = items.slice(startIndex, endIndex)
  const hasNextPage = endIndex < items.length
  const hasPreviousPage = startIndex > 0

  return {
    data: paginatedItems,
    endCursor:
      paginatedItems.length > 0
        ? paginatedItems[paginatedItems.length - 1].id
        : null,
    hasNextPage,
    hasPreviousPage,
    startCursor: paginatedItems.length > 0 ? paginatedItems[0].id : null,
  }
}
