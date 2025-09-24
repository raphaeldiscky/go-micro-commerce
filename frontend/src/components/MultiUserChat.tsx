import { useCallback, useState } from 'react'
import { Button } from './ui/button'
import { UserChatPanel } from './UserChatPanel'

interface User {
  id: string
  name: string
  ticket: string
}

export function MultiUserChat() {
  const [users, setUsers] = useState<Array<User>>([
    { id: '1', name: 'User 1', ticket: 'user1-ticket' },
    { id: '2', name: 'User 2', ticket: 'user2-ticket' },
  ])

  const addUser = useCallback(() => {
    const newUserId = (users.length + 1).toString()
    const newUser: User = {
      id: newUserId,
      name: `User ${newUserId}`,
      ticket: `user${newUserId}-ticket`,
    }
    setUsers((prev) => [...prev, newUser])
  }, [users.length])

  const removeUser = useCallback((userId: string) => {
    setUsers((prev) => prev.filter((user) => user.id !== userId))
  }, [])

  return (
    <div className="space-y-6">
      <div className="flex justify-center">
        <Button onClick={addUser} variant="outline">
          Add User Panel
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {users.map((user) => (
          <UserChatPanel
            key={user.id}
            user={user}
            onRemove={users.length > 1 ? () => removeUser(user.id) : undefined}
          />
        ))}
      </div>
    </div>
  )
}
