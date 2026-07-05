export default defineEventHandler(async (event) => {
  const alertId = getRouterParam(event, 'alertId')?.trim()
  if (!alertId) {
    throw createError({ statusCode: 400, statusMessage: 'Missing alert_id' })
  }

  const config = useRuntimeConfig(event)
  const apiBase = (config.public.apiBase as string).replace(/\/$/, '')
  const apiKey = config.public.apiKey as string

  const upstream = await fetch(
    `${apiBase}/api/v1/alerts/${encodeURIComponent(alertId)}/image`,
    {
      headers: {
        Accept: 'application/json',
        ...(apiKey ? { 'X-API-Key': apiKey } : {}),
      },
    },
  )

  if (!upstream.ok) {
    throw createError({
      statusCode: upstream.status,
      statusMessage: `Alert image upstream failed: ${upstream.statusText}`,
    })
  }

  return await upstream.json()
})
