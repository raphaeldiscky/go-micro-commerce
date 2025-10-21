import { ChatListPage } from '@/components/chat/pages/ChatListPage'
import type { FileRoutesByPath } from '@tanstack/react-router'
import { createFileRoute } from '@tanstack/react-router'
import { PATH } from '../../constants'

export const Route = createFileRoute(PATH.chat.root as keyof FileRoutesByPath)({
  component: ChatListPage,
})
