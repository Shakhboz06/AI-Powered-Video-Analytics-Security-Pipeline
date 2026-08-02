/**
 * Resolves a representative still image URL for a camera.
 *
 * There is no live frame endpoint yet, so this points at per-camera placeholder
 * imagery under `/public/camera-backgrounds/`. Components should fall back
 * gracefully (hide the image) when the asset is missing.
 */
function sanitizeCameraFilePart(name: string) {
  const s = name.replace(/[^a-zA-Z0-9_-]/g, '_').slice(0, 64)
  return s || 'camera'
}

export function cameraImageUrl(camera?: string | null): string {
  if (!camera) return '/camera-backgrounds/cam1.jpg'
  return `/camera-backgrounds/${sanitizeCameraFilePart(camera)}.jpg`
}
