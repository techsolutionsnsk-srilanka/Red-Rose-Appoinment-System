'use client'

import { useState } from 'react'
import Link from 'next/link'
import { toast } from 'sonner'
import { bookAppointment } from '@/lib/api'

const SERVICES = ['Consultation', 'Follow-up', 'Full session', 'Assessment']

export default function BookPage() {
  const [submitting, setSubmitting] = useState(false)
  const [done, setDone] = useState<string | null>(null)

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = e.currentTarget
    const data = new FormData(form)
    const date = String(data.get('date'))
    const time = String(data.get('time'))

    setSubmitting(true)
    try {
      const { appointmentId } = await bookAppointment({
        customer_name: String(data.get('name')),
        customer_email: String(data.get('email')),
        customer_phone: String(data.get('phone') || ''),
        service: String(data.get('service')),
        starts_at: new Date(`${date}T${time}`).toISOString(),
        notes: String(data.get('notes') || ''),
      })
      setDone(appointmentId)
      toast.success('Appointment booked! Check your email.')
      form.reset()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Booking failed')
    } finally {
      setSubmitting(false)
    }
  }

  const label = 'block text-sm font-medium text-muted-foreground mb-1'
  const input =
    'w-full rounded-lg border border-border bg-card px-3 py-2 outline-none focus:ring-2 focus:ring-primary'

  return (
    <main className="mx-auto max-w-lg px-6 py-12">
      <Link href="/" className="text-sm text-muted-foreground hover:underline">
        ← Back
      </Link>
      <h1 className="mt-3 text-3xl font-bold">Book an appointment</h1>

      {done ? (
        <div className="mt-8 rounded-xl border border-border bg-card p-6">
          <p className="text-lg font-semibold">🎉 You&apos;re booked!</p>
          <p className="mt-2 text-muted-foreground">
            Reference: <code className="font-mono">{done}</code>. We&apos;ve emailed
            your confirmation.
          </p>
          <button
            onClick={() => setDone(null)}
            className="mt-4 rounded-lg bg-primary px-4 py-2 font-semibold text-primary-foreground hover:opacity-90"
          >
            Book another
          </button>
        </div>
      ) : (
        <form onSubmit={onSubmit} className="mt-8 space-y-4">
          <div>
            <label className={label}>Full name</label>
            <input name="name" required className={input} />
          </div>
          <div>
            <label className={label}>Email</label>
            <input name="email" type="email" required className={input} />
          </div>
          <div>
            <label className={label}>Phone (optional)</label>
            <input name="phone" className={input} />
          </div>
          <div>
            <label className={label}>Service</label>
            <select name="service" required className={input}>
              {SERVICES.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </div>
          <div className="flex gap-4">
            <div className="flex-1">
              <label className={label}>Date</label>
              <input name="date" type="date" required className={input} />
            </div>
            <div className="flex-1">
              <label className={label}>Time</label>
              <input name="time" type="time" required className={input} />
            </div>
          </div>
          <div>
            <label className={label}>Notes (optional)</label>
            <textarea name="notes" rows={3} className={input} />
          </div>
          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-lg bg-primary px-4 py-3 font-semibold text-primary-foreground hover:opacity-90 disabled:opacity-50"
          >
            {submitting ? 'Booking…' : 'Confirm booking'}
          </button>
        </form>
      )}
    </main>
  )
}
