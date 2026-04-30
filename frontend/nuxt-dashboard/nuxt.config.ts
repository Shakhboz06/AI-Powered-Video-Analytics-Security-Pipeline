// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@pinia/nuxt'],
  ignore: [
    'app/pages/UiElements/**',
    // Do not ignore Auth/login.vue or Auth/register.vue — only legacy demo auth pages.
    'app/pages/Auth/Signin.vue',
    'app/pages/Auth/Signup.vue',
    'app/pages/Forms/**',
    'app/pages/Chart/**',
    'app/pages/Errors/**',
    'app/pages/Others/**',
    'app/pages/Pages/**',
    'app/pages/Tables/**',
  ],
  css: ['./app/assets/css/main.css'],
  vite: {
    plugins: [
      tailwindcss(),
    ],
  },
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8081',
      apiKey: process.env.NUXT_PUBLIC_API_KEY || '',
    },
  },
  app: {
    head: {
      htmlAttrs: {
        class: 'dark',
      },
      bodyAttrs: {
        class: 'bg-[#0f1419] text-gray-100 antialiased',
      },
      link: [
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Outfit:wght@100..900&display=swap'
        }
      ]
    }
  },
  devServer: {
    port: 8080
  }
})