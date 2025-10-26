import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Skeleton } from '@/components/ui/skeleton'
import { useAddresses } from '@/hooks/address'
import { useCheckoutSessionStore } from '@/store/checkoutSessionStore'
import type { Address } from '@/types/__generated__/graphql'
import { CheckCircle, MapPin, Plus } from 'lucide-react'
import { useEffect, useState } from 'react'

export function AddressSelector() {
  const { selectedAddress, selectedAddressId, setAddress } =
    useCheckoutSessionStore()
  const { data, isLoading } = useAddresses(10)
  const [_isAddingNew, setIsAddingNew] = useState(false)

  const addresses = data?.edges.map((edge) => edge.node) || []

  // Reconstruct selectedAddress from selectedAddressId after fetch
  useEffect(() => {
    if (selectedAddressId && !selectedAddress && addresses.length > 0) {
      const address = addresses.find((addr) => addr.id === selectedAddressId)
      if (address) {
        // Just update the UI object in store (not calling backend)
        useCheckoutSessionStore.setState({ selectedAddress: address })
      }
    }
  }, [selectedAddressId, selectedAddress, addresses])

  // Auto-select default address on mount
  useEffect(() => {
    if (!selectedAddress && !selectedAddressId && addresses.length > 0) {
      const defaultAddress = addresses.find((addr) => addr.isDefault)
      if (defaultAddress) {
        setAddress(defaultAddress.id, defaultAddress)
      }
    }
  }, [addresses, selectedAddress, selectedAddressId, setAddress])

  const handleAddressChange = (addressId: string) => {
    const address = addresses.find((addr) => addr.id === addressId)
    if (address) {
      setAddress(addressId, address)
    }
  }

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MapPin className="h-5 w-5" />
            Delivery Address
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-20 w-full" />
        </CardContent>
      </Card>
    )
  }

  if (addresses.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MapPin className="h-5 w-5" />
            Delivery Address
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-muted-foreground text-sm">
            No saved addresses found. Please add a delivery address to continue.
          </p>
          <Button onClick={() => setIsAddingNew(true)} className="w-full">
            <Plus className="h-4 w-4 mr-2" />
            Add New Address
          </Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <MapPin className="h-5 w-5" />
          Delivery Address
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <RadioGroup
          value={selectedAddress?.id || selectedAddressId || ''}
          onValueChange={handleAddressChange}
        >
          {addresses.map((address: Address) => (
            <div key={address.id} className="space-y-2">
              <div className="flex items-start space-x-3">
                <RadioGroupItem
                  id={address.id}
                  value={address.id}
                  className="mt-1"
                />
                <div className="flex-1 space-y-1">
                  <label
                    htmlFor={address.id}
                    className="flex items-center justify-between font-medium cursor-pointer"
                  >
                    <div className="flex items-center gap-2">
                      <span>{address.receiverName}</span>
                      {address.isDefault && (
                        <Badge variant="default" className="text-xs">
                          Default
                        </Badge>
                      )}
                      {selectedAddress?.id === address.id && (
                        <CheckCircle className="h-4 w-4 text-green-600" />
                      )}
                    </div>
                  </label>
                  <div className="text-sm text-muted-foreground">
                    <p>
                      {address.addressLine1}
                      {address.addressLine2 && `, ${address.addressLine2}`}
                    </p>
                    <p>
                      {address.city}
                      {address.state && `, ${address.state}`}{' '}
                      {address.postalCode}
                    </p>
                    <p>{address.countryCode}</p>
                  </div>
                  {address.note && (
                    <p className="text-xs text-muted-foreground italic">
                      Note: {address.note}
                    </p>
                  )}
                </div>
              </div>
            </div>
          ))}
        </RadioGroup>

        <div className="pt-2 border-t">
          <Button
            variant="outline"
            onClick={() => setIsAddingNew(true)}
            className="w-full"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add New Address
          </Button>
        </div>

        {!selectedAddress && (
          <p className="text-sm text-muted-foreground">
            Please select a delivery address to continue
          </p>
        )}
      </CardContent>
    </Card>
  )
}
