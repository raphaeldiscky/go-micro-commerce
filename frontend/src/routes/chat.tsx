import { createFileRoute } from '@tanstack/react-router'
import { MultiUserChat } from '../components/MultiUserChat'

export const Route = createFileRoute('/chat')({
  component: ChatDemo,
})

function ChatDemo() {
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-4">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
            Multi-User WebSocket Chat Demo
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Demonstrating multiple simultaneous WebSocket connections in one
            page
          </p>
        </div>
        <MultiUserChat />
      </div>
    </div>
  )
}
