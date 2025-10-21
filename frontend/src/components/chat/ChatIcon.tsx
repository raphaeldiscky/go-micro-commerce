import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useNavigate } from '@tanstack/react-router'
import { MessageCircle } from 'lucide-react'
import { PATH } from '../../constants'

export function ChatIcon() {
  const navigate = useNavigate()
  const handleClick = () => {
    navigate({ to: PATH.chat.root })
  }

  return (
    <Button
      aria-label="Open chat"
      className={cn(
        'relative h-10 w-10 rounded-full transition-all duration-200 hover:scale-105 active:scale-95',
      )}
      onClick={handleClick}
      size="icon"
      variant="ghost"
    >
      <MessageCircle className="h-5 w-5" />
    </Button>
  )
}
