export default defineNuxtPlugin(async () => {
  const auth = useAuthStore()
  const token = useCookie<string | null>('auth_token')
  if (!token.value || auth.user) return
  try {
    await auth.fetchUser()
  }
  catch {
    auth.logout()
  }
})
