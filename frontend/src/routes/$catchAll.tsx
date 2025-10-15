import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/$catchAll')({
  beforeLoad: () => {
    // Redirect any unknown path to the 404 page
    throw redirect({
      to: '/404',
    })
  },
})