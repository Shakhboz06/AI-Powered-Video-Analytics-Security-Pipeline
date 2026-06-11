/**
 * Central visual metadata for security alert types.
 */

export interface AlertTypeMeta {
  key: string
  label: string
  iconPaths: string[]
  badgeClass: string
  stripeClass: string
  accentText: string
  boxColor: string
  sceneLevel?: boolean
}

const ICONS = {
  fire: [
    'M15.362 5.214A8.252 8.252 0 0 1 12 21 8.25 8.25 0 0 1 6.038 7.047 8.287 8.287 0 0 0 9 9.601a8.983 8.983 0 0 1 3.361-6.867 8.21 8.21 0 0 0 3 2.48Z',
    'M12 18a3.75 3.75 0 0 0 .495-7.467 5.99 5.99 0 0 0-1.925 3.546 5.974 5.974 0 0 1-2.133-1.001A3.75 3.75 0 0 0 12 18Z',
  ],
  shield: [
    'M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z',
  ],
  bolt: ['M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z'],
  warning: [
    'M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z',
  ],
  trendingDown: ['M2.25 6 9 12.75l4.306-4.307a11.95 11.95 0 0 1 5.814 5.519l.93 1.95m0 0 .61-4.288m-.61 4.287-4.287-.61'],
  archive: [
    'M20.25 7.5l-.625 10.632a2.25 2.25 0 0 1-2.247 2.118H6.622a2.25 2.25 0 0 1-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125Z',
  ],
  exclaimCircle: ['M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z'],
  clock: ['M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z'],
} as const

const DEFAULT_META: AlertTypeMeta = {
  key: 'alert',
  label: 'Alert',
  iconPaths: ICONS.warning,
  badgeClass: 'border-gray-600 bg-gray-900/60 text-gray-300',
  stripeClass: 'bg-gray-500',
  accentText: 'text-gray-300',
  boxColor: '#9ca3af',
}

const ALERT_META: Record<string, AlertTypeMeta> = {
  fighting: {
    key: 'fighting',
    label: 'Fighting',
    iconPaths: ICONS.fire,
    badgeClass: 'border-orange-700/60 bg-orange-950/40 text-orange-200',
    stripeClass: 'bg-orange-500',
    accentText: 'text-orange-300',
    boxColor: '#fb923c',
    sceneLevel: true,
  },
  falling: {
    key: 'falling',
    label: 'Fall detected',
    iconPaths: ICONS.trendingDown,
    badgeClass: 'border-amber-700/60 bg-amber-950/40 text-amber-200',
    stripeClass: 'bg-amber-500',
    accentText: 'text-amber-300',
    boxColor: '#fbbf24',
  },
  intrusion: {
    key: 'intrusion',
    label: 'Zone intrusion',
    iconPaths: ICONS.shield,
    badgeClass: 'border-red-800/60 bg-red-950/40 text-red-200',
    stripeClass: 'bg-red-500',
    accentText: 'text-red-300',
    boxColor: '#f87171',
  },
  running: {
    key: 'running',
    label: 'Running',
    iconPaths: ICONS.bolt,
    badgeClass: 'border-sky-800/60 bg-sky-950/40 text-sky-200',
    stripeClass: 'bg-sky-500',
    accentText: 'text-sky-300',
    boxColor: '#38bdf8',
  },
  brandishing: {
    key: 'brandishing',
    label: 'Weapon brandishing',
    iconPaths: ICONS.exclaimCircle,
    badgeClass: 'border-rose-800/60 bg-rose-950/40 text-rose-200',
    stripeClass: 'bg-rose-500',
    accentText: 'text-rose-300',
    boxColor: '#fb7185',
  },
  abandoned_object: {
    key: 'abandoned_object',
    label: 'Abandoned object',
    iconPaths: ICONS.archive,
    badgeClass: 'border-violet-800/60 bg-violet-950/40 text-violet-200',
    stripeClass: 'bg-violet-500',
    accentText: 'text-violet-300',
    boxColor: '#a78bfa',
  },
  loitering: {
    key: 'loitering',
    label: 'Loitering',
    iconPaths: ICONS.clock,
    badgeClass: 'border-teal-800/60 bg-teal-950/40 text-teal-200',
    stripeClass: 'bg-teal-500',
    accentText: 'text-teal-300',
    boxColor: '#2dd4bf',
  },
}

const ALIASES: Record<string, string> = {
  abandoned_obj: 'abandoned_object',
  fall: 'falling',
  fall_detected: 'falling',
  fight: 'fighting',
}

export function normalizeAlertType(type?: string | null): string {
  if (!type) return ''
  const key = type.trim().toLowerCase()
  return ALIASES[key] ?? key
}

export function getAlertMeta(type?: string | null): AlertTypeMeta {
  const key = normalizeAlertType(type)
  return ALERT_META[key] ?? { ...DEFAULT_META, label: prettyFallback(type) }
}

function prettyFallback(type?: string | null): string {
  if (!type) return 'Alert'
  return type.replace(/[_-]+/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

export const ALERT_TYPE_FILTER_OPTIONS = [
  { value: 'all', label: 'All types' },
  { value: 'fighting', label: 'Fighting' },
  { value: 'falling', label: 'Fall detected' },
  { value: 'intrusion', label: 'Zone intrusion' },
  { value: 'running', label: 'Running' },
  { value: 'brandishing', label: 'Weapon brandishing' },
  { value: 'abandoned_object', label: 'Abandoned object' },
  { value: 'loitering', label: 'Loitering' },
] as const
