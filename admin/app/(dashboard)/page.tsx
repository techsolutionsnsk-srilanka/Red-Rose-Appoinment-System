'use client'

import { useEffect, useState } from 'react'
import { api, type AppointmentStats } from '@/lib/api'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'

const cards: { key: keyof AppointmentStats; label: string }[] = [
  { key: 'total_appointments', label: 'Total' },
  { key: 'upcoming_today', label: 'Upcoming today' },
  { key: 'pending_appointments', label: 'Pending' },
  { key: 'confirmed_appointments', label: 'Confirmed' },
  { key: 'completed_appointments', label: 'Completed' },
]

export default function DashboardPage() {
  const [stats, setStats] = useState<AppointmentStats | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api
      .getStats()
      .then(setStats)
      .catch((e) => setError(e.message))
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">Overview of your appointments</p>
      </div>

      {error && (
        <Card>
          <CardContent className="pt-6 text-sm text-primary">
            Could not reach the API: {error}
          </CardContent>
        </Card>
      )}

      <div className="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-5">
        {cards.map(({ key, label }) => (
          <Card key={key}>
            <CardHeader>
              <CardTitle className="text-sm text-muted-foreground">{label}</CardTitle>
              <p className="text-3xl font-bold">{stats ? stats[key] : '—'}</p>
            </CardHeader>
          </Card>
        ))}
      </div>
    </div>
  )
}
