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
