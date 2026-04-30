<template>
  <div class="w-full max-w-md rounded-xl border border-gray-800 bg-[#12181f] p-8 shadow-xl">
    <div class="mb-8 text-center">
      <h1 class="text-xl font-semibold text-gray-100">Create account</h1>
      <p class="mt-2 text-sm text-gray-500">Register for the security dashboard</p>
    </div>
    <form class="space-y-5" @submit.prevent="onSubmit">
      <div>
        <label for="username" class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1.5">Username</label>
        <input
          id="username"
          v-model="username"
          type="text"
          required
          autocomplete="username"
          class="w-full rounded-lg border border-gray-700 bg-gray-900/60 px-3 py-2.5 text-sm text-gray-100 placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-teal-500/40 focus:border-teal-600"
        >
      </div>
      <div>
        <label for="email" class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1.5">Email</label>
        <input
          id="email"
          v-model="email"
          type="email"
          required
          autocomplete="email"
          class="w-full rounded-lg border border-gray-700 bg-gray-900/60 px-3 py-2.5 text-sm text-gray-100 placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-teal-500/40 focus:border-teal-600"
          placeholder="you@example.com"
        >
      </div>
      <div>
        <label for="password" class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1.5">Password</label>
        <input
          id="password"
          v-model="password"
          type="password"
          required
          autocomplete="new-password"
          class="w-full rounded-lg border border-gray-700 bg-gray-900/60 px-3 py-2.5 text-sm text-gray-100 placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-teal-500/40 focus:border-teal-600"
        >
      </div>
      <div v-if="error" class="rounded-lg bg-red-950/40 border border-red-900/50 px-3 py-2 text-sm text-red-300">
        {{ error }}
      </div>
      <button
        type="submit"
        class="w-full rounded-lg bg-teal-600 hover:bg-teal-500 text-white text-sm font-medium py-2.5 transition-colors disabled:opacity-50"
        :disabled="pending"
      >
        {{ pending ? 'Creating account…' : 'Register' }}
      </button>
    </form>
    <p class="mt-6 text-center text-sm text-gray-500">
      Already have an account?
      <NuxtLink to="/auth/login" class="text-teal-400 hover:text-teal-300 font-medium">Sign in</NuxtLink>
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
    await router.push('/')
  }
  catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Registration failed'
  }
  finally {
    pending.value = false
  }
}
</script>
