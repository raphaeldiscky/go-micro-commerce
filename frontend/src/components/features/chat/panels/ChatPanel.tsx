import { WebSocketSendProvider } from '@/contexts/WebSocketSendContext'
import { useEffect, useState } from 'react'
import { Card } from '../../../ui/card'
import { Dialog, DialogContent } from '../../../ui/dialog'
import { ConversationPage } from '../pages/ConversationPage'

interface ChatPanelProps {
  conversationId: string
  defaultFullscreen?: boolean
}

export function ChatPanel({
  conversationId,
  defaultFullscreen = false,
}: ChatPanelProps) {
  const [isFullscreen, setIsFullscreen] = useState(defaultFullscreen)
  const [isMobile, setIsMobile] = useState(false)

  // Check for mobile screen size
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768) // Tailwind md breakpoint
    }

    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  // Auto-fullscreen on mobile by default
  useEffect(() => {
    if (isMobile && !defaultFullscreen) {
      setIsFullscreen(true)
    }
  }, [isMobile, defaultFullscreen])

  // Handle ESC key to exit fullscreen
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && isFullscreen) {
        setIsFullscreen(false)
      }
    }

    if (isFullscreen) {
      document.addEventListener('keydown', handleKeyDown)
      return () => document.removeEventListener('keydown', handleKeyDown)
    }
  }, [isFullscreen])

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen)
  }

  const handleDialogOpenChange = (open: boolean) => {
    if (!open) {
      setIsFullscreen(false)
    }
  }

  if (isFullscreen) {
    return (
      <WebSocketSendProvider>
        <Dialog open={isFullscreen} onOpenChange={handleDialogOpenChange}>
          <DialogContent
            className="w-screen h-screen max-w-none max-h-none m-0 p-0 border-0 bg-transparent"
            showCloseButton={false}
          >
            <div className="w-full h-full">
              <ConversationPage
                conversationId={conversationId}
                isFullscreen={true}
                onToggleFullscreen={toggleFullscreen}
              />
            </div>
          </DialogContent>
        </Dialog>
      </WebSocketSendProvider>
    )
  }

  return (
    <WebSocketSendProvider>
      <Card className="h-[calc(100vh-8rem)] min-h-[500px] max-h-[800px] md:max-w-4xl mx-auto overflow-hidden w-full">
        <div className="h-full flex flex-col">
          <ConversationPage
            conversationId={conversationId}
            isFullscreen={false}
            onToggleFullscreen={!isMobile ? toggleFullscreen : undefined}
            showToggle={!isMobile}
          />
        </div>
      </Card>
    </WebSocketSendProvider>
  )
}
