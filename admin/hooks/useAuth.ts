import { useAuth as useClerkAuth, useUser } from '@clerk/nextjs'

export function useAuth() {
  const { isLoaded, isSignedIn, getToken } = useClerkAuth()
  const { user } = useUser()

  return {
    user: user
      ? {
          id: user.id,
          email: user.primaryEmailAddress?.emailAddress ?? '',
          name: user.fullName ?? user.firstName ?? '',
        }
      : null,
    loading: !isLoaded,
    isAuthenticated: isSignedIn ?? false,
    getToken,
  }
}
