import { Button } from '@/components/ui/button'
import { DialogFooter } from '@/components/ui/dialog'
import { FormField } from '@/components/ui/form-field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import type { AddressFormValues } from '@/schemas/account'
import { addressSchema } from '@/schemas/account'
import type { Address, CreateAddressInput } from '@/types/__generated__/graphql'
import { useForm } from '@tanstack/react-form'

interface AddressFormProps {
  address?: Address
  onSubmit: (data: CreateAddressInput) => void
  onCancel: () => void
  isSubmitting: boolean
}

export function AddressForm({
  address,
  onSubmit,
  onCancel,
  isSubmitting,
}: AddressFormProps) {
  const defaultValues: AddressFormValues = {
    receiverName: address?.receiverName || '',
    addressLine1: address?.addressLine1 || '',
    addressLine2: address?.addressLine2 || '',
    city: address?.city || '',
    state: address?.state || '',
    postalCode: address?.postalCode || '',
    countryCode: address?.countryCode || 'US',
    note: address?.note || '',
    isDefault: address?.isDefault || false,
  }

  const form = useForm({
    defaultValues,
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
          <FormField field={field} label="Receiver Name" required>
            {(f) => (
              <Input
                id={f.name}
                value={f.state.value}
                onChange={(e) => f.handleChange(e.target.value)}
                onBlur={f.handleBlur}
                placeholder="Full name"
                className={
                  f.state.meta.errors.length > 0 ? 'border-destructive' : ''
                }
              />
            )}
          </FormField>
        )}
      </form.Field>

      <form.Field name="addressLine1">
        {(field) => (
          <FormField field={field} label="Address Line 1" required>
            {(f) => (
              <Input
                id={f.name}
                value={f.state.value}
                onChange={(e) => f.handleChange(e.target.value)}
                onBlur={f.handleBlur}
                placeholder="123 Main Street"
                className={
                  f.state.meta.errors.length > 0 ? 'border-destructive' : ''
                }
              />
            )}
          </FormField>
        )}
      </form.Field>

      <form.Field name="addressLine2">
        {(field) => (
          <FormField field={field} label="Apartment, Suite, Unit">
            {(f) => (
              <Input
                id={f.name}
                value={f.state.value}
                onChange={(e) => f.handleChange(e.target.value)}
                onBlur={f.handleBlur}
                placeholder="Apt 4B, Suite 200"
                className={
                  f.state.meta.errors.length > 0 ? 'border-destructive' : ''
                }
              />
            )}
          </FormField>
        )}
      </form.Field>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <form.Field name="city">
          {(field) => (
            <FormField field={field} label="City" required>
              {(f) => (
                <Input
                  id={f.name}
                  value={f.state.value}
                  onChange={(e) => f.handleChange(e.target.value)}
                  onBlur={f.handleBlur}
                  placeholder="New York"
                  className={
                    f.state.meta.errors.length > 0 ? 'border-destructive' : ''
                  }
                />
              )}
            </FormField>
          )}
        </form.Field>

        <form.Field name="state">
          {(field) => (
            <FormField field={field} label="State/Province">
              {(f) => (
                <Input
                  id={f.name}
                  value={f.state.value}
                  onChange={(e) => f.handleChange(e.target.value)}
                  onBlur={f.handleBlur}
                  placeholder="NY"
                  className={
                    f.state.meta.errors.length > 0 ? 'border-destructive' : ''
                  }
                />
              )}
            </FormField>
          )}
        </form.Field>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <form.Field name="postalCode">
          {(field) => (
            <FormField field={field} label="Postal Code" required>
              {(f) => (
                <Input
                  id={f.name}
                  value={f.state.value}
                  onChange={(e) => f.handleChange(e.target.value)}
                  onBlur={f.handleBlur}
                  placeholder="10001"
                  className={
                    f.state.meta.errors.length > 0 ? 'border-destructive' : ''
                  }
                />
              )}
            </FormField>
          )}
        </form.Field>

        <form.Field name="countryCode">
          {(field) => (
            <FormField field={field} label="Country" required>
              {(f) => (
                <Select
                  value={f.state.value}
                  onValueChange={(value) => f.handleChange(value)}
                >
                  <SelectTrigger
                    className={
                      f.state.meta.errors.length > 0 ? 'border-destructive' : ''
                    }
                  >
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
              )}
            </FormField>
          )}
        </form.Field>
      </div>

      <form.Field name="note">
        {(field) => (
          <FormField field={field} label="Notes">
            {(f) => (
              <Textarea
                id={f.name}
                value={f.state.value}
                onChange={(e) => f.handleChange(e.target.value)}
                onBlur={f.handleBlur}
                placeholder="Ring doorbell twice, leave at front door, etc."
                rows={3}
                className={
                  f.state.meta.errors.length > 0 ? 'border-destructive' : ''
                }
              />
            )}
          </FormField>
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
        <Button type="submit" disabled={isSubmitting || !form.state.canSubmit}>
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
