import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/dashboard/fulfillments/')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/dashboard/fulfillment/"!</div>
}
