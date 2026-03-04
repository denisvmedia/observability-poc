import axios from 'axios'

// --- Types ---

export interface UploadResult {
  rows_inserted: number
  rows_skipped: number
  errors: string[]
}

export interface KPIDimension {
  name: string
  version_a: number
  version_b: number
  winner: 'A' | 'B' | 'tie'
  lower_better: boolean
}

export interface Recommendation {
  winner: string
  wins_a: number
  wins_b: number
  dimensions: KPIDimension[]
  reason: string
}

export interface Alert {
  code: string
  message: string
}

export interface VersionKPIs {
  version: string
  session_count: number
  vsf_rate: number
  vpf_rate: number
  cirr_rate: number
  avg_vst: number
  play_rate: number
  completion_rate: number
}

export interface DashboardResponse {
  versions: { a: VersionKPIs; b: VersionKPIs }
  recommendation: Recommendation
  alerts: { a: Alert[]; b: Alert[] }
}

// --- API calls ---

const http = axios.create({ baseURL: '/api/v1' })

export async function uploadFile(file: File): Promise<UploadResult> {
  const form = new FormData()
  form.append('file', file)
  const { data } = await http.post<UploadResult>('/upload', form)
  return data
}

export async function listVersions(): Promise<string[]> {
  const { data } = await http.get<string[]>('/versions')
  return data
}

export async function getDashboard(v1: string, v2: string): Promise<DashboardResponse> {
  const { data } = await http.get<DashboardResponse>('/dashboard', { params: { v1, v2 } })
  return data
}

