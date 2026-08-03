// Same-origin proxy for upload job status polling.
export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const apiBase = (config.public.apiBase as string).replace(/\/$/, '')
  const jobId = getRouterParam(event, 'jobId')

  return proxyRequest(event, `${apiBase}/api/uploads/${encodeURIComponent(jobId ?? '')}`)
})
