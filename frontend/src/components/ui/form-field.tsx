import type { ReactNode } from 'react'
import { Label } from './label'

interface FormFieldProps {
  field: any
  label: string
  description?: string
  required?: boolean
  children: (field: any) => ReactNode
}

export function FormField({
  field,
  label,
  description,
  required,
  children,
}: FormFieldProps) {
  const hasError = field.state.meta.errors.length > 0
  const errorId = hasError ? `${String(field.name)}-error` : undefined
  const descriptionId = description
    ? `${String(field.name)}-description`
    : undefined

  return (
    <div className="space-y-2">
      <Label htmlFor={String(field.name)}>
        {label}
        {required && <span className="text-destructive -ml-1">*</span>}
      </Label>
      {description && (
        <p id={descriptionId} className="text-sm text-muted-foreground">
          {description}
        </p>
      )}
      {children(field)}
      {hasError && (
        <p id={errorId} className="text-sm text-destructive">
          {field.state.meta.errors[0]?.message || field.state.meta.errors[0]}
        </p>
      )}
    </div>
  )
}
