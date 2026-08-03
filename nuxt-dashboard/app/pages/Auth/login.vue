<template>
  <div class="p-8">
    <div class="mb-7 text-center">
      <div class="mb-3 flex items-center justify-center">
        <svg class="h-6 w-6 text-teal-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
      </div>
      <h2 class="text-xl font-semibold text-gray-100">Welcome back</h2>
      <p class="mt-1.5 text-sm text-gray-500">Sign in to your operations dashboard</p>
    </div>
    <form class="animate-fade-up space-y-5" @submit.prevent="onSubmit">
      <div>
        <label for="email" class="app-label mb-1.5 block">Email</label>
        <div class="relative">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            autocomplete="email"
            class="app-input w-full pl-10 placeholder:text-gray-600"
            placeholder="you@example.com"
          >
        </div>
      </div>
      <div>
        <label for="password" class="app-label mb-1.5 block">Password</label>
        <div class="relative">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <input
            id="password"
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            required
            autocomplete="current-password"
            class="app-input w-full pl-10 pr-10 placeholder:text-gray-600"
          >
          <button
            type="button"
            class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-500 hover:text-gray-300"
            @click="showPassword = !showPassword"
          >
            <svg v-if="!showPassword" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
            <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878l4.242 4.242M21 21l-3.122-3.122" />
            </svg>
          </button>
        </div>
      </div>
      <div class="flex items-center justify-between">
        <label class="flex items-center gap-2 text-sm text-gray-400">
          <input
            v-model="rememberMe"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-600 bg-gray-800 text-teal-500 focus:ring-teal-500"
          >
          Remember me
        </label>
        <a href="#" class="text-sm font-medium text-teal-400 hover:text-teal-300">Forgot password?</a>
      </div>
      <div v-if="error" class="app-banner-error">{{ error }}</div>
      <button
        type="submit"
        class="btn-primary w-full disabled:opacity-50"
        :disabled="pending"
      >
        <svg v-if="pending" class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24" aria-hidden="true">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        {{ pending ? 'Signing in…' : 'Sign in' }}
      </button>
      <div class="relative my-2">
        <div class="absolute inset-0 flex items-center">
          <div class="w-full border-t border-gray-700" />
        </div>
        <div class="relative flex justify-center text-sm">
          <span class="bg-gray-900 px-3 text-gray-500">or continue with</span>
        </div>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <button
          type="button"
          class="flex items-center justify-center gap-2 rounded-lg border border-gray-700 bg-transparent px-4 py-2.5 text-sm font-medium text-gray-300 transition hover:bg-gray-800"
          @click.prevent
        >
          <svg class="h-5 w-5" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 01-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4" />
            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853" />
            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18A10.96 10.96 0 001 12c0 1.77.42 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05" />
            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335" />
          </svg>
          Google
        </button>
        <button
          type="button"
          class="flex items-center justify-center gap-2 rounded-lg border border-gray-700 bg-transparent px-4 py-2.5 text-sm font-medium text-gray-300 transition hover:bg-gray-800"
          @click.prevent
        >
          <svg class="h-5 w-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
          </svg>
          GitHub
        </button>
      </div>
    </form>
    <p class="mt-6 text-center text-sm text-gray-500">
      No account?
      <NuxtLink to="/auth/register" class="font-medium text-teal-400 hover:text-teal-300">Register</NuxtLink>
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'auth',
  middleware: 'guest',
})

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const rememberMe = ref(false)
const pending = ref(false)
const error = ref<string | null>(null)

async function onSubmit() {
  error.value = null
  pending.value = true
  try {
    await auth.login(email.value.trim(), password.value)
    await router.push('/dashboard')
  }
  catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Login failed'
  }
  finally {
    pending.value = false
  }
}
</script>

<style scoped>
.animate-fade-up {
  animation: fadeUp 0.4s ease-out both;
}

@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
