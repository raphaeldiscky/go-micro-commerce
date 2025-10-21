import { Link } from '@tanstack/react-router'
import { ChevronRight } from 'lucide-react'

interface BreadcrumbItem {
  label: string
  href?: string
}

interface DashboardHeaderProps {
  title: string
  breadcrumbs?: Array<BreadcrumbItem>
  action?: React.ReactNode
}

export function DashboardHeader({
  action,
  breadcrumbs,
  title,
}: DashboardHeaderProps) {
  return (
    <div className="border-b bg-card">
      <div className="flex items-center justify-between px-6 py-4">
        <div className="space-y-1">
          {breadcrumbs && breadcrumbs.length > 0 && (
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              {breadcrumbs.map((item, index) => (
                <div key={index} className="flex items-center gap-1">
                  {item.href ? (
                    <Link to={item.href} className="hover:text-foreground">
                      {item.label}
                    </Link>
                  ) : (
                    <span>{item.label}</span>
                  )}
                  {index < breadcrumbs.length - 1 && (
                    <ChevronRight className="h-4 w-4" />
                  )}
                </div>
              ))}
            </div>
          )}
          <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
        </div>
        {action && <div>{action}</div>}
      </div>
    </div>
  )
}
