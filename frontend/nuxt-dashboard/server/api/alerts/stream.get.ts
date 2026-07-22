export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const apiBase = (config.public.apiBase as string).replace(/\/$/, '')

  const controller = new AbortController()
  event.node.req.on('close', () => controller.abort())

  // This runs server-side (no cookie jar), so `credentials: 'include'` would
  // do nothing — forward the browser's auth_token cookie to the upstream
  // stream explicitly instead of an X-API-Key header.
  const upstream = await fetch(`${apiBase}/api/v1/alerts/stream`, {
    headers: {
      Accept: 'text/event-stream',
      Cookie: getRequestHeader(event, 'Cookie') ?? '',
    },
    signal: controller.signal,
  })

  if (!upstream.ok || !upstream.body) {
    throw createError({
      statusCode: upstream.status,
      statusMessage: `Alert stream upstream failed: ${upstream.statusText}`,
    })
  }

  return new Response(upstream.body, {
    status: 200,
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache, no-transform',
      Connection: 'keep-alive',
    },
  })
})
