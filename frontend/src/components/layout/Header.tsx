import {
  APP_CONFIG,
  EXPLORE_ITEMS,
  PATH,
  PATH_AUTH,
  PATH_ROOT,
} from '@/constants'
import { useIsAuthenticated, useLogout, useUser } from '@/hooks/auth'
import { cn } from '@/lib/utils'
import { Link, useRouterState } from '@tanstack/react-router'
import {
  ChevronDown,
  Home,
  Info,
  LogIn,
  LogOut,
  Menu,
  Settings,
  User,
  UserPlus,
  X,
} from 'lucide-react'
import { useState } from 'react'
import { CartDrawer } from '../cart/CartDrawer'
import { CartIcon } from '../cart/CartIcon'
import { ChatIcon } from '../chat/ChatIcon'
import { NotificationBell } from '../notification/NotificationBell'
import { Avatar, AvatarFallback, AvatarImage } from '../ui/avatar'
import { Button } from '../ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
  navigationMenuTriggerStyle,
} from '../ui/navigation-menu'

export default function Header() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const router = useRouterState()
  const currentPath = router.location.pathname
  const isAuthenticated = useIsAuthenticated()
  const user = useUser()
  const logoutMutation = useLogout()

  const isActive = (path: string) => currentPath === path

  const handleLogout = () => {
    logoutMutation.mutate()
  }

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto px-4">
        <div className="grid grid-cols-3 h-16 items-center">
          {/* Logo/Brand - Left */}
          <div className="flex items-center space-x-2 justify-self-start">
            <Link
              className="flex items-center space-x-2 hover:opacity-80 transition-opacity"
              to={PATH_ROOT.home}
            >
              <span className="hidden font-bold sm:inline-block">
                {APP_CONFIG.NAME}
              </span>
            </Link>
          </div>

          {/* Desktop Navigation - Centered */}
          <div className="hidden md:flex items-center justify-center space-x-6">
            <NavigationMenu viewport={false}>
              <NavigationMenuList>
                {/* Home */}
                <NavigationMenuItem>
                  <NavigationMenuLink asChild>
                    <Link
                      className={cn(
                        navigationMenuTriggerStyle(),
                        'inline-flex items-center',
                        isActive(PATH_ROOT.home)
                          ? 'bg-accent text-accent-foreground'
                          : '',
                      )}
                      to={PATH_ROOT.home}
                    >
                      <Home className="mr-1 h-4 w-4" />
                      Home
                    </Link>
                  </NavigationMenuLink>
                </NavigationMenuItem>

                {/* Explore Dropdown */}
                <NavigationMenuItem>
                  <NavigationMenuTrigger className="inline-flex items-center mt-3">
                    Explore
                  </NavigationMenuTrigger>
                  <NavigationMenuContent>
                    <ul className="grid w-[400px] gap-3 p-4 md:w-[500px] md:grid-cols-2 lg:w-[600px]">
                      {EXPLORE_ITEMS.map((feature) => (
                        <li key={feature.path}>
                          <NavigationMenuLink asChild>
                            <Link
                              className={cn(
                                'block select-none space-y-1 rounded-md p-3 leading-none no-underline outline-none transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground',
                                isActive(feature.path) ? 'bg-accent/50' : '',
                              )}
                              to={feature.path}
                            >
                              <div className="flex items-center space-x-2">
                                <feature.icon className="h-5 w-5" />
                                <div className="text-sm font-medium leading-none">
                                  {feature.name}
                                </div>
                              </div>
                              <p className="line-clamp-2 text-sm leading-snug text-muted-foreground">
                                {feature.description}
                              </p>
                            </Link>
                          </NavigationMenuLink>
                        </li>
                      ))}
                    </ul>
                  </NavigationMenuContent>
                </NavigationMenuItem>

                {/* Services */}
                <NavigationMenuItem>
                  <NavigationMenuLink asChild>
                    <Link
                      className={cn(
                        navigationMenuTriggerStyle(),
                        'inline-flex items-center',
                        isActive(PATH_ROOT.services)
                          ? 'bg-accent text-accent-foreground'
                          : '',
                      )}
                      to={PATH_ROOT.services}
                    >
                      <Settings className="mr-1 h-4 w-4" />
                      Services
                    </Link>
                  </NavigationMenuLink>
                </NavigationMenuItem>

                {/* About */}
                <NavigationMenuItem>
                  <NavigationMenuLink asChild>
                    <Link
                      className={cn(
                        navigationMenuTriggerStyle(),
                        'inline-flex items-center',
                        isActive(PATH_ROOT.about)
                          ? 'bg-accent text-accent-foreground'
                          : '',
                      )}
                      to={PATH_ROOT.about}
                    >
                      <Info className="mr-1 h-4 w-4" />
                      About
                    </Link>
                  </NavigationMenuLink>
                </NavigationMenuItem>
              </NavigationMenuList>
            </NavigationMenu>
          </div>

          {/* Right side - Auth & GitHub / Mobile menu button */}
          <div className="flex items-center space-x-4 justify-self-end">
            {/* Desktop Auth & GitHub */}
            <div className="hidden md:flex items-center space-x-4">
              {isAuthenticated ? (
                <>
                  <div className="flex items-center gap-1">
                    <NotificationBell />
                    <ChatIcon />
                    <CartIcon />
                  </div>

                  {/* Profile Dropdown */}
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        className="flex items-center space-x-2 h-10 px-3"
                        variant="ghost"
                      >
                        <Avatar className="h-8 w-8">
                          <AvatarImage alt={user?.firstName} />
                          <AvatarFallback className="bg-primary text-primary-foreground text-sm font-medium">
                            {user?.firstName.charAt(0).toUpperCase() || 'U'}
                          </AvatarFallback>
                        </Avatar>
                        <ChevronDown className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent
                      className="w-56"
                      align="end"
                      forceMount
                    >
                      <DropdownMenuLabel className="font-normal">
                        <div className="flex flex-col space-y-1">
                          <p className="text-sm font-medium leading-none">
                            {user?.firstName} {user?.lastName}
                          </p>
                          <p className="text-xs leading-none text-muted-foreground">
                            {user?.email}
                          </p>
                        </div>
                      </DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem asChild>
                        <Link
                          className="flex items-center cursor-pointer"
                          to={PATH.account.root}
                        >
                          <User className="mr-2 h-4 w-4" />
                          <span>Account</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        className="cursor-pointer text-red-600 focus:text-red-600"
                        disabled={logoutMutation.isPending}
                        onClick={handleLogout}
                      >
                        <LogOut className="mr-2 h-4 w-4" />
                        <span>Log out</span>
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </>
              ) : (
                <>
                  <Button asChild size="sm" variant="ghost">
                    <Link
                      className="flex items-center space-x-1"
                      to={PATH_AUTH.login}
                    >
                      <LogIn className="h-4 w-4" />
                      <span>Login</span>
                    </Link>
                  </Button>
                  <Button asChild size="sm" variant="default">
                    <Link
                      className="flex items-center space-x-1"
                      to={PATH_AUTH.register}
                    >
                      <UserPlus className="h-4 w-4" />
                      <span>Sign Up</span>
                    </Link>
                  </Button>
                </>
              )}
            </div>

            {/* Mobile menu button */}
            <Button
              className="md:hidden"
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              size="sm"
              variant="ghost"
            >
              {isMobileMenuOpen ? (
                <X className="h-6 w-6" />
              ) : (
                <Menu className="h-6 w-6" />
              )}
            </Button>
          </div>
        </div>

        {/* Mobile Navigation */}
        {isMobileMenuOpen && (
          <div className="md:hidden">
            <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3 border-t">
              {/* Home */}
              <Link
                className={cn(
                  'flex items-center px-3 py-2 rounded-md text-base font-medium transition-colors',
                  isActive('/')
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent',
                )}
                onClick={() => setIsMobileMenuOpen(false)}
                to={PATH_ROOT.home}
              >
                <Home className="mr-2 h-4 w-4" />
                Home
              </Link>

              {/* Explore Section */}
              <div className="px-3 py-2">
                <div className="flex items-center text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                  Explore
                </div>
                <div className="space-y-1 ml-4">
                  {EXPLORE_ITEMS.map((feature) => (
                    <Link
                      className={cn(
                        'flex items-start px-3 py-2 rounded-md text-sm font-medium transition-colors',
                        isActive(feature.path)
                          ? 'bg-primary text-primary-foreground'
                          : 'text-muted-foreground hover:text-foreground hover:bg-accent',
                      )}
                      key={feature.path}
                      onClick={() => setIsMobileMenuOpen(false)}
                      to={feature.path}
                    >
                      <feature.icon className="mr-2 h-4 w-4 mt-0.5 flex-shrink-0" />
                      <div>
                        <div>{feature.name}</div>
                        <div className="text-xs text-muted-foreground mt-0.5">
                          {feature.description}
                        </div>
                      </div>
                    </Link>
                  ))}
                </div>
              </div>

              {/* Services */}
              <Link
                className={cn(
                  'flex items-center px-3 py-2 rounded-md text-base font-medium transition-colors',
                  isActive(PATH_ROOT.services)
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent',
                )}
                onClick={() => setIsMobileMenuOpen(false)}
                to={PATH_ROOT.services}
              >
                <Settings className="mr-2 h-4 w-4" />
                Services
              </Link>

              {/* About */}
              <Link
                className={cn(
                  'flex items-center px-3 py-2 rounded-md text-base font-medium transition-colors',
                  isActive('/about')
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent',
                )}
                onClick={() => setIsMobileMenuOpen(false)}
                to={PATH_ROOT.about}
              >
                <Info className="mr-2 h-4 w-4" />
                About
              </Link>
              {/* Mobile Auth Section */}
              <div className="px-3 py-2 space-y-2">
                {isAuthenticated ? (
                  <>
                    {/* Mobile Profile */}
                    <div className="flex items-center px-3 py-3 space-x-3 border-b">
                      <Avatar className="h-10 w-10">
                        <AvatarImage alt={user?.firstName} />
                        <AvatarFallback className="bg-primary text-primary-foreground text-sm font-medium">
                          {user?.firstName.charAt(0).toUpperCase() || 'U'}
                        </AvatarFallback>
                      </Avatar>
                      <div className="flex flex-col">
                        <p className="text-sm font-medium">
                          {user?.firstName} {user?.lastName}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {user?.email}
                        </p>
                      </div>
                    </div>

                    {/* Mobile Menu Items */}
                    <Link
                      className="flex items-center px-3 py-2 text-sm hover:bg-accent rounded-md transition-colors"
                      onClick={() => setIsMobileMenuOpen(false)}
                      to={PATH.account.root}
                    >
                      <User className="mr-2 h-4 w-4" />
                      <span>Account</span>
                    </Link>

                    <div className="flex items-center justify-center py-3 gap-1 border-t">
                      <NotificationBell />
                      <ChatIcon />
                      <CartIcon />
                    </div>

                    <Button
                      className="w-full flex items-center justify-center space-x-1 text-red-600 hover:text-red-700 hover:bg-red-50"
                      disabled={logoutMutation.isPending}
                      onClick={handleLogout}
                      size="sm"
                      variant="ghost"
                    >
                      <LogOut className="h-4 w-4" />
                      <span>Log out</span>
                    </Button>
                  </>
                ) : (
                  <>
                    <Button
                      asChild
                      className="w-full"
                      size="sm"
                      variant="outline"
                    >
                      <Link
                        className="flex items-center justify-center space-x-1"
                        onClick={() => setIsMobileMenuOpen(false)}
                        to={PATH_AUTH.login}
                      >
                        <LogIn className="h-4 w-4" />
                        <span>Login</span>
                      </Link>
                    </Button>
                    <Button
                      asChild
                      className="w-full"
                      size="sm"
                      variant="default"
                    >
                      <Link
                        className="flex items-center justify-center space-x-1"
                        onClick={() => setIsMobileMenuOpen(false)}
                        to={PATH_AUTH.register}
                      >
                        <UserPlus className="h-4 w-4" />
                        <span>Sign Up</span>
                      </Link>
                    </Button>
                  </>
                )}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Cart Drawer */}
      <CartDrawer />
    </header>
  )
}
