export default defineNuxtRouteMiddleware(async () => {
  const auth = useAuthStore()
  if (auth.isAuthenticated) return navigateTo('/dashboard')

  // Can't read the httpOnly cookie — verify against the API. If a valid
  // session exists, bounce away from guest-only pages (login/register).
  if (import.meta.client && !auth.user) {
    try {
      await auth.fetchUser()
      return navigateTo('/dashboard')
    }
    catch {
      // anonymous — stay on the guest page
    }
  }
})
