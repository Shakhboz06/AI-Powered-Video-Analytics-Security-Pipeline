/**
 * Landing-page "Detected in real footage" gallery data.
 *
 * Every entry is taken straight from the demo Metadata/*.txt files, so the
 * cards show real detections. Images are local static assets under
 * `public/detections/`. Edit this list freely — DetectionGallery renders it.
 *
 * Fields (all present in the source metadata):
 *  - image:     path under /public (e.g. '/detections/falling-1.jpg')
 *  - type:      detection kind — drives the badge icon/colour via useAlertMeta
 *               ('brandishing' | 'fighting' | 'falling' | 'abandoned_object'
 *                | 'running' | 'intrusion' | 'loitering')
 *  - label:     display text on the badge
 *  - object:    what was detected (metadata "Object", e.g. 'person', 'suitcase')
 *  - camera:    source camera (metadata "Camera")
 *  - timestamp: capture time (metadata "Recorded at", time-of-day)
 *  - severity:  'critical' | 'high' | 'medium' | 'low' — drives colour coding
 */
export interface DetectionCard {
  image: string
  type: string
  label: string
  object: string
  camera: string
  timestamp: string
  severity: 'critical' | 'high' | 'medium' | 'low'
}

export const detectionCards: DetectionCard[] = [
  { image: '/detections/weapon-brandishing-1.jpg', type: 'brandishing', label: 'Weapon brandishing', object: 'person', camera: 'yard', timestamp: '9:42:31 PM', severity: 'critical' },
  { image: '/detections/fighting-1.jpg', type: 'fighting', label: 'Fight', object: 'Group of people', camera: 'warehouse', timestamp: '10:34:13 PM', severity: 'critical' },
  { image: '/detections/falling-2.jpg', type: 'falling', label: 'Fall', object: 'person', camera: 'room 2', timestamp: '8:23:14 PM', severity: 'critical' },
  { image: '/detections/running-1.jpg', type: 'running', label: 'Running', object: 'person', camera: 'main square', timestamp: '10:55:51 PM', severity: 'medium' },
  { image: '/detections/abandoned-object-1.jpg', type: 'abandoned_object', label: 'Abandoned object', object: 'suitcase', camera: 'living room 1', timestamp: '9:26:07 PM', severity: 'medium' },
  { image: '/detections/weapon-brandishing-2.jpg', type: 'brandishing', label: 'Weapon brandishing', object: 'person', camera: 'yard', timestamp: '9:42:46 PM', severity: 'critical' },
  { image: '/detections/falling-1.jpg', type: 'falling', label: 'Fall', object: 'person', camera: 'cam 1', timestamp: '7:35:52 PM', severity: 'medium' },
  { image: '/detections/fighting-2.jpg', type: 'fighting', label: 'Fight', object: 'Group of people', camera: 'shop', timestamp: '10:47:39 PM', severity: 'critical' },
  { image: '/detections/weapon-brandishing-3.jpg', type: 'brandishing', label: 'Weapon brandishing', object: 'person', camera: 'house garden', timestamp: '9:50:54 PM', severity: 'critical' },
  { image: '/detections/falling-5.jpg', type: 'falling', label: 'Fall', object: 'person', camera: 'lobby', timestamp: '10:42:44 PM', severity: 'critical' },
  { image: '/detections/falling-4.jpg', type: 'falling', label: 'Fall', object: 'person', camera: 'cam 2', timestamp: '8:17:30 PM', severity: 'critical' },
  { image: '/detections/weapon-brandishing-4.jpg', type: 'brandishing', label: 'Weapon brandishing', object: 'person', camera: 'kitchen', timestamp: '10:29:16 PM', severity: 'critical' },
  { image: '/detections/falling-3.jpg', type: 'falling', label: 'Fall', object: 'person', camera: 'office 2', timestamp: '8:35:20 PM', severity: 'medium' },
]
