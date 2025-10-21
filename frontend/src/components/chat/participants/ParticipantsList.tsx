import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import {
  useConversationDetails,
  useConversationParticipants,
} from '@/hooks/chat/useConversationDetails'
import { cn } from '@/lib/utils'
import { Crown, MoreVertical, UserMinus, X } from 'lucide-react'
import { useState } from 'react'

interface ParticipantsListProps {
  conversationId: string
  currentUserId: string
  isUserOnline: (userId: string) => boolean
  onClose: () => void
}

export function ParticipantsList({
  conversationId,
  currentUserId,
  isUserOnline,
  onClose,
}: ParticipantsListProps) {
  const [selectedParticipant, setSelectedParticipant] = useState<null | string>(
    null,
  )

  const {
    data: conversation,
    error: conversationError,
    isLoading: conversationLoading,
  } = useConversationDetails(conversationId)

  const {
    data: participants,
    error: participantsError,
    isLoading: participantsLoading,
  } = useConversationParticipants(conversationId)

  const isLoading = conversationLoading || participantsLoading
  const error = conversationError || participantsError

  const getInitials = (userId: string) => {
    // Use user_id as fallback since we don't have display names
    return userId.slice(0, 2).toUpperCase()
  }

  const getRoleIcon = (userType: string) => {
    switch (userType) {
      case 'ADMIN':
        return <Crown className="h-3 w-3 text-yellow-500" />
      default:
        return null
    }
  }

  const getRoleColor = (userType: string) => {
    switch (userType) {
      case 'ADMIN':
        return 'text-yellow-600 dark:text-yellow-400'
      default:
        return 'text-gray-500 dark:text-gray-400'
    }
  }

  if (isLoading) {
    return (
      <Card className="w-full h-full flex flex-col">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">Participants</CardTitle>
            <Button onClick={onClose} size="sm" variant="ghost">
              <X className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="flex-1">
          <div className="space-y-3">
            {Array.from({ length: 8 }).map((_, i) => (
              <div className="flex items-center space-x-3" key={i}>
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="space-y-2 flex-1">
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-3 w-3/4" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error || !conversation) {
    return (
      <Card className="w-full h-full flex flex-col">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">Participants</CardTitle>
            <Button onClick={onClose} size="sm" variant="ghost">
              <X className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="text-center text-muted-foreground py-8">
            <p>Failed to load participants</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  const participantsList = participants || []
  const sortedParticipants = [...participantsList].sort((a, b) => {
    // Sort by online status first, then by userType, then by userId
    const aOnline = isUserOnline(a.userId)
    const bOnline = isUserOnline(b.userId)

    if (aOnline !== bOnline) return bOnline ? 1 : -1

    const roleOrder = { ADMIN: 0, USER: 1 }
    const aRoleOrder = roleOrder[a.userType]
    const bRoleOrder = roleOrder[b.userType]

    if (aRoleOrder !== bRoleOrder) return aRoleOrder - bRoleOrder

    return a.userId.localeCompare(b.userId)
  })

  const onlineCount = participantsList.filter((p) =>
    isUserOnline(p.userId),
  ).length

  return (
    <Card className="w-full h-full flex flex-col">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg">Participants</CardTitle>
            <p className="text-sm text-muted-foreground">
              {onlineCount} online • {participantsList.length} total
            </p>
          </div>
          <Button onClick={onClose} size="sm" variant="ghost">
            <X className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex-1 overflow-hidden p-0">
        <ScrollArea className="h-full">
          <div className="p-4 space-y-1">
            {sortedParticipants.map((participant) => {
              const isOnline = isUserOnline(participant.userId)
              const isCurrentUser = participant.userId === currentUserId

              return (
                <div
                  className={cn(
                    'flex items-center space-x-3 p-2 rounded-lg cursor-pointer transition-colors',
                    'hover:bg-gray-100 dark:hover:bg-gray-700',
                    selectedParticipant === participant.userId &&
                      'bg-gray-100 dark:bg-gray-700',
                  )}
                  key={participant.userId}
                  onClick={() =>
                    setSelectedParticipant(
                      selectedParticipant === participant.userId
                        ? null
                        : participant.userId,
                    )
                  }
                >
                  {/* Avatar with Online Status */}
                  <div className="relative">
                    <Avatar className="h-10 w-10">
                      <AvatarImage alt={participant.userId} src={undefined} />
                      <AvatarFallback className="text-sm bg-gradient-to-br from-blue-500 to-purple-600 text-white">
                        {getInitials(participant.userId)}
                      </AvatarFallback>
                    </Avatar>

                    {/* Online Status Indicator */}
                    <div
                      className={cn(
                        'absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-white dark:border-gray-800',
                        isOnline ? 'bg-green-500' : 'bg-gray-400',
                      )}
                    />
                  </div>

                  {/* User Info */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2">
                      <p
                        className={cn(
                          'font-medium text-sm truncate',
                          isCurrentUser && 'text-blue-600 dark:text-blue-400',
                        )}
                      >
                        {participant.userId}
                        {isCurrentUser && ' (You)'}
                      </p>
                      {getRoleIcon(participant.userType)}
                    </div>

                    <div className="flex items-center space-x-2">
                      <p
                        className={cn(
                          'text-xs capitalize',
                          getRoleColor(participant.userType),
                        )}
                      >
                        {participant.userType.toLowerCase()}
                      </p>
                      {isOnline && (
                        <span className="text-xs text-green-600 dark:text-green-400">
                          Online
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Actions Menu */}
                  {!isCurrentUser && (
                    <div className="flex items-center space-x-1">
                      <Button
                        className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={(e) => {
                          e.stopPropagation()
                          // Handle participant actions
                        }}
                        size="sm"
                        variant="ghost"
                      >
                        <MoreVertical className="h-3 w-3" />
                      </Button>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </ScrollArea>
      </CardContent>

      {/* Actions Footer */}
      <div className="border-t p-3">
        <div className="flex justify-between items-center">
          <Button className="text-xs" size="sm" variant="outline">
            Invite People
          </Button>
          <Button
            className="text-xs text-red-600 hover:text-red-700"
            size="sm"
            variant="ghost"
          >
            <UserMinus className="h-3 w-3 mr-1" />
            Leave
          </Button>
        </div>
      </div>
    </Card>
  )
}
