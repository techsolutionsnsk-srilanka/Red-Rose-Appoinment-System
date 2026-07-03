import Link from 'next/link'

export default function Home() {
  return (
    <main className="mx-auto flex min-h-screen max-w-3xl flex-col items-center justify-center px-6 text-center">
      <div className="text-6xl">🌹</div>
      <h1 className="mt-4 text-4xl font-bold tracking-tight">RedRose Appointments</h1>
      <p className="mt-3 max-w-md text-muted-foreground">
        Book your appointment online in under a minute. Choose a service, pick a
        time, and we&apos;ll email you a confirmation.
      </p>
      <Link
        href="/book"
        className="mt-8 inline-flex items-center rounded-lg bg-primary px-6 py-3 font-semibold text-primary-foreground transition-opacity hover:opacity-90"
      >
        Book an appointment →
      </Link>
    </main>
  )
}
