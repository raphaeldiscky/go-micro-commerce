function path(root: string, sublink: string) {
  return `${root}${sublink}`
}

const ROOTS_AUTH = '/auth'
const ROOTS_FEATURES = '/features'

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

export const PATH_FEATURES = {
  root: ROOTS_FEATURES,
  products: {
    root: path(ROOTS_FEATURES, '/products'),
    detail: (id: string) => path(ROOTS_FEATURES, `/products/${id}`),
  },
  chat: {
    root: path(ROOTS_FEATURES, '/chat'),
    detail: (id: string) => path(ROOTS_FEATURES, `/chat/${id}`),
    $conversationId: '/features/chat/$conversationId' as const,
  },
  checkout: {
    root: path(ROOTS_FEATURES, '/checkout'),
    detail: (id: string) => path(ROOTS_FEATURES, `/checkout/${id}`),
    $checkoutId: '/features/checkout/$checkoutId' as const,
  },
  account: {
    root: path(ROOTS_FEATURES, '/account'),
    detail: (id: string) => path(ROOTS_FEATURES, `/account/${id}`),
  },
  orders: {
    root: path(ROOTS_FEATURES, '/orders'),
  },
  order: {
    root: path(ROOTS_FEATURES, '/order'),
    pendingPayment: (paymentId: string) =>
      path(ROOTS_FEATURES, `/order/pending-payment/${paymentId}`),
    $pendingPayment: '/features/order/pending-payment/$paymentId' as const,
  },
}
