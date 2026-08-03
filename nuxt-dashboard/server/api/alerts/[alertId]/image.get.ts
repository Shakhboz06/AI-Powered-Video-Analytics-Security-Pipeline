export default defineEventHandler(async (event) => {
  const alertId = getRouterParam(event, 'alertId')?.trim()
  if (!alertId) {
    throw createError({ statusCode: 400, statusMessage: 'Missing alert_id' })
  }

  const config = useRuntimeConfig(event)
  const apiBase = config.apiInternal.replace(/\/$/, '')

  const upstream = await fetch(
    `${apiBase}/api/v1/alerts/${encodeURIComponent(alertId)}/image`,
    {
      headers: {
        Accept: 'application/json',
        Cookie: getRequestHeader(event, 'Cookie') ?? '',
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
