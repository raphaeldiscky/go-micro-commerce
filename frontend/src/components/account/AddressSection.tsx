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
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  useAddresses,
  useCreateAddress,
  useDeleteAddress,
  useSetDefaultAddress,
  useUpdateAddress,
} from '@/hooks/address'
import type { Address, CreateAddressInput } from '@/types/__generated__/graphql'
import { MapPin, Plus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import { AddressCard } from './AddressCard'
import { AddressForm } from './AddressForm'

export function AddressSection() {
  // GraphQL hooks
  const { data: addressData, isLoading } = useAddresses(20)
  const createAddressMutation = useCreateAddress()
  const updateAddressMutation = useUpdateAddress()
  const deleteAddressMutation = useDeleteAddress()
  const setDefaultMutation = useSetDefaultAddress()

  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
  const [editingAddress, setEditingAddress] = useState<Address | null>(null)

  // Extract addresses from edges
  const addresses = addressData?.edges.map((edge) => edge.node) ?? []

  const handleAddAddress = async (data: CreateAddressInput) => {
    try {
      await createAddressMutation.mutateAsync(data)
      setIsAddDialogOpen(false)
      toast.success('Address added successfully')
    } catch (error) {
      console.error('Failed to add address:', error)
      toast.error('Failed to add address')
    }
  }

  const handleUpdateAddress = async (data: CreateAddressInput) => {
    if (!editingAddress) return

    try {
      await updateAddressMutation.mutateAsync({
        id: editingAddress.id,
        input: data,
      })
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
        await deleteAddressMutation.mutateAsync(id)
        toast.success('Address deleted successfully')
      } catch (error) {
        console.error('Failed to delete address:', error)
        toast.error('Failed to delete address')
      }
    }
  }

  const handleSetDefault = async (id: string) => {
    try {
      await setDefaultMutation.mutateAsync(id)
      toast.success('Default address updated successfully')
    } catch (error) {
      console.error('Failed to set default address:', error)
      toast.error('Failed to set default address')
    }
  }

  const isUpdating =
    createAddressMutation.isPending ||
    updateAddressMutation.isPending ||
    deleteAddressMutation.isPending ||
    setDefaultMutation.isPending

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
