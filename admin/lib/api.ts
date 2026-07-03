import { getAuthHeaders } from './auth'

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

export type AppointmentStatus =
  | 'pending'
  | 'confirmed'
  | 'cancelled'
  | 'completed'

export interface Appointment {
  id: string
  customer_name: string
  customer_email: string
  customer_phone?: string
  service: string
  starts_at: string
  duration_min: number
  status: AppointmentStatus
  notes?: string
  admin_notes?: string
  created_by?: string
  created_at: string
  updated_at: string
}

export interface AppointmentStats {
  total_appointments: number
  pending_appointments: number
  confirmed_appointments: number
  completed_appointments: number
  upcoming_today: number
}

interface ApiEnvelope<T> {
  success: boolean
  data?: T
  error?: string
  count?: number
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = await getAuthHeaders()
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...headers,
      ...(init?.headers ?? {}),
    },
  })

  const body = (await res.json().catch(() => ({}))) as ApiEnvelope<T>
  if (!res.ok || body.success === false) {
    throw new Error(body.error || `Request failed (${res.status})`)
  }
  return body.data as T
}

export const api = {
  listAppointments: () => request<Appointment[]>('/api/appointments'),
  getStats: () => request<AppointmentStats>('/api/appointments/stats'),
  updateStatus: (id: string, status: AppointmentStatus) =>
    request<null>(`/api/appointments/${id}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    }),
  deleteAppointment: (id: string) =>
    request<null>(`/api/appointments/${id}`, { method: 'DELETE' }),
}
