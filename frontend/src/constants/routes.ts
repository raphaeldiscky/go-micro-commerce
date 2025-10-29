function path(root: string, sublink: string) {
  return `${root}${sublink}`
}

const ROOTS = ''
const ROOTS_AUTH = '/auth'
const ROOTS_DASHBOARD = '/dashboard'

export const PATH_ROOT = {
  home: '/',
  comingSoon: '/coming-soon',
  maintenance: '/maintenance',
  about: '/about',
  services: 'services',
  page403: '/403',
  page404: '/404',
  page500: '/500',
}

export const PATH_AUTH = {
  root: ROOTS_AUTH,
  login: path(ROOTS_AUTH, '/login'),
  register: path(ROOTS_AUTH, '/register'),
}

export const PATH = {
  root: ROOTS,
  products: {
    root: path(ROOTS, '/products'),
    detail: (id: string) => path(ROOTS, `/products/${id}`),
  },
  chat: {
    root: path(ROOTS, '/chat'),
    detail: (id: string) => path(ROOTS, `/chat/${id}`),
    $conversationId: '/chat/$conversationId' as const,
  },
  checkout: {
    root: path(ROOTS, '/checkout'),
    detail: (id: string) => path(ROOTS, `/checkout/${id}`),
    $checkoutId: '/checkout/$checkoutId' as const,
  },
  account: {
    root: path(ROOTS, '/account'),
    detail: (id: string) => path(ROOTS, `/account/${id}`),
  },
  payment: {
    root: path(ROOTS, '/payment'),
    detail: (orderId: string) => path(ROOTS, `/payment/${orderId}`),
    success: (orderId: string) => path(ROOTS, `/payment/${orderId}/success`),
    $orderId: '/payment/$orderId' as const,
  },
  orders: {
    root: path(ROOTS, '/orders'),
    detail: (orderId: string) => path(ROOTS, `/orders/${orderId}`),
    $orderId: '/orders/$orderId' as const,
  },
}

export const PATH_DASHBOARD = {
  root: ROOTS_DASHBOARD,
  analytics: path(ROOTS_DASHBOARD, '/analytics'),
  revenue: path(ROOTS_DASHBOARD, '/revenue'),
  orders: path(ROOTS_DASHBOARD, '/orders'),
  products: path(ROOTS_DASHBOARD, '/products'),
  users: path(ROOTS_DASHBOARD, '/users'),
  fulfillments: {
    root: path(ROOTS_DASHBOARD, '/fulfillments'),
    detail: (id: string) => path(ROOTS_DASHBOARD, `/fulfillments/${id}`),
  },
}
