export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const apiBase = (config.public.apiBase as string).replace(/\/$/, '')
  const apiKey = config.public.apiKey as string

  const controller = new AbortController()
  event.node.req.on('close', () => controller.abort())

  const upstream = await fetch(`${apiBase}/api/v1/alerts/stream`, {
    headers: {
      Accept: 'text/event-stream',
      ...(apiKey ? { 'X-API-Key': apiKey } : {}),
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
