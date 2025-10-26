import { Badge } from '@/components/ui/badge'
import type { PaymentStatus } from '@/types/__generated__/graphql'
import {
  AlertCircle,
  CheckCircle2,
  Clock,
  Loader2,
  XCircle,
} from 'lucide-react'

interface PaymentStatusBadgeProps {
  status: PaymentStatus
  className?: string
}

/**
 * Visual badge component for payment status
 * Shows appropriate icon and color for each status
 */
export function PaymentStatusBadge({
  status,
  className,
}: PaymentStatusBadgeProps) {
  const statusConfig: Record<
    PaymentStatus,
    {
      label: string
      variant: 'default' | 'destructive' | 'secondary'
      icon: typeof Clock
      className?: string
      iconClassName?: string
    }
  > = {
    PENDING: {
      label: 'Pending',
      variant: 'secondary' as const,
      icon: Clock,
    },
    PROCESSING: {
      label: 'Processing',
      variant: 'default' as const,
      icon: Loader2,
      iconClassName: 'animate-spin',
    },
    COMPLETED: {
      label: 'Completed',
      variant: 'default' as const,
      icon: CheckCircle2,
      className: 'bg-green-500 hover:bg-green-600',
    },
    FAILED: {
      label: 'Failed',
      variant: 'destructive' as const,
      icon: XCircle,
    },
    TIMEOUT: {
      label: 'Expired',
      variant: 'destructive' as const,
      icon: AlertCircle,
    },
    REFUNDED: {
      label: 'Refunded',
      variant: 'secondary' as const,
      icon: AlertCircle,
    },
  }

  const config = statusConfig[status]
  const Icon = config.icon

  return (
    <Badge
      variant={config.variant}
      className={`${config.className || ''} ${className || ''}`}
    >
      <Icon className={`mr-1 h-3 w-3 ${config.iconClassName || ''}`} />
      {config.label}
    </Badge>
  )
}
