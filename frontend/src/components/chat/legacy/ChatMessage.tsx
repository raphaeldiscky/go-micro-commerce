import { cn } from '@/lib/utils'

export interface ChatMessage {
  content: string
  id: string
  senderId: string
  senderName: string
  timestamp: Date
  type: 'received' | 'sent'
}

interface ChatMessageProps {
  currentUserId: string
  message: ChatMessage
}

export function ChatMessageComponent({
  currentUserId,
  message,
}: ChatMessageProps) {
  const isSent = message.senderId === currentUserId

  return (
    <div className={cn('flex mb-4', isSent ? 'justify-end' : 'justify-start')}>
      <div
        className={cn('max-w-xs lg:max-w-md px-4 py-2 rounded-lg', {
          'bg-blue-500 text-white rounded-br-none': isSent,
          'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white rounded-bl-none':
            !isSent,
        })}
      >
        {!isSent && (
          <div className="text-xs font-semibold mb-1 text-gray-600 dark:text-gray-400">
            {message.senderName}
          </div>
        )}
        <div className="text-sm">{message.content}</div>
        <div
          className={cn('text-xs mt-1', {
            'text-blue-100': isSent,
            'text-gray-500 dark:text-gray-400': !isSent,
          })}
        >
          {message.timestamp.toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit',
          })}
        </div>
      </div>
    </div>
  )
}
