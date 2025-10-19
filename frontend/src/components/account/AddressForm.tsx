import { Button } from '@/components/ui/button'
import { DialogFooter } from '@/components/ui/dialog'
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
import type { AddressRequest, CustomerAddress } from '@/types/account'
import { addressSchema } from '@/types/account'
import { useForm } from '@tanstack/react-form'

interface AddressFormProps {
  address?: CustomerAddress
  onSubmit: (data: AddressRequest) => void
  onCancel: () => void
  isSubmitting: boolean
}

export function AddressForm({
  address,
  onSubmit,
  onCancel,
  isSubmitting,
}: AddressFormProps) {
  const form = useForm({
    defaultValues: {
      receiverName: address?.receiverName || '',
      addressLine1: address?.addressLine1 || '',
      addressLine2: address?.addressLine2 || '',
      city: address?.city || '',
      state: address?.state || '',
      postalCode: address?.postalCode || '',
      countryCode: address?.countryCode || 'US',
      note: address?.note || '',
      isDefault: address?.isDefault || false,
    },
    validators: {
      onChange: addressSchema,
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
      <form.Field name="receiverName">
        {(field) => (
          <div className="space-y-2">
            <Label htmlFor="receiverName">Receiver Name</Label>
            <Input
              id="receiverName"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              placeholder="Full name"
            />
          </div>
        )}
      </form.Field>

      <form.Field name="addressLine1">
        {(field) => (
          <div className="space-y-2">
            <Label htmlFor="addressLine1">Address Line 1</Label>
            <Input
              id="addressLine1"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              placeholder="123 Main Street"
            />
          </div>
        )}
      </form.Field>

      <form.Field name="addressLine2">
        {(field) => (
          <div className="space-y-2">
            <Label htmlFor="addressLine2">
              Apartment, Suite, Unit (Optional)
            </Label>
            <Input
              id="addressLine2"
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

        <form.Field name="countryCode">
          {(field) => (
            <div className="space-y-2">
              <Label htmlFor="countryCode">Country</Label>
              <Select
                value={field.state.value}
                onValueChange={(value) => field.handleChange(value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select country code" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="US">United States</SelectItem>
                  <SelectItem value="CA">Canada</SelectItem>
                  <SelectItem value="UK">United Kingdom</SelectItem>
                  <SelectItem value="AU">Australia</SelectItem>
                  <SelectItem value="GM">Germany</SelectItem>
                  <SelectItem value="FR">France</SelectItem>
                  <SelectItem value="JP">Japan</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}
        </form.Field>
      </div>

      <form.Field name="note">
        {(field) => (
          <div className="space-y-2">
            <Label htmlFor="note">Delivery Instructions (Optional)</Label>
            <Textarea
              id="note"
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
