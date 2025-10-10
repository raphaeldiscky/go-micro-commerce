import { useMessageReceipts } from '@/hooks/chat/useMessageReceipts'
import { useMessages } from '@/hooks/chat/useMessages'
import { areMessagesConsecutive } from '@/lib/utils/date'
import type {
  Message,
  TypingIndicator as TypingIndicatorType,
} from '@/types/__generated__/graphql'
import { ChevronDown, Loader2 } from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Button } from '../../../ui/button'
import { ScrollArea } from '../../../ui/scroll-area'
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
  const markedAsReadRef = useRef<Set<string>>(new Set())
  const lastProcessedMessageIdRef = useRef<string | null>(null)
  const isMarkingRef = useRef(false)

  const {
    data: messagesData,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useMessages(conversationId)

  const { markAsRead } = useMessageReceipts(conversationId)

  useEffect(() => {
    markedAsReadRef.current.clear()
    lastProcessedMessageIdRef.current = null
    isMarkingRef.current = false
  }, [conversationId])

  const messages = useMemo(() => {
    if (!messagesData?.pages) return []
    return messagesData.pages.flatMap((page) =>
      page.conversationMessages.edges.map((edge) => edge.node),
    )
  }, [messagesData?.pages])

  const isConsecutiveMessage = (
    currentMsg: Message,
    prevMsg: Message | undefined,
  ) => {
    if (!prevMsg) return false
    if (currentMsg.senderId !== prevMsg.senderId) return false

    return areMessagesConsecutive(currentMsg.createdAt, prevMsg.createdAt)
  }

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

  const handleScroll = useCallback(
    (e: Event) => {
      const target = e.target as HTMLElement

      const { clientHeight, scrollHeight, scrollTop } = target
      const isAtBottom = scrollHeight - scrollTop <= clientHeight + 100

      setAutoScroll(isAtBottom)
      setShowScrollButton(!isAtBottom)

      if (scrollTop < 100 && hasNextPage && !isFetchingNextPage) {
        fetchNextPage()
      }
    },
    [hasNextPage, isFetchingNextPage, fetchNextPage],
  )

  useEffect(() => {
    if (isMarkingRef.current) return

    const unreadMessages = messages.filter(
      (msg) => msg.senderId !== currentUserId,
    )

    if (unreadMessages.length === 0 || !autoScroll) return

    const lastMessageId = messages[messages.length - 1]?.id
    if (lastMessageId === lastProcessedMessageIdRef.current) return

    const messagesToMark = unreadMessages
      .slice(-3)
      .filter((msg) => !markedAsReadRef.current.has(msg.id))
      .filter((msg) => !msg.id.startsWith('temp-'))

    if (messagesToMark.length > 0) {
      isMarkingRef.current = true
      lastProcessedMessageIdRef.current = lastMessageId

      Promise.all(
        messagesToMark.map((msg) => {
          markedAsReadRef.current.add(msg.id)
          return markAsRead(msg.id)
        }),
      ).finally(() => {
        isMarkingRef.current = false
      })
    }
  }, [messages, currentUserId, autoScroll, markAsRead])

  useEffect(() => {
    scrollToBottom()
  }, [messages.length, typingUsers.length])

  useEffect(() => {
    const scrollElement = scrollAreaRef.current?.querySelector(
      '[data-radix-scroll-area-viewport]',
    )
    if (scrollElement) {
      scrollElement.addEventListener('scroll', handleScroll)
      return () => scrollElement.removeEventListener('scroll', handleScroll)
    }
  }, [handleScroll])

  return (
    <div className="flex-1 relative min-h-0">
      <ScrollArea className="h-full" ref={scrollAreaRef}>
        <div className="p-4 space-y-1">
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

          {typingUsers.length > 0 && (
            <TypingIndicator typingUsers={typingUsers} />
          )}
        </div>
      </ScrollArea>

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
