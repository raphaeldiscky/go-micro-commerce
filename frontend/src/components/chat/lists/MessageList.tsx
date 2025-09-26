import { useMessageReceipts } from '@/hooks/chat/useMessageReceipts'
import { useMessages } from '@/hooks/chat/useMessages'
import type { Message, TypingIndicator as TypingIndicatorType } from '@/lib/api'
import { areMessagesConsecutive } from '@/lib/utils/date'
import { ChevronDown, Loader2 } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { Button } from '../../ui/button'
import { ScrollArea } from '../../ui/scroll-area'
import { TypingIndicator } from '../indicators/TypingIndicator'
import { MessageItem } from '../items/MessageItem'

interface MessageListProps {
  conversationId: string
  currentUserId: string
  onReply?: (message: Message) => void
  typingUsers: Array<TypingIndicatorType>
}

export function MessageList({
  conversationId,
  currentUserId,
  onReply,
  typingUsers,
}: MessageListProps) {
  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const [showScrollButton, setShowScrollButton] = useState(false)
  const [autoScroll, setAutoScroll] = useState(true)

  const {
    data: messagesData,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useMessages(conversationId)

  const { markAsRead } = useMessageReceipts(conversationId)

  // Flatten all message pages
  // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
  const messages = messagesData?.pages.flat() || []

  // Check if message is consecutive (same sender within 5 minutes)
  const isConsecutiveMessage = (
    currentMsg: Message,
    prevMsg: Message | undefined,
  ) => {
    if (!prevMsg) return false
    if (currentMsg.sender_id !== prevMsg.sender_id) return false

    return areMessagesConsecutive(currentMsg.created_at, prevMsg.created_at)
  }

  // Scroll to bottom
  const scrollToBottom = (force = false) => {
    if (scrollAreaRef.current && (autoScroll || force)) {
      const scrollElement = scrollAreaRef.current.querySelector(
        '[data-radix-scroll-area-viewport]',
      )
      if (scrollElement) {
        scrollElement.scrollTop = scrollElement.scrollHeight
      }
    }
  }

  // Handle scroll events
  const handleScroll = (e: Event) => {
    const target = e.target as HTMLElement

    const { clientHeight, scrollHeight, scrollTop } = target
    const isAtBottom = scrollHeight - scrollTop <= clientHeight + 100

    setAutoScroll(isAtBottom)
    setShowScrollButton(!isAtBottom)

    // Load more messages when scrolling to top
    if (scrollTop < 100 && hasNextPage && !isFetchingNextPage) {
      fetchNextPage()
    }
  }

  // Mark messages as read when they come into view
  useEffect(() => {
    const unreadMessages = messages.filter(
      (msg) =>
        msg.sender_id !== currentUserId && msg.delivery_status !== 'read',
    )

    if (unreadMessages.length > 0 && autoScroll) {
      // Mark the last few messages as read
      const messagesToMark = unreadMessages.slice(-3)
      messagesToMark.forEach((msg) => markAsRead(msg.id))
    }
  }, [messages, currentUserId, markAsRead, autoScroll])

  // Auto-scroll when new messages arrive
  useEffect(() => {
    scrollToBottom()
  }, [messages.length, typingUsers.length])

  // Setup scroll listener
  useEffect(() => {
    const scrollElement = scrollAreaRef.current?.querySelector(
      '[data-radix-scroll-area-viewport]',
    )
    if (scrollElement) {
      scrollElement.addEventListener('scroll', handleScroll)
      return () => scrollElement.removeEventListener('scroll', handleScroll)
    }
  }, [hasNextPage, isFetchingNextPage])

  return (
    <div className="flex-1 relative min-h-0">
      <ScrollArea className="h-full" ref={scrollAreaRef}>
        <div className="p-4 space-y-1">
          {/* Load More Button */}
          {hasNextPage && (
            <div className="flex justify-center py-4">
              <Button
                className="text-muted-foreground"
                disabled={isFetchingNextPage}
                onClick={() => fetchNextPage()}
                size="sm"
                variant="ghost"
              >
                {isFetchingNextPage ? (
                  <>
                    <Loader2 className="h-3 w-3 mr-2 animate-spin" />
                    Loading...
                  </>
                ) : (
                  'Load older messages'
                )}
              </Button>
            </div>
          )}

          {/* Messages */}
          {messages.length === 0 ? (
            <div className="flex items-center justify-center py-12 text-muted-foreground">
              <div className="text-center">
                <div className="text-lg font-medium mb-2">No messages yet</div>
                <div className="text-sm">
                  Start the conversation by sending a message!
                </div>
              </div>
            </div>
          ) : (
            messages.map((message, index) => (
              <MessageItem
                currentUserId={currentUserId}
                isConsecutive={isConsecutiveMessage(
                  message,
                  messages[index - 1],
                )}
                key={message.id}
                message={message}
                onReply={onReply}
              />
            ))
          )}

          {/* Typing Indicators */}
          {typingUsers.length > 0 && (
            <TypingIndicator typingUsers={typingUsers} />
          )}
        </div>
      </ScrollArea>

      {/* Scroll to Bottom Button */}
      {showScrollButton && (
        <div className="absolute bottom-4 right-4">
          <Button
            className="rounded-full shadow-lg"
            onClick={() => {
              setAutoScroll(true)
              scrollToBottom(true)
            }}
            size="sm"
          >
            <ChevronDown className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  )
}
