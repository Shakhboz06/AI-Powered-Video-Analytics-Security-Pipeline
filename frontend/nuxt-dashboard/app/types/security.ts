export interface AuthUser {
  id: number
  username: string
  email: string
}

export interface LoginResponse {
  user: AuthUser
  token: string
}

export interface RegisterResponse {
  user: AuthUser
  token: string
}

export interface CamerasResponse {
  cameras: string[]
}

export interface CameraConfig {
  camera_id: number
  camera_name: string
  video_source: string
  is_active: boolean
  created_at?: string
  updated_at?: string
}

export interface CameraConfigResponse {
  cameras: CameraConfig[]
}

export interface DetectionFrame {
  total_detections: number | number[]
  total_objects: Record<string, number> | Record<string, number>[]
  latency_ms: number | number[]
  recorded_at: string | string[]
}

export interface MetricsResponse {
  detections: DetectionFrame[]
}

export interface LiveResponse {
  detections: Array<{
    total_detections: number
    total_objects: Record<string, number>
    latency_ms: number
    recorded_at: string
  }>
}

export interface AggregatedSummaryRow {
  bucket: [string, string]
  avg_total_detections: number[]
  max_total_detections: number[]
  sum_total_detections: number[]
  avg_latency_ms: number[]
}

export interface SummaryResponse {
  aggregated_summary: AggregatedSummaryRow[]
}

/** After `getClasses()`, `class_summary` is always normalized to a flat map (see `normalizeClassSummary`). */
export interface ClassesResponse {
  class_summary: Record<string, number>
}

export interface HealthResponse {
  health_summary: {
    camera: string[]
    total_detections: number[]
    latency_ms: number[]
    recorded_at: string[]
  }
}

export interface CameraStatusRow {
  camera: string
  totalDetections: number | null
  latencyMs: number | null
  recordedAt: string | null
}

export interface Point {
  x: number
  y: number
}

export interface SecurityZone {
  id: number
  name: string
  camera: string
  polygon: Point[]
  is_active: boolean
  active_from: string | null
  active_until: string | null
  loiter_threshold_seconds?: number | null
  default_severity?: string | null
  created_at?: string
  updated_at?: string
}

export interface ZoneListResponse {
  data: SecurityZone[]
}

export type AlertStatus = 'new' | 'acknowledged' | 'resolved'

export interface SecurityAlert {
  id: number
  camera: string
  zone_name: string
  zone_id: number | null
  tracker_id: number
  bound_box: [number, number, number, number]
  label: string
  status: AlertStatus
  /** Optional backend extension (migration 000008). */
  alert_type?: string | null
  severity?: string | null
  recorded_at: string
}
