import { createFileRoute, useNavigate } from '@tanstack/react-router'

export const Route = createFileRoute('/features/account/$accountId')({
  component: AccountRedirect,
})

function AccountRedirect() {
  const navigate = useNavigate()

  // Redirect to main account page
  navigate({ to: '/features/account', replace: true })

  return (
    <div className="flex items-center justify-center min-h-[400px]">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>
  )
}
