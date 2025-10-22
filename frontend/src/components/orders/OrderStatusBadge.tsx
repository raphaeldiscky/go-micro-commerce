import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { OrderStatus } from '@/types/__generated__/graphql'
import {
  CheckCircle2Icon,
  ClockIcon,
  PackageIcon,
  TruckIcon,
  XCircleIcon,
} from 'lucide-react'

interface StatusConfig {
  label: string
  variant: 'default' | 'secondary' | 'destructive' | 'outline'
  className: string
  icon: React.ComponentType<{ className?: string }>
}

const STATUS_CONFIGS: Record<OrderStatus, StatusConfig> = {
  PENDING: {
    label: 'Pending',
    variant: 'secondary',
    className:
      'bg-yellow-100 text-yellow-800 border-yellow-200 hover:bg-yellow-200',
    icon: ClockIcon,
  },
  PROCESSING: {
    label: 'Processing',
    variant: 'secondary',
    className: 'bg-blue-100 text-blue-800 border-blue-200 hover:bg-blue-200',
    icon: PackageIcon,
  },
  PAYMENT_PENDING: {
    label: 'Payment Pending',
    variant: 'secondary',
    className:
      'bg-orange-100 text-orange-800 border-orange-200 hover:bg-orange-200',
    icon: ClockIcon,
  },
  PAYMENT_EXPIRED: {
    label: 'Payment Expired',
    variant: 'destructive',
    className: 'bg-red-100 text-red-800 border-red-200 hover:bg-red-200',
    icon: XCircleIcon,
  },
  PAID: {
    label: 'Paid',
    variant: 'secondary',
    className:
      'bg-green-100 text-green-800 border-green-200 hover:bg-green-200',
    icon: CheckCircle2Icon,
  },
  SHIPPED: {
    label: 'Shipped',
    variant: 'secondary',
    className:
      'bg-purple-100 text-purple-800 border-purple-200 hover:bg-purple-200',
    icon: TruckIcon,
  },
  DELIVERED: {
    label: 'Delivered',
    variant: 'secondary',
    className: 'bg-teal-100 text-teal-800 border-teal-200 hover:bg-teal-200',
    icon: PackageIcon,
  },
  COMPLETED: {
    label: 'Completed',
    variant: 'default',
    className:
      'bg-emerald-100 text-emerald-800 border-emerald-200 hover:bg-emerald-200',
    icon: CheckCircle2Icon,
  },
  FAILED: {
    label: 'Failed',
    variant: 'destructive',
    className: 'bg-red-100 text-red-800 border-red-200 hover:bg-red-200',
    icon: XCircleIcon,
  },
  CANCELED: {
    label: 'Canceled',
    variant: 'outline',
    className: 'bg-gray-100 text-gray-800 border-gray-200 hover:bg-gray-200',
    icon: XCircleIcon,
  },
}

interface OrderStatusBadgeProps {
  status: OrderStatus
  className?: string
  showIcon?: boolean
  size?: 'sm' | 'default' | 'lg'
}

export function OrderStatusBadge({
  status,
  className,
  showIcon = true,
  size = 'default',
}: OrderStatusBadgeProps) {
  const config = STATUS_CONFIGS[status]
  const Icon = config.icon

  const sizeClasses = {
    sm: 'px-2 py-0.5 text-xs',
    default: 'px-2.5 py-1 text-sm',
    lg: 'px-3 py-1.5 text-base',
  }

  const iconSizes = {
    sm: 'h-3 w-3',
    default: 'h-4 w-4',
    lg: 'h-5 w-5',
  }

  return (
    <Badge
      variant={config.variant}
      className={cn(
        'inline-flex items-center gap-1.5 font-medium',
        config.className,
        sizeClasses[size],
        className,
      )}
    >
      {showIcon && <Icon className={cn(iconSizes[size])} />}
      {config.label}
    </Badge>
  )
}
