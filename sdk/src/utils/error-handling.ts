import type { FluxbaseError, FluxbaseResponse, VoidResponse } from '../types'

/**
 * Normalize a thrown value into the FluxbaseError shape carried by response
 * envelopes. Errors thrown by the SDK's fetch layer are already hydrated with
 * `status` / `code` / `details` — they are returned as-is so that context is
 * preserved. Plain Errors and non-Error throwables are wrapped in an Error
 * carrying the FluxbaseError shape.
 */
function toFluxbaseError(error: unknown): FluxbaseError {
  if (error instanceof Error) {
    return error as FluxbaseError
  }
  const wrapped = new Error(
    typeof error === 'string' ? error : JSON.stringify(error) || String(error)
  ) as FluxbaseError
  return wrapped
}

/**
 * Wraps an async operation with try-catch, returning Fluxbase response format
 * @param operation The async operation to wrap
 * @returns Promise resolving to { data, error: null } on success or { data: null, error } on failure
 */
export async function wrapAsync<T>(
  operation: () => Promise<T>
): Promise<FluxbaseResponse<T>> {
  try {
    const data = await operation()
    return { data, error: null }
  } catch (error) {
    return {
      data: null,
      error: toFluxbaseError(error)
    }
  }
}

/**
 * Wraps void async operations (operations that don't return data)
 * @param operation The async void operation to wrap
 * @returns Promise resolving to { error: null } on success or { error } on failure
 */
export async function wrapAsyncVoid(
  operation: () => Promise<void>
): Promise<VoidResponse> {
  try {
    await operation()
    return { error: null }
  } catch (error) {
    return {
      error: toFluxbaseError(error)
    }
  }
}

/**
 * Wraps synchronous operations (for consistency)
 * @param operation The sync operation to wrap
 * @returns { data, error: null } on success or { data, error } on failure
 */
export function wrapSync<T>(
  operation: () => T
): FluxbaseResponse<T> {
  try {
    const data = operation()
    return { data, error: null }
  } catch (error) {
    return {
      data: null,
      error: toFluxbaseError(error)
    }
  }
}
