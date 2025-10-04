import { APP_CONFIG, PATH_ROOT } from '@/constants'
import { getCurrentYear } from '@/lib/utils/date'
import { Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import React from 'react'
import { Button } from '../ui/button'

interface AuthLayoutProps {
  children: React.ReactNode
  showBackButton?: boolean
  title?: string
}

export function AuthLayout({
  children,
  showBackButton = true,
  title,
}: AuthLayoutProps) {
  return (
    <div className="min-h-screen flex flex-col bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
      {/* Header */}
      <header className="relative z-10 flex items-center justify-between p-4 md:p-6">
        {showBackButton ? (
          <Button asChild size="sm" variant="ghost">
            <Link className="flex items-center gap-2" to={PATH_ROOT.home}>
              <ArrowLeft className="h-4 w-4" />
              <span className="hidden sm:inline">Back to Home</span>
            </Link>
          </Button>
        ) : (
          <div />
        )}
        <Link
          className="flex items-center gap-2 hover:opacity-80 transition-opacity"
          to={PATH_ROOT.home}
        >
          <span className="font-bold text-lg">{APP_CONFIG.NAME}</span>
        </Link>
        <div className="w-[88px]" /> {/* Spacer for center alignment */}
      </header>

      {/* Main Content */}
      <main className="flex-1 flex items-center justify-center p-4 relative">
        {/* Background decoration */}
        <div className="absolute inset-0 -z-10 h-full w-full bg-white dark:bg-gray-900 [background:radial-gradient(125%_125%_at_50%_10%,#fff_40%,#63e_100%)] dark:[background:radial-gradient(125%_125%_at_50%_10%,#000_40%,#63e_100%)] opacity-20" />

        <div className="w-full max-w-md">
          {title && (
            <div className="text-center mb-8">
              <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
                {title}
              </h1>
            </div>
          )}
          {children}
        </div>
      </main>

      {/* Footer */}
      <footer className="relative z-10 py-6 text-center text-sm text-muted-foreground">
        <p>
          &copy; {getCurrentYear()} {APP_CONFIG.NAME}. Built for educational
          purposes.
        </p>
      </footer>
    </div>
  )
}
