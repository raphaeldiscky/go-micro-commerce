import { ChatPanel } from '@/components/features/chat/panels/ChatPanel'
import { createFileRoute } from '@tanstack/react-router'
import { PATH_FEATURES } from '../../../constants'

export const Route = createFileRoute(PATH_FEATURES.chat.$conversationId)({
  component: ConversationRoute,
})

function ConversationRoute() {
  const { conversationId } = Route.useParams()

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-2 md:p-4">
      <ChatPanel conversationId={conversationId} />
    </div>
  )
}
