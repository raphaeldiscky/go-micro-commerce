import { cn } from '@/lib/utils'
import type { Message } from '@/types/__generated__/graphql'
import { Paperclip, Send, Smile, X } from 'lucide-react'
import { useCallback, useRef, useState } from 'react'
import { Button } from '../../../ui/button'
import { Textarea } from '../../../ui/textarea'

interface MessageInputProps {
  disabled?: boolean
  isLoading?: boolean
  onCancelReply?: () => void
  onSendMessage: (content: string) => void
  onTypingStart?: () => void
  onTypingStop?: () => void
  placeholder?: string
  replyingTo?: Message | null
}

export function MessageInput({
  disabled = false,
  isLoading = false,
  onCancelReply,
  onSendMessage,
  onTypingStart,
  onTypingStop,
  placeholder = 'Type a message...',
  replyingTo,
}: MessageInputProps) {
  const [message, setMessage] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()
      if (!message.trim() || disabled || isLoading) return

      onSendMessage(message.trim())
      setMessage('')

      // Clear typing indicator
      if (isTyping) {
        onTypingStop?.()
        setIsTyping(false)
      }

      // Reset textarea height
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto'
      }
    },
    [message, disabled, isLoading, onSendMessage, onTypingStop, isTyping],
  )

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        handleSubmit(e)
      }
    },
    [handleSubmit],
  )

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const value = e.target.value
      setMessage(value)

      // Auto-resize textarea
      const textarea = e.target
      textarea.style.height = 'auto'
      textarea.style.height = `${Math.min(textarea.scrollHeight, 120)}px`

      // Handle typing indicators
      if (value.trim() && !isTyping) {
        setIsTyping(true)
        onTypingStart?.()
      }
    },
    [isTyping, onTypingStart, onTypingStop],
  )

  return (
    <div className="border-t bg-white dark:bg-gray-800 p-4 flex-shrink-0">
      {/* Reply Context */}
      {replyingTo && (
        <div className="mb-2 p-2 bg-gray-50 dark:bg-gray-700 rounded-lg border-l-2 border-blue-500">
          <div className="flex items-center justify-between">
            <div className="flex-1 min-w-0">
              <p className="text-xs font-medium text-gray-700 dark:text-gray-300">
                Replying to{' '}
                {replyingTo.sender
                  ? `${replyingTo.sender.firstName} ${replyingTo.sender.lastName}`
                  : 'Unknown'}
              </p>
              <p className="text-xs text-gray-500 dark:text-gray-400 truncate">
                {replyingTo.content}
              </p>
            </div>
            <Button
              className="h-6 w-6 p-0 ml-2"
              onClick={onCancelReply}
              size="sm"
              variant="ghost"
            >
              <X className="h-3 w-3" />
            </Button>
          </div>
        </div>
      )}

      {/* Message Input Form */}
      <form className="flex items-end space-x-2" onSubmit={handleSubmit}>
        {/* File Attachment Button */}
        <Button
          className="h-9 w-9 p-0 shrink-0"
          disabled={disabled}
          size="sm"
          type="button"
          variant="ghost"
        >
          <Paperclip className="h-4 w-4" />
        </Button>

        {/* Message Input */}
        <div className="flex-1 relative">
          <Textarea
            className={cn(
              'min-h-[36px] max-h-[120px] resize-none py-2 pr-12',
              'border-gray-200 dark:border-gray-600',
              'focus:border-blue-500 focus:ring-blue-500',
            )}
            disabled={disabled}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            ref={textareaRef}
            value={message}
          />

          {/* Emoji Button */}
          <Button
            className="absolute right-2 top-2 h-6 w-6 p-0"
            disabled={disabled}
            size="sm"
            type="button"
            variant="ghost"
          >
            <Smile className="h-4 w-4" />
          </Button>
        </div>

        {/* Send Button */}
        <Button
          className="h-9 w-9 p-0 shrink-0"
          disabled={disabled || !message.trim() || isLoading}
          type="submit"
        >
          <Send className="h-4 w-4" />
        </Button>
      </form>
    </div>
  )
}
