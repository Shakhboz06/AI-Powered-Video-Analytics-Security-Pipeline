/**
 * Landing-page "Detected in real footage" gallery data.
 *
 * Edit this list freely — the DetectionGallery component renders whatever is
 * here (3–6 cards read best). Images are local static assets under
 * `public/detections/`; reference them with a root-absolute path.
 *
 * Fields:
 *  - image:      path under /public (e.g. '/detections/fall.jpg')
 *  - type:       detection kind — drives the badge icon/colour via useAlertMeta
 *                ('brandishing' | 'fighting' | 'falling' | 'abandoned_object'
 *                 | 'running' | 'intrusion' | 'loitering')
 *  - label:      display text on the badge (overrides the type's default label)
 *  - timestamp:  where in the clip it was detected (free text, e.g. '00:42')
 *  - severity:   'critical' | 'high' | 'medium' | 'low' — drives colour coding
 *  - confidence: model confidence 0–1 (shown as a percentage)
 */
export interface DetectionCard {
  image: string
  type: string
  label: string
  timestamp: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  confidence: number
}

export const detectionCards: DetectionCard[] = [
  {
    image: '/detections/weapon-brandishing.jpg',
    type: 'brandishing',
    label: 'Weapon brandishing',
    timestamp: '00:12',
    severity: 'critical',
    confidence: 0.96,
  },
  {
    image: '/detections/fighting.jpg',
    type: 'fighting',
    label: 'Fight',
    timestamp: '00:34',
    severity: 'critical',
    confidence: 0.91,
  },
  {
    image: '/detections/fall.jpg',
    type: 'falling',
    label: 'Fall',
    timestamp: '00:07',
    severity: 'medium',
    confidence: 0.88,
  },
  {
    image: '/detections/abandoned-object.jpg',
    type: 'abandoned_object',
    label: 'Abandoned object',
    timestamp: '01:26',
    severity: 'medium',
    confidence: 0.83,
  },
  {
    image: '/detections/weapon-brandishing-2.jpg',
    type: 'brandishing',
    label: 'Weapon brandishing',
    timestamp: '00:58',
    severity: 'critical',
    confidence: 0.94,
  },
  {
    image: '/detections/fall-2.jpg',
    type: 'falling',
    label: 'Fall',
    timestamp: '00:21',
    severity: 'medium',
    confidence: 0.86,
  },
]
