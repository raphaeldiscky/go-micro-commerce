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
import { useAccountStore } from '@/store/accountStore'
import type { AddressRequest, CustomerAddress } from '@/types/account'
import { MapPin, Plus } from 'lucide-react'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { AddressCard } from './AddressCard'
import { AddressForm } from './AddressForm'

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
