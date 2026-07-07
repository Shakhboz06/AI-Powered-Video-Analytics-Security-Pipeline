// Same-origin proxy for self-serve video uploads — forwards the multipart
// body to the Go API so the browser stays host-agnostic (mirrors /api/alerts/*).
export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const apiBase = (config.public.apiBase as string).replace(/\/$/, '')

  return proxyRequest(event, `${apiBase}/api/uploads`)
})
