/** Consistent ApexCharts palette for security dashboard (dark SOC theme). */
export function useChartTheme() {
  const seriesColors = ['#22d3ee', '#3b82f6', '#a78bfa', '#f472b6', '#34d399', '#fbbf24', '#fb923c', '#f87171']

  const base = {
    chart: {
      fontFamily: 'Outfit, sans-serif',
      foreColor: '#9ca3af',
      background: 'transparent',
      toolbar: { show: false },
      zoom: { enabled: false },
    },
    theme: { mode: 'dark' as const },
    grid: {
      borderColor: '#1f2937',
      strokeDashArray: 4,
    },
    legend: {
      labels: { colors: '#9ca3af' },
    },
    dataLabels: { enabled: false },
    stroke: { curve: 'smooth' as const, width: 2 },
  }

  return { seriesColors, base }
}

export function latencyColorClass(ms: number): string {
  if (ms < 300) return 'text-emerald-400'
  if (ms <= 500) return 'text-amber-400'
  return 'text-red-400'
}

export function latencyBgClass(ms: number): string {
  if (ms < 300) return 'bg-emerald-500/15 border-emerald-500/40'
  if (ms <= 500) return 'bg-amber-500/15 border-amber-500/40'
  return 'bg-red-500/15 border-red-500/40'
}

/**
 * ApexCharts axis labels are often passed as strings with float noise (e.g. 1000.0000000000000000).
 * Use for ms, counts, and other numeric axes.
 */
export function formatAxisInteger(val: string | number): string {
  const n = typeof val === 'number' ? val : Number.parseFloat(String(val).replace(/,/g, ''))
  if (!Number.isFinite(n)) return String(val)
  return Math.round(n).toLocaleString(undefined, { maximumFractionDigits: 0 })
}
