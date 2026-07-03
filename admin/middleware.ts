import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

// Routes that do NOT require authentication.
const isPublicRoute = createRouteMatcher(['/login(.*)', '/signup(.*)'])

export default clerkMiddleware(async (auth, request) => {
  if (!isPublicRoute(request)) {
    await auth.protect()
  }
})

export const config = {
  matcher: [
    // Skip Next.js internals and static files, unless found in search params.
    '/((?!_next/static|_next/image|favicon.ico|.*\\..*).*)',
    // Always run for API routes so clerkMiddleware/auth() are available.
    '/(api|trpc)(.*)',
  ],
}
