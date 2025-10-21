import { ChatPanel } from '@/components/chat/panels/ChatPanel'
import { createFileRoute } from '@tanstack/react-router'
import { PATH } from '../../constants'

export const Route = createFileRoute(PATH.chat.$conversationId)({
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
