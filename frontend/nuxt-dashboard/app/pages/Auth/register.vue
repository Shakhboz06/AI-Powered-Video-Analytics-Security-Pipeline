<template>
  <div class="p-8">
    <div class="mb-7 text-center">
      <h2 class="text-xl font-semibold text-gray-100">Create account</h2>
      <p class="mt-1.5 text-sm text-gray-500">Start monitoring with AI-powered video analytics</p>
    </div>
    <form class="space-y-5" @submit.prevent="onSubmit">
      <div>
        <label for="username" class="app-label mb-1.5 block">Username</label>
        <input
          id="username"
          v-model="username"
          type="text"
          required
          autocomplete="username"
          class="app-input w-full placeholder:text-gray-600"
          placeholder="operator"
        >
      </div>
      <div>
        <label for="email" class="app-label mb-1.5 block">Email</label>
        <input
          id="email"
          v-model="email"
          type="email"
          required
          autocomplete="email"
          class="app-input w-full placeholder:text-gray-600"
          placeholder="you@example.com"
        >
      </div>
      <div>
        <label for="password" class="app-label mb-1.5 block">Password</label>
        <input
          id="password"
          v-model="password"
          type="password"
          required
          autocomplete="new-password"
          class="app-input w-full placeholder:text-gray-600"
        >
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
        {{ pending ? 'Creating account…' : 'Create account' }}
      </button>
    </form>
    <p class="mt-6 text-center text-sm text-gray-500">
      Already have an account?
      <NuxtLink to="/auth/login" class="font-medium text-teal-400 hover:text-teal-300">Sign in</NuxtLink>
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

const username = ref('')
const email = ref('')
const password = ref('')
const pending = ref(false)
const error = ref<string | null>(null)

async function onSubmit() {
  error.value = null
  pending.value = true
  try {
    await auth.register(username.value.trim(), email.value.trim(), password.value)
    await router.push('/dashboard')
  }
  catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Registration failed'
  }
  finally {
    pending.value = false
  }
}
</script>
