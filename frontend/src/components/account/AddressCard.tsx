import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { CustomerAddress } from '@/types/account'
import { Edit, MapPin, Trash2 } from 'lucide-react'

interface AddressCardProps {
  address: CustomerAddress
  onEdit: (address: CustomerAddress) => void
  onDelete: (id: string) => void
  onSetDefault: (id: string) => void
  isUpdating: boolean
}

export function AddressCard({
  address,
  onEdit,
  onDelete,
  onSetDefault,
  isUpdating,
}: AddressCardProps) {
  return (
    <Card className={address.isDefault ? 'border-primary' : ''}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <MapPin className="h-4 w-4" />
            <CardTitle className="text-lg">Address</CardTitle>
            {address.isDefault && (
              <Badge variant="default" className="text-xs">
                Default
              </Badge>
            )}
          </div>
          <div className="flex gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onEdit(address)}
              disabled={isUpdating}
            >
              <Edit className="h-4 w-4" />
            </Button>
            {!address.isDefault && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onDelete(address.id)}
                disabled={isUpdating}
                className="text-destructive hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div>
          <p className="font-medium">{address.recipientName}</p>
          <p className="text-sm text-muted-foreground">
            {address.street}
            {address.apartment && `, ${address.apartment}`}
          </p>
          <p className="text-sm text-muted-foreground">
            {address.city}, {address.state} {address.postalCode}
          </p>
          <p className="text-sm text-muted-foreground">{address.country}</p>
        </div>

        {address.phone && (
          <div className="text-sm">
            <span className="font-medium">Phone:</span> {address.phone}
          </div>
        )}

        {address.instructions && (
          <div className="text-sm">
            <span className="font-medium">Instructions:</span>{' '}
            {address.instructions}
          </div>
        )}

        {!address.isDefault && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => onSetDefault(address.id)}
            disabled={isUpdating}
            className="w-full"
          >
            Set as Default
          </Button>
        )}
      </CardContent>
    </Card>
  )
}
