import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { useAccountStore } from '@/store/accountStore'
import type { AddressRequest, CustomerAddress } from '@/types/account'
import { useForm } from '@tanstack/react-form'
import { Edit, MapPin, Plus, Trash2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'

interface AddressCardProps {
  address: CustomerAddress
  onEdit: (address: CustomerAddress) => void
  onDelete: (id: string) => void
  onSetDefault: (id: string) => void
  isUpdating: boolean
}

function AddressCard({
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

interface AddressFormProps {
  address?: CustomerAddress
  onSubmit: (data: AddressRequest) => void
  onCancel: () => void
  isSubmitting: boolean
}

function AddressForm({
  address,
  onSubmit,
  onCancel,
  isSubmitting,
}: AddressFormProps) {
  const form = useForm({
    defaultValues: {
      recipientName: address?.recipientName || '',
      street: address?.street || '',
      apartment: address?.apartment || '',
      city: address?.city || '',
      state: address?.state || '',
      postalCode: address?.postalCode || '',
      country: address?.country || 'United States',
      phone: address?.phone || '',
      instructions: address?.instructions || '',
      isDefault: address?.isDefault || false,
    },
    onSubmit: ({ value }) => {
      onSubmit(value)
    },
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        e.stopPropagation()
        form.handleSubmit()
      }}
      className="space-y-4"
    >
      <form.Field name="recipientName">
        {(field) => (
          <div className="space-y-2">
            <Label htmlFor="recipientName">Recipient Name</Label>
            <Input
              id="recipientName"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              placeholder="Full name"
            />
          </div>
        )}
      </form.Field>

      <form.Field name="street">
        {(field) => (
          <div className="space-y-2">
            <Label htmlFor="street">Street Address</Label>
            <Input
              id="street"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              placeholder="123 Main Street"
            />
          </div>
        )}
      </form.Field>

      <form.Field name="apartment">
        {(field) => (
          <div className="space-y-2">
            <Label htmlFor="apartment">Apartment, Suite, Unit (Optional)</Label>
            <Input
              id="apartment"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              placeholder="Apt 4B, Suite 200"
            />
          </div>
        )}
      </form.Field>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <form.Field name="city">
          {(field) => (
            <div className="space-y-2">
              <Label htmlFor="city">City</Label>
              <Input
                id="city"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                placeholder="New York"
              />
            </div>
          )}
        </form.Field>

        <form.Field name="state">
          {(field) => (
            <div className="space-y-2">
              <Label htmlFor="state">State/Province</Label>
              <Input
                id="state"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                placeholder="NY"
              />
            </div>
          )}
        </form.Field>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <form.Field name="postalCode">
          {(field) => (
            <div className="space-y-2">
              <Label htmlFor="postalCode">Postal Code</Label>
              <Input
                id="postalCode"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                placeholder="10001"
              />
            </div>
          )}
        </form.Field>

        <form.Field name="country">
          {(field) => (
            <div className="space-y-2">
              <Label htmlFor="country">Country</Label>
              <Select
                value={field.state.value}
                onValueChange={(value) => field.handleChange(value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select country" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="United States">United States</SelectItem>
                  <SelectItem value="Canada">Canada</SelectItem>
                  <SelectItem value="United Kingdom">United Kingdom</SelectItem>
                  <SelectItem value="Australia">Australia</SelectItem>
                  <SelectItem value="Germany">Germany</SelectItem>
                  <SelectItem value="France">France</SelectItem>
                  <SelectItem value="Japan">Japan</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}
        </form.Field>
      </div>

      <form.Field name="phone">
        {(field) => (
          <div className="space-y-2">
            <Label htmlFor="phone">Phone Number (Optional)</Label>
            <Input
              id="phone"
              type="tel"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              placeholder="+1 (555) 123-4567"
            />
          </div>
        )}
      </form.Field>

      <form.Field name="instructions">
        {(field) => (
          <div className="space-y-2">
            <Label htmlFor="instructions">
              Delivery Instructions (Optional)
            </Label>
            <Textarea
              id="instructions"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              placeholder="Ring doorbell twice, leave at front door, etc."
              rows={3}
            />
          </div>
        )}
      </form.Field>

      <form.Field name="isDefault">
        {(field) => (
          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="isDefault"
              checked={field.state.value}
              onChange={(e) => field.handleChange(e.target.checked)}
              className="rounded border-gray-300"
            />
            <Label htmlFor="isDefault">Set as default address</Label>
          </div>
        )}
      </form.Field>

      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isSubmitting}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting
            ? 'Saving...'
            : address
              ? 'Update Address'
              : 'Add Address'}
        </Button>
      </DialogFooter>
    </form>
  )
}

export function AddressSection() {
  const addresses = useAccountStore((state) => state.addresses)
  const addAddress = useAccountStore((state) => state.addAddress)
  const updateAddress = useAccountStore((state) => state.updateAddress)
  const deleteAddress = useAccountStore((state) => state.deleteAddress)
  const setDefaultAddress = useAccountStore((state) => state.setDefaultAddress)
  const loadAddresses = useAccountStore((state) => state.loadAddresses)
  const isLoading = useAccountStore((state) => state.isLoading)
  const isUpdating = useAccountStore((state) => state.isUpdating)

  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
  const [editingAddress, setEditingAddress] = useState<CustomerAddress | null>(
    null,
  )

  // Load addresses on component mount
  useEffect(() => {
    if (addresses.length === 0) {
      loadAddresses()
    }
  }, [addresses.length, loadAddresses])

  const handleAddAddress = async (data: AddressRequest) => {
    try {
      await addAddress(data)
      setIsAddDialogOpen(false)
      toast.success('Address added successfully')
    } catch (error) {
      console.error('Failed to add address:', error)
      toast.error('Failed to add address')
    }
  }

  const handleUpdateAddress = async (data: AddressRequest) => {
    if (!editingAddress) return

    try {
      await updateAddress(editingAddress.id, data)
      setEditingAddress(null)
      toast.success('Address updated successfully')
    } catch (error) {
      console.error('Failed to update address:', error)
      toast.error('Failed to update address')
    }
  }

  const handleDeleteAddress = async (id: string) => {
    if (confirm('Are you sure you want to delete this address?')) {
      try {
        await deleteAddress(id)
        toast.success('Address deleted successfully')
      } catch (error) {
        console.error('Failed to delete address:', error)
        toast.error('Failed to delete address')
      }
    }
  }

  const handleSetDefault = async (id: string) => {
    try {
      await setDefaultAddress(id)
      toast.success('Default address updated successfully')
    } catch (error) {
      console.error('Failed to set default address:', error)
      toast.error('Failed to set default address')
    }
  }

  if (isLoading && addresses.length === 0) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center h-32">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div>
            <p className="text-muted-foreground">Loading addresses...</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <MapPin className="h-5 w-5" />
                Shipping Addresses
              </CardTitle>
              <CardDescription>
                Manage your shipping addresses for faster checkout
              </CardDescription>
            </div>
            <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="h-4 w-4 mr-2" />
                  Add Address
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle>Add New Address</DialogTitle>
                  <DialogDescription>
                    Enter the details for your new shipping address
                  </DialogDescription>
                </DialogHeader>
                <AddressForm
                  onSubmit={handleAddAddress}
                  onCancel={() => setIsAddDialogOpen(false)}
                  isSubmitting={isUpdating}
                />
              </DialogContent>
            </Dialog>
          </div>
        </CardHeader>
        <CardContent>
          {addresses.length === 0 ? (
            <div className="text-center py-12">
              <MapPin className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-medium mb-2">No addresses yet</h3>
              <p className="text-muted-foreground mb-4">
                Add your first shipping address to make checkout faster
              </p>
              <Button onClick={() => setIsAddDialogOpen(true)}>
                <Plus className="h-4 w-4 mr-2" />
                Add Your First Address
              </Button>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {addresses.map((address) => (
                <AddressCard
                  key={address.id}
                  address={address}
                  onEdit={setEditingAddress}
                  onDelete={handleDeleteAddress}
                  onSetDefault={handleSetDefault}
                  isUpdating={isUpdating}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Edit Address Dialog */}
      {editingAddress && (
        <Dialog
          open={!!editingAddress}
          onOpenChange={(open) => !open && setEditingAddress(null)}
        >
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Edit Address</DialogTitle>
              <DialogDescription>
                Update the details for this shipping address
              </DialogDescription>
            </DialogHeader>
            <AddressForm
              address={editingAddress}
              onSubmit={handleUpdateAddress}
              onCancel={() => setEditingAddress(null)}
              isSubmitting={isUpdating}
            />
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}
