import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { useCheckoutSessionStore } from '@/store/checkoutSessionStore'
import { FileText } from 'lucide-react'

export function OrderNotes() {
  const { orderNote, setOrderNote } = useCheckoutSessionStore()

  const handleNoteChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setOrderNote(e.target.value)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="h-5 w-5" />
          Order Notes (Optional)
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="order-notes">
            Add special instructions for your order
          </Label>
          <Textarea
            id="order-notes"
            placeholder="e.g., Delivery instructions, gift wrapping, special requests..."
            value={orderNote || ''}
            onChange={handleNoteChange}
            rows={3}
            className="resize-none"
          />
        </div>
        <p className="text-xs text-muted-foreground">
          Any special requests or delivery instructions you'd like us to know
          about.
        </p>
      </CardContent>
    </Card>
  )
}
