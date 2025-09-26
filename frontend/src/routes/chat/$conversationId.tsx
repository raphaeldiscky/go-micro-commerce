import { createFileRoute } from '@tanstack/react-router'
import { ConversationPage } from '../../components/chat/pages/ConversationPage'

export const Route = createFileRoute('/chat/$conversationId')({
  component: ConversationRoute,
})

function ConversationRoute() {
  const { conversationId } = Route.useParams()

  return <ConversationPage conversationId={conversationId} />
}
