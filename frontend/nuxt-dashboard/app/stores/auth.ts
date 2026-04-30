import type { AuthUser } from '~/types/security'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const token = useCookie<string | null>('auth_token', {
    default: () => null,
    maxAge: 60 * 60 * 24 * 7,
    sameSite: 'lax',
    path: '/',
  })

  const isAuthenticated = computed(() => Boolean(token.value))

  async function login(email: string, password: string) {
    const api = useApi()
    const res = await api.login({ email, password })
    token.value = res.token
    user.value = res.user
  }

  async function register(username: string, email: string, password: string) {
    const api = useApi()
    const res = await api.register({ username, email, password })
    token.value = res.token
    user.value = res.user
  }

  async function fetchUser() {
    const api = useApi()
    const u = await api.getCurrentUser()
    user.value = u
  }

  function logout() {
    token.value = null
    user.value = null
  }

  return {
    user,
    token,
    isAuthenticated,
    login,
    register,
    fetchUser,
    logout,
  }
})
