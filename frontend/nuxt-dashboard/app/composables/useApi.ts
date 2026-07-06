import type {
  CameraConfig,
  CameraConfigResponse,
  CamerasResponse,
  ClassesResponse,
  HealthResponse,
  LiveResponse,
  LoginResponse,
  MetricsResponse,
  Point,
  RegisterResponse,
  SecurityAlert,
  SecurityZone,
  SummaryResponse,
  UploadJobResponse,
  ZoneListResponse,
} from '~/types/security'
import type { AuthUser } from '~/types/security'

/**
 * Backend may return `class_summary` as either a single object or (e.g. Go) a one-element array of maps:
 * `{ "person": 12 }` or `[{ "person": 12, "car": 3 }]`. Charts need a flat `Record<string, number>`.
 */
export function normalizeClassSummary(raw: unknown): Record<string, number> {
  const out: Record<string, number> = {}
  if (raw == null) return out

  const rows: Record<string, unknown>[] = []
  if (Array.isArray(raw)) {
    for (const item of raw) {
      if (item && typeof item === 'object' && !Array.isArray(item))
        rows.push(item as Record<string, unknown>)
    }
  }
  else if (typeof raw === 'object') {
    rows.push(raw as Record<string, unknown>)
  }

  for (const row of rows) {
    for (const [k, v] of Object.entries(row)) {
      const n = typeof v === 'number' ? v : Number(v)
      if (!Number.isNaN(n)) out[k] = (out[k] ?? 0) + n
    }
  }
  return out
}

export function useApi() {
  const config = useRuntimeConfig()
  const token = useCookie<string | null>('auth_token')

  const base = () => (config.public.apiBase as string).replace(/\/$/, '')

  function headersAuthOnly(): Record<string, string> {
    const h: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    if (token.value) h.Authorization = `Bearer ${token.value}`
    return h
  }

  function headersData(): Record<string, string> {
    const h = headersAuthOnly()
    const key = config.public.apiKey as string
    if (key) h['X-API-Key'] = key
    return h
  }

  async function register(body: { username: string; email: string; password: string }) {
    return await $fetch<RegisterResponse>(`${base()}/api/v1/dashboard/user/register`, {
      method: 'POST',
      body,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  async function login(body: { email: string; password: string }) {
    return await $fetch<LoginResponse>(`${base()}/api/v1/dashboard/user/login`, {
      method: 'POST',
      body,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  async function getCurrentUser() {
    return await $fetch<AuthUser>(`${base()}/api/v1/dashboard/user`, {
      headers: headersAuthOnly(),
    })
  }

  async function getCameras() {
    const res = await $fetch<{ cameras: Array<CameraConfig | string> }>(`${base()}/api/v1/cameras`, {
      headers: headersData(),
    })
    return {
      cameras: (res.cameras ?? [])
        .filter((camera) => typeof camera === 'string' || camera.is_active)
        .map((camera) => typeof camera === 'string' ? camera : camera.camera_name)
        .filter(Boolean),
    } satisfies CamerasResponse
  }

  async function getCameraConfigs() {
    return await $fetch<CameraConfigResponse>(`${base()}/api/v1/cameras`, {
      headers: headersData(),
    })
  }

  type CameraBody = {
    camera_name: string
    video_source: string
    is_active: boolean
  }

  async function createCamera(body: CameraBody) {
    return await $fetch<{ camera: CameraConfig }>(`${base()}/api/v1/cameras`, {
      method: 'POST',
      headers: headersData(),
      body,
    })
  }

  async function updateCamera(id: number, body: CameraBody) {
    return await $fetch<{ camera: CameraConfig }>(`${base()}/api/v1/cameras/${id}`, {
      method: 'PUT',
      headers: headersData(),
      body,
    })
  }

  async function deleteCamera(id: number) {
    await $fetch<void>(`${base()}/api/v1/cameras/${id}`, {
      method: 'DELETE',
      headers: headersData(),
    })
  }

  async function getDetectionCameras() {
    return await $fetch<CamerasResponse>(`${base()}/api/v1/detections/cameras`, {
      headers: headersData(),
    })
  }

  async function getMetrics(camera: string, time: string) {
    return await $fetch<MetricsResponse>(`${base()}/api/v1/detections/metrics`, {
      headers: headersData(),
      query: { camera, time },
    })
  }

  async function getLive(camera: string) {
    return await $fetch<LiveResponse>(`${base()}/api/v1/detections/live`, {
      headers: headersData(),
      query: { camera },
    })
  }

  async function getSummary(params: {
    camera: string
    bucket: string
    start: string
    end: string
  }) {
    return await $fetch<SummaryResponse>(`${base()}/api/v1/detections/summary`, {
      headers: headersData(),
      query: {
        camera: params.camera,
        bucket: params.bucket,
        start: params.start,
        end: params.end,
      },
    })
  }

  async function getClasses(camera: string, start: string, end: string) {
    const res = await $fetch<{ class_summary: unknown }>(`${base()}/api/v1/detections/class/summary`, {
      headers: headersData(),
      query: { camera, start, end },
    })
    return {
      class_summary: normalizeClassSummary(res.class_summary),
    } satisfies ClassesResponse
  }

  async function getHealth(params: {
    camera: string
    start: string
    end: string
    threshold?: number
  }) {
    return await $fetch<HealthResponse>(`${base()}/api/v1/detections/latency`, {
      headers: headersData(),
      query: {
        camera: params.camera,
        start: params.start,
        end: params.end,
        ...(params.threshold != null ? { threshold: String(params.threshold) } : {}),
      },
    })
  }

  type ZoneBody = {
    name: string
    camera: string
    polygon: Point[]
    is_active: boolean
    active_from: string | null
    active_until: string | null
    loiter_threshold_seconds?: number | null
    default_severity?: string | null
  }

  async function listZones(camera: string) {
    return await $fetch<ZoneListResponse>(`${base()}/api/v1/zones/list`, {
      headers: headersData(),
      query: { camera },
    })
  }

  async function createZone(body: ZoneBody) {
    return await $fetch<{ data: SecurityZone }>(`${base()}/api/v1/zones`, {
      method: 'POST',
      headers: headersData(),
      body,
    })
  }

  async function updateZone(id: number, body: ZoneBody) {
    return await $fetch<{ data: SecurityZone }>(`${base()}/api/v1/zones/${id}`, {
      method: 'PUT',
      headers: headersData(),
      body,
    })
  }

  async function deleteZone(id: number) {
    await $fetch<void>(`${base()}/api/v1/zones/${id}`, {
      method: 'DELETE',
      headers: headersData(),
    })
  }

  async function listAlerts(params: { camera?: string; status?: 'all' | 'new' | 'acknowledged' | 'resolved' }) {
    return await $fetch<{ alerts: SecurityAlert[] }>(`${base()}/api/v1/alerts`, {
      headers: headersData(),
      query: {
        ...(params.camera ? { camera: params.camera } : {}),
        ...(params.status && params.status !== 'all' ? { status: params.status } : {}),
      },
    })
  }

  async function patchAlertStatus(id: number, status: 'acknowledged' | 'resolved' | 'new') {
    return await $fetch<{ alert: SecurityAlert }>(`${base()}/api/v1/alerts/${id}`, {
      method: 'PATCH',
      headers: headersData(),
      body: { status },
    })
  }

  async function getAlertImage(alertId: string) {
    const id = alertId.trim()
    // Same-origin proxy (server adds X-API-Key) — mirrors /api/alerts/stream
    return await $fetch<{ signed_url: string }>(`/api/alerts/${encodeURIComponent(id)}/image`)
  }

  function alertStreamUrl() {
    return '/api/alerts/stream'
  }

  // ── Public video analysis (no auth / API key required) ─────────────

  /** Uploads a video for anonymous analysis, reporting upload progress via `onProgress` (0–100). */
  function uploadVideo(file: File, onProgress?: (percent: number) => void) {
    return new Promise<UploadJobResponse['job']>((resolve, reject) => {
      const form = new FormData()
      form.append('video', file)

      const xhr = new XMLHttpRequest()
      xhr.open('POST', `${base()}/api/v1/public/uploads`)

      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable && onProgress)
          onProgress(Math.round((e.loaded / e.total) * 100))
      }
      xhr.onload = () => {
        try {
          const body = JSON.parse(xhr.responseText || '{}')
          if (xhr.status >= 200 && xhr.status < 300 && body.job)
            resolve(body.job)
          else
            reject(new Error(body.error || `upload failed (${xhr.status})`))
        }
        catch {
          reject(new Error(`upload failed (${xhr.status})`))
        }
      }
      xhr.onerror = () => reject(new Error('network error during upload'))
      xhr.send(form)
    })
  }

  async function getUploadJob(jobId: string) {
    return await $fetch<UploadJobResponse>(`${base()}/api/v1/public/uploads/${encodeURIComponent(jobId)}`)
  }

  async function getUploadAlertImage(jobId: string, alertId: string) {
    return await $fetch<{ signed_url: string }>(
      `${base()}/api/v1/public/uploads/${encodeURIComponent(jobId)}/alerts/${encodeURIComponent(alertId)}/image`,
    )
  }

  return {
    register,
    login,
    getCurrentUser,
    getCameras,
    getCameraConfigs,
    createCamera,
    updateCamera,
    deleteCamera,
    getDetectionCameras,
    getMetrics,
    getLive,
    getSummary,
    getClasses,
    getHealth,
    listZones,
    createZone,
    updateZone,
    deleteZone,
    listAlerts,
    patchAlertStatus,
    getAlertImage,
    alertStreamUrl,
    uploadVideo,
    getUploadJob,
    getUploadAlertImage,
  }
}
