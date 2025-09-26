import { createFileRoute } from '@tanstack/react-router'
import { ChatListPage } from '../components/chat/pages/ChatListPage'

export const Route = createFileRoute('/chat')({
  component: ChatListPage,
})
