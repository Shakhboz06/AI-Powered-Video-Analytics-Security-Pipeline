export default defineNuxtRouteMiddleware(async () => {
  const auth = useAuthStore()
  if (auth.isAuthenticated) return

  // The auth_token cookie is httpOnly (unreadable in JS) and cross-origin,
  // so gate on the API: getCurrentUser succeeds when the cookie is valid,
  // 401s otherwise. Only meaningful on the client, where the browser holds
  // the cookie — defer to the client on server render.
  if (import.meta.client) {
    try {
      await auth.fetchUser()
      return
    }
    catch {
      return navigateTo('/auth/login')
    }
  }
})
