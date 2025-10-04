import { ChatListPage } from '@/components/features/chat/pages/ChatListPage'
import type { FileRoutesByPath } from '@tanstack/react-router'
import { createFileRoute } from '@tanstack/react-router'
import { PATH_FEATURES } from '../../../constants'

export const Route = createFileRoute(
  PATH_FEATURES.chat.root as keyof FileRoutesByPath,
)({
  component: ChatListPage,
})
