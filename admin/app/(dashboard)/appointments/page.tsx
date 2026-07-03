'use client'

import { useCallback, useEffect, useState } from 'react'
import { format } from 'date-fns'
import { toast } from 'sonner'
import { Trash2 } from 'lucide-react'
import { api, type Appointment, type AppointmentStatus } from '@/lib/api'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { Select } from '@/components/ui/select'
import { Button } from '@/components/ui/button'

const STATUSES: AppointmentStatus[] = ['pending', 'confirmed', 'cancelled', 'completed']

export default function AppointmentsPage() {
  const [rows, setRows] = useState<Appointment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      setRows(await api.listAppointments())
      setError(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  async function changeStatus(id: string, status: AppointmentStatus) {
    try {
      await api.updateStatus(id, status)
      setRows((prev) => prev.map((r) => (r.id === id ? { ...r, status } : r)))
      toast.success(`Marked ${status}`)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Update failed')
    }
  }

  async function remove(id: string) {
    try {
      await api.deleteAppointment(id)
      setRows((prev) => prev.filter((r) => r.id !== id))
      toast.success('Deleted')
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Appointments</h1>
        <p className="text-muted-foreground">Manage customer bookings</p>
      </div>

      <Card>
        <CardContent className="pt-6">
          {loading && <p className="text-muted-foreground">Loading…</p>}
          {error && <p className="text-primary">Could not reach the API: {error}</p>}
          {!loading && !error && rows.length === 0 && (
            <p className="text-muted-foreground">No appointments yet.</p>
          )}
          {rows.length > 0 && (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Customer</TableHead>
                  <TableHead>Service</TableHead>
                  <TableHead>When</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((a) => (
                  <TableRow key={a.id}>
                    <TableCell>
                      <div className="font-medium">{a.customer_name}</div>
                      <div className="text-xs text-muted-foreground">{a.customer_email}</div>
                    </TableCell>
                    <TableCell>{a.service}</TableCell>
                    <TableCell>{format(new Date(a.starts_at), 'PP p')}</TableCell>
                    <TableCell>
                      <Select
                        value={a.status}
                        onChange={(e) => changeStatus(a.id, e.target.value as AppointmentStatus)}
                      >
                        {STATUSES.map((s) => (
                          <option key={s} value={s}>
                            {s}
                          </option>
                        ))}
                      </Select>
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => remove(a.id)}
                        aria-label="Delete"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
