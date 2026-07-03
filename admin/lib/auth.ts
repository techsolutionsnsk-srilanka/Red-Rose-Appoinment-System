/**
 * Auth helpers — Clerk edition.
 * The bearer token is read from the active Clerk session (RS256 JWT) and sent
 * to the Go backend, which verifies it against Clerk's JWKS.
 */

export async function getAuthHeaders(): Promise<Record<string, string>> {
  if (typeof window === 'undefined') return {}
  try {
    const token = await (window as any).Clerk?.session?.getToken()
    if (!token) return {}
    return { Authorization: `Bearer ${token}` }
  } catch {
    return {}
  }
}

export async function logout(): Promise<void> {
  if (typeof window === 'undefined') return
  try {
    await (window as any).Clerk?.signOut()
  } finally {
    window.location.href = '/login'
  }
}
