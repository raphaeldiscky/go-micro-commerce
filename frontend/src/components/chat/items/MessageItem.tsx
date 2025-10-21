import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { formatTime } from '@/lib/utils/date'
import type { Message } from '@/types/__generated__/graphql'
import { Check, Reply } from 'lucide-react'

interface MessageItemProps {
  currentUserId: string
  isConsecutive?: boolean
  message: Message
  onReply?: (message: Message) => void
}

export function MessageItem({
  currentUserId,
  isConsecutive = false,
  message,
  onReply,
}: MessageItemProps) {
  const isOwn = message.senderId === currentUserId
  const isSystem = message.isSystem

  const getDeliveryStatusIcon = () => {
    // GraphQL doesn't have delivery_status, so we show a simple sent icon
    return <Check className="h-3 w-3 text-gray-400" />
  }

  const getInitials = (name?: string) => {
    if (!name) return 'U' // Default fallback for undefined names

    return name
      .split(' ')
      .map((word) => word[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  if (isSystem) {
    return (
      <div className="flex justify-center my-4">
        <div className="bg-gray-100 dark:bg-gray-800 px-3 py-1 rounded-full">
          <span className="text-xs text-gray-600 dark:text-gray-400">
            {message.content}
          </span>
        </div>
      </div>
    )
  }

  return (
    <div
      className={cn(
        'flex w-full mb-4 group',
        isOwn ? 'justify-end' : 'justify-start',
      )}
    >
      <div
        className={cn(
          'flex max-w-[70%] space-x-2',
          isOwn ? 'flex-row-reverse space-x-reverse' : 'flex-row',
        )}
      >
        {/* Avatar */}
        {!isOwn && (
          <div className="flex-shrink-0">
            {!isConsecutive && (
              <Avatar className="h-8 w-8">
                <AvatarImage
                  alt={
                    message.sender
                      ? `${message.sender.firstName} ${message.sender.lastName}`
                      : 'User'
                  }
                  src={undefined}
                />
                <AvatarFallback className="text-xs bg-gradient-to-br from-blue-500 to-purple-600 text-white">
                  {getInitials(
                    message.sender
                      ? `${message.sender.firstName} ${message.sender.lastName}`
                      : undefined,
                  )}
                </AvatarFallback>
              </Avatar>
            )}
            {isConsecutive && <div className="h-8 w-8" />}
          </div>
        )}

        {/* Message Content */}
        <div
          className={cn('flex flex-col', isOwn ? 'items-end' : 'items-start')}
        >
          {/* Sender Name and Time */}
          {!isConsecutive && !isOwn && (
            <div className="flex items-center space-x-2 mb-1">
              <span className="text-xs font-semibold text-gray-700 dark:text-gray-300">
                {message.sender
                  ? `${message.sender.firstName} ${message.sender.lastName}`
                  : message.senderId
                    ? `User ${message.senderId.slice(-4)}`
                    : 'System'}
              </span>
              <span className="text-xs text-gray-500 dark:text-gray-400">
                {formatTime(message.createdAt)}
              </span>
            </div>
          )}

          {/* Message Bubble */}
          <div
            className={cn(
              'relative rounded-2xl px-4 py-2 max-w-full break-words',
              isOwn
                ? 'bg-blue-500 text-white rounded-br-md'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white rounded-bl-md',
            )}
          >
            {/* Message Content */}
            <div className="text-sm leading-relaxed whitespace-pre-wrap">
              {message.content}
            </div>

            {/* Message Time and Status for Own Messages */}
            <div
              className={cn(
                'flex items-center justify-end mt-1 space-x-1',
                isOwn ? 'text-blue-100' : 'text-gray-500 dark:text-gray-400',
              )}
            >
              <span className="text-xs">{formatTime(message.createdAt)}</span>
              {isOwn && getDeliveryStatusIcon()}
            </div>
          </div>

          {/* Message Actions */}
          <div
            className={cn(
              'opacity-0 group-hover:opacity-100 transition-opacity mt-1 flex space-x-1',
              isOwn ? 'flex-row-reverse' : 'flex-row',
            )}
          >
            {onReply && (
              <Button
                className="h-6 px-2 text-xs text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                onClick={() => onReply(message)}
                size="sm"
                variant="ghost"
              >
                <Reply className="h-3 w-3 mr-1" />
                Reply
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
