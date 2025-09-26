import type { TypingIndicator as TypingIndicatorType } from '@/lib/api'
import { Avatar, AvatarFallback, AvatarImage } from '../../ui/avatar'

interface TypingIndicatorProps {
  typingUsers: Array<TypingIndicatorType>
}

export function TypingIndicator({ typingUsers }: TypingIndicatorProps) {
  if (!typingUsers.length) return null

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((word) => word[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  const getTypingMessage = () => {
    const count = typingUsers.length
    if (count === 1) {
      return `${typingUsers[0].username} is typing...`
    } else if (count === 2) {
      return `${typingUsers[0].username} and ${typingUsers[1].username} are typing...`
    } else {
      return `${typingUsers[0].username} and ${count - 1} others are typing...`
    }
  }

  return (
    <div className="flex items-center space-x-2 px-4 py-2">
      {/* User Avatars */}
      <div className="flex -space-x-1">
        {typingUsers.slice(0, 3).map((user) => (
          <Avatar
            className="h-6 w-6 border border-white dark:border-gray-800"
            key={user.user_id}
          >
            <AvatarImage alt={user.username} src={undefined} />
            <AvatarFallback className="text-xs bg-gradient-to-br from-blue-500 to-purple-600 text-white">
              {getInitials(user.username)}
            </AvatarFallback>
          </Avatar>
        ))}
      </div>

      {/* Typing Message */}
      <div className="flex items-center space-x-2">
        <span className="text-sm text-gray-600 dark:text-gray-400">
          {getTypingMessage()}
        </span>

        {/* Animated Dots */}
        <div className="flex space-x-1">
          <div
            className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce"
            style={{ animationDelay: '0ms' }}
          />
          <div
            className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce"
            style={{ animationDelay: '150ms' }}
          />
          <div
            className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce"
            style={{ animationDelay: '300ms' }}
          />
        </div>
      </div>
    </div>
  )
}
