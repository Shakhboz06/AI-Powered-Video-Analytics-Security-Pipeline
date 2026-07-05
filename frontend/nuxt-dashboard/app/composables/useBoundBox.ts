/** Ingestor resizes every frame to 640×480 before detection (see ingestor/ingestor.py). */
export const PIPELINE_FRAME_WIDTH = 640
export const PIPELINE_FRAME_HEIGHT = 480

function toNumbers(raw: unknown): number[] | null {
  if (raw == null) return null
  if (Array.isArray(raw)) {
    if (raw.length >= 4 && raw.every((v) => typeof v === 'number' || typeof v === 'string')) {
      const nums = raw.slice(0, 4).map((v) => (typeof v === 'number' ? v : Number(v)))
      if (!nums.some((n) => Number.isNaN(n))) return nums
    }
    if (raw.length > 0) return toNumbers(raw[0])
  }
  return null
}

/** Normalize backend boxes to pixel xyxy in the pipeline frame (640×480). */
export function normalizeBoundBox(
  raw: unknown,
  frameW = PIPELINE_FRAME_WIDTH,
  frameH = PIPELINE_FRAME_HEIGHT,
): [number, number, number, number] | null {
  const nums = toNumbers(raw)
  if (!nums) return null

  let [a, b, c, d] = nums

  // Normalized 0–1 coordinates (all values within [0, 1])
  if (a >= 0 && b >= 0 && c >= 0 && d >= 0 && a <= 1 && b <= 1 && c <= 1 && d <= 1) {
    a *= frameW
    b *= frameH
    c *= frameW
    d *= frameH
  }

  let x1: number
  let y1: number
  let x2: number
  let y2: number

  if (c > a && d > b) {
    x1 = a
    y1 = b
    x2 = c
    y2 = d
  }
  else {
    x1 = a
    y1 = b
    x2 = a + Math.abs(c)
    y2 = b + Math.abs(d)
  }

  x1 = Math.max(0, Math.min(x1, frameW))
  y1 = Math.max(0, Math.min(y1, frameH))
  x2 = Math.max(0, Math.min(x2, frameW))
  y2 = Math.max(0, Math.min(y2, frameH))

  if (x2 - x1 <= 0 || y2 - y1 <= 0) return null
  return [x1, y1, x2, y2]
}

export function hasValidBoundBox(
  raw: unknown,
  frameW = PIPELINE_FRAME_WIDTH,
  frameH = PIPELINE_FRAME_HEIGHT,
): boolean {
  return normalizeBoundBox(raw, frameW, frameH) !== null
}
