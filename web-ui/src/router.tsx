import {
  createRootRoute,
  createRoute,
  createRouter,
} from '@tanstack/react-router'
import {
  AdminShell,
  EmailAccountsPage,
  FortySixElksPage,
  OverviewPage,
  SendMessagePage,
  ServicesPage,
  UsersPage,
} from './pages'

const rootRoute = createRootRoute({
  component: AdminShell,
})

const overviewRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: OverviewPage,
})

const servicesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: 'services',
  component: ServicesPage,
})

const emailAccountsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: 'email-accounts',
  component: EmailAccountsPage,
})

const fortySixElksRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '46elks',
  component: FortySixElksPage,
})

const sendMessageRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: 'send-message',
  component: SendMessagePage,
})

const usersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: 'users',
  component: UsersPage,
})

const routeTree = rootRoute.addChildren([
  overviewRoute,
  servicesRoute,
  usersRoute,
  emailAccountsRoute,
  fortySixElksRoute,
  sendMessageRoute,
])

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
