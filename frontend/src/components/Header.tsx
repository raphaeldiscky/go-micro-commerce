import { APP_CONFIG, NAVIGATION_ITEMS, PROFILE_IMAGE_URL } from '@/constants'
import { useIsAuthenticated, useLogout, useUser } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'
import { Link, useRouterState } from '@tanstack/react-router'
import { Github, LogIn, LogOut, Menu, User, UserPlus, X } from 'lucide-react'
import { useState } from 'react'
import { Button } from './ui/button'
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  navigationMenuTriggerStyle,
} from './ui/navigation-menu'

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
        <div className="flex h-16 items-center relative">
          {/* Logo/Brand - Left */}
          <div className="flex items-center space-x-2">
            <Link
              className="flex items-center space-x-2 hover:opacity-80 transition-opacity"
              to="/"
            >
              <img
                alt={APP_CONFIG.BRAND.LOGO_ALT}
                className="h-8 w-8 rounded-lg object-cover"
                src={PROFILE_IMAGE_URL}
              />
              <span className="hidden font-bold sm:inline-block">
                {APP_CONFIG.NAME}
              </span>
            </Link>
          </div>

          {/* Desktop Navigation - Centered */}
          <div className="hidden md:flex absolute left-1/2 transform -translate-x-1/2">
            <NavigationMenu>
              <NavigationMenuList>
                {NAVIGATION_ITEMS.map((item) => (
                  <NavigationMenuItem key={item.path}>
                    <NavigationMenuLink asChild>
                      <Link
                        className={cn(
                          navigationMenuTriggerStyle(),
                          'inline-flex items-center',
                          isActive(item.path)
                            ? 'bg-accent text-accent-foreground'
                            : '',
                        )}
                        to={item.path}
                      >
                        <item.icon className="mr-1 h-4 w-4" />
                        {item.name}
                      </Link>
                    </NavigationMenuLink>
                  </NavigationMenuItem>
                ))}
              </NavigationMenuList>
            </NavigationMenu>
          </div>

          {/* Right side - Auth & GitHub */}
          <div className="hidden md:flex items-center space-x-4 ml-auto">
            {isAuthenticated ? (
              <>
                <span className="text-sm text-muted-foreground">
                  Welcome, {user?.first_name}!
                </span>
                <Button
                  className="flex items-center space-x-1"
                  disabled={logoutMutation.isPending}
                  onClick={handleLogout}
                  size="sm"
                  variant="ghost"
                >
                  <LogOut className="h-4 w-4" />
                  <span>Logout</span>
                </Button>
              </>
            ) : (
              <>
                <Button asChild size="sm" variant="ghost">
                  <Link
                    className="flex items-center space-x-1"
                    to="/auth/login"
                  >
                    <LogIn className="h-4 w-4" />
                    <span>Login</span>
                  </Link>
                </Button>
                <Button asChild size="sm" variant="default">
                  <Link
                    className="flex items-center space-x-1"
                    to="/auth/register"
                  >
                    <UserPlus className="h-4 w-4" />
                    <span>Sign Up</span>
                  </Link>
                </Button>
              </>
            )}

            <Button asChild size="sm" variant="outline">
              <a
                className="flex items-center space-x-1"
                href="{GITHUB_REPO_URL}"
                rel="noopener noreferrer"
                target="_blank"
              >
                <Github className="h-4 w-4" />
                <span>GitHub</span>
              </a>
            </Button>
          </div>

          {/* Mobile menu button */}
          <Button
            className="md:hidden ml-auto"
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

        {/* Mobile Navigation */}
        {isMobileMenuOpen && (
          <div className="md:hidden">
            <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3 border-t">
              {NAVIGATION_ITEMS.map((item) => (
                <Link
                  className={cn(
                    'flex items-center px-3 py-2 rounded-md text-base font-medium transition-colors',
                    isActive(item.path)
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:text-foreground hover:bg-accent',
                  )}
                  key={item.path}
                  onClick={() => setIsMobileMenuOpen(false)}
                  to={item.path}
                >
                  <item.icon className="mr-2 h-4 w-4" />
                  {item.name}
                </Link>
              ))}
              {/* Mobile Auth Section */}
              <div className="px-3 py-2 space-y-2">
                {isAuthenticated ? (
                  <>
                    <div className="flex items-center px-3 py-2 text-sm text-muted-foreground">
                      <User className="mr-2 h-4 w-4" />
                      {user?.first_name} {user?.last_name}
                    </div>
                    <Button
                      className="w-full flex items-center justify-center space-x-1"
                      disabled={logoutMutation.isPending}
                      onClick={handleLogout}
                      size="sm"
                      variant="outline"
                    >
                      <LogOut className="h-4 w-4" />
                      <span>Logout</span>
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
                        to="/auth/login"
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
                        to="/auth/register"
                      >
                        <UserPlus className="h-4 w-4" />
                        <span>Sign Up</span>
                      </Link>
                    </Button>
                  </>
                )}

                <Button asChild className="w-full" size="sm" variant="outline">
                  <a
                    className="flex items-center justify-center space-x-1"
                    href="{GITHUB_REPO_URL}"
                    rel="noopener noreferrer"
                    target="_blank"
                  >
                    <Github className="h-4 w-4" />
                    <span>GitHub</span>
                  </a>
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </header>
  )
}
