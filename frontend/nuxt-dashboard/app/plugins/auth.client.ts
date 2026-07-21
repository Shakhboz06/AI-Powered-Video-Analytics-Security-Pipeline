export default defineNuxtPlugin(async () => {
  const auth = useAuthStore()
  if (auth.user) return
  // The auth_token cookie is httpOnly, so we can't check it in JS. Ask the
  // API who we are (the browser sends the cookie) — success hydrates the
  // user, a 401 just means we're anonymous.
  try {
    await auth.fetchUser()
  }
  catch {
    auth.logout()
  }
})
