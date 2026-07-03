const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

export interface BookingInput {
  customer_name: string
  customer_email: string
  customer_phone?: string
  service: string
  starts_at: string // ISO 8601
  notes?: string
}

export async function bookAppointment(input: BookingInput): Promise<{ appointmentId: string }> {
  const res = await fetch(`${API_URL}/api/appointments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  const body = await res.json().catch(() => ({}))
  if (!res.ok || body.success === false) {
    throw new Error(body.error || `Booking failed (${res.status})`)
  }
  return { appointmentId: body.appointmentId }
}
