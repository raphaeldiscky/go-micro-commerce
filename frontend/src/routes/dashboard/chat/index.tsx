import { ChatListPage } from '@/components/dashboard/chat/pages/ChatListPage'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/dashboard/chat/')({
  component: ChatListPage,
})
