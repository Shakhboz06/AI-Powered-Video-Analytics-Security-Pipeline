import type { AuthUser } from '~/types/security'

export const useAuthStore = defineStore('auth', () => {
  // The JWT lives in an httpOnly `auth_token` cookie the browser manages;
  // JS can't read it, so auth state is derived purely from whether we hold
  // a user object (populated by login/register or a successful getCurrentUser).
  const user = ref<AuthUser | null>(null)

  const isAuthenticated = computed(() => Boolean(user.value))

  async function login(email: string, password: string) {
    const api = useApi()
    // The response still carries a `token` field, but auth now rides on the
    // Set-Cookie the browser stores — we only keep the user object.
    const res = await api.login({ email, password })
    user.value = res.user
  }

  async function register(username: string, email: string, password: string) {
    const api = useApi()
    const res = await api.register({ username, email, password })
    user.value = res.user
  }

  async function fetchUser() {
    const api = useApi()
    const u = await api.getCurrentUser()
    user.value = u
  }

  function logout() {
    // JS cannot clear an httpOnly cookie; we drop local state and let the
    // cookie expire server-side. (Follow-up: add a backend logout endpoint
    // that clears the cookie, then call it here.)
    user.value = null
  }

  return {
    user,
    isAuthenticated,
    login,
    register,
    fetchUser,
    logout,
  }
})
