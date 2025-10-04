function path(root: string, sublink: string) {
  return `${root}${sublink}`
}

const ROOTS_AUTH = '/auth'
const ROOTS_DASHBOARD = '/dashboard'

export const PATH_ROOT = {
  home: '/',
  comingSoon: '/coming-soon',
  maintenance: '/maintenance',
  about: '/about-us',
  page403: '/403',
  page404: '/404',
  page500: '/500',
}

export const PATH_AUTH = {
  root: ROOTS_AUTH,
  login: path(ROOTS_AUTH, '/login'),
  register: path(ROOTS_AUTH, '/register'),
}

export const PATH_DASHBOARD = {
  root: ROOTS_DASHBOARD,
  products: {
    root: path(ROOTS_DASHBOARD, '/products'),
    detail: (id: string) => path(ROOTS_DASHBOARD, `/products/${id}`),
  },
  services: {
    root: path(ROOTS_DASHBOARD, '/services'),
    detail: (id: string) => path(ROOTS_DASHBOARD, `/services/${id}`),
  },
  chat: {
    root: path(ROOTS_DASHBOARD, '/chat'),
    detail: (id: string) => path(ROOTS_DASHBOARD, `/chat/${id}`),
  },
}
