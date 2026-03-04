import type { VersionKPIs } from '@/services/api'

// Each metric is normalised to a 0-100 "quality score" so that
// higher always means better on the radar chart.
//
// For lower-is-better metrics the normalisation is:
//   score = clamp(100 * (1 - value / threshold), 0, 100)
// For higher-is-better metrics (already 0-1 ratios):
//   score = value * 100

function clamp(v: number, lo: number, hi: number): number {
  return Math.min(hi, Math.max(lo, v))
}

export function toQualityScores(kpis: VersionKPIs): number[] {
  return [
    clamp(100 * (1 - kpis.vsf_rate / 0.05), 0, 100), // VSF Rate   threshold 5%
    clamp(100 * (1 - kpis.vpf_rate / 0.05), 0, 100), // VPF Rate   threshold 5%
    clamp(100 * (1 - kpis.cirr_rate / 0.10), 0, 100), // CIRR       threshold 10%
    clamp(100 * (1 - kpis.avg_vst / 5.0), 0, 100), // Avg VST    threshold 5s
    kpis.play_rate * 100, // Play Rate  higher is better
    kpis.completion_rate * 100, // Completion higher is better
  ]
}

export const RADAR_LABELS = ['VSF Rate', 'VPF Rate', 'CIRR', 'Avg VST', 'Play Rate', 'Completion Rate']

