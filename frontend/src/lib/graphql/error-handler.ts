/**
 * Extract user-friendly error message from GraphQL error
 */
export function extractGraphQLError(error: unknown): string {
  // Handle GraphQL Client Error
  if (error && typeof error === 'object' && 'response' in error) {
    const graphqlError = error as {
      response?: { errors?: Array<{ message: string }> }
    }
    return graphqlError.response?.errors?.[0]?.message || 'An error occurred'
  }

  // Handle standard Error
  if (error instanceof Error) {
    return error.message
  }

  // Fallback for unknown error types
  return 'An unexpected error occurred'
}

/**
 * Wrap GraphQL request with error handling
 */
export async function handleGraphQLRequest<T>(
  requestFn: () => Promise<T>,
  defaultErrorMessage = 'Request failed',
): Promise<T> {
  try {
    return await requestFn()
  } catch (error) {
    const errorMessage = extractGraphQLError(error)
    throw new Error(errorMessage || defaultErrorMessage)
  }
}
