function path(root: string, sublink: string) {
  return `${root}${sublink}`
}

const ROOTS_AUTH = '/auth'
const ROOTS = ''

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
  orders: {
    root: path(ROOTS, '/orders'),
  },
  order: {
    root: path(ROOTS, '/order'),
    pendingPayment: (paymentId: string) =>
      path(ROOTS, `/order/pending-payment/${paymentId}`),
    $pendingPayment: '/order/pending-payment/$paymentId' as const,
  },
}
