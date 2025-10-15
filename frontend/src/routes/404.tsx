import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { PATH_AUTH, PATH_FEATURES, PATH_ROOT } from '@/constants/routes'
import { createFileRoute, Link } from '@tanstack/react-router'
import {
  ArrowLeft,
  ExternalLink,
  Home,
  Mail,
  MessageCircle,
  Phone,
  Search,
  ShoppingBag,
  User,
} from 'lucide-react'
import { useState } from 'react'

export const Route = createFileRoute('/404')({
  component: RouteComponent,
})

function RouteComponent() {
  const [searchQuery, setSearchQuery] = useState('')

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (searchQuery.trim()) {
      // Redirect to products page with search query
      window.location.href = `${PATH_FEATURES.products.root}?search=${encodeURIComponent(searchQuery.trim())}`
    }
  }

  const popularPages = [
    {
      name: 'Home',
      href: PATH_ROOT.home,
      icon: Home,
      description: 'Back to main page',
    },
    {
      name: 'Products',
      href: PATH_FEATURES.products.root,
      icon: ShoppingBag,
      description: 'Browse our catalog',
    },
    {
      name: 'Chat',
      href: PATH_FEATURES.chat.root,
      icon: MessageCircle,
      description: 'Start a conversation',
    },
    {
      name: 'Account',
      href: PATH_FEATURES.account.root,
      icon: User,
      description: 'Manage your profile',
    },
  ]

  const helpOptions = [
    {
      title: 'Contact Support',
      description: 'Get help from our support team',
      icon: Mail,
      action: () => (window.location.href = 'mailto:support@example.com'),
    },
    {
      title: 'Call Us',
      description: 'Speak with our customer service',
      icon: Phone,
      action: () => (window.location.href = 'tel:+1234567890'),
    },
  ]

  return (
    <div className="min-h-screen bg-gray-50/40">
      <div className="flex min-h-screen flex-col">
        {/* Main Content */}
        <main className="flex-1 flex items-center justify-center p-4 sm:p-6 lg:p-8">
          <div className="w-full max-w-4xl">
            {/* 404 Hero Section */}
            <div className="text-center mb-12">
              <div className="mb-8">
                <div className="mx-auto h-32 w-32 rounded-full bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-blue-950 dark:to-indigo-950 flex items-center justify-center shadow-lg">
                  <span className="text-6xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
                    404
                  </span>
                </div>
              </div>

              <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-gray-50 mb-4">
                Page Not Found
              </h1>

              <p className="text-xl text-muted-foreground mb-2 max-w-2xl mx-auto">
                Oops! The page you're looking for doesn't exist or has been
                moved.
              </p>

              <p className="text-muted-foreground max-w-2xl mx-auto">
                Don't worry, let's help you get back on track. Try searching for
                what you need or choose from the options below.
              </p>
            </div>

            {/* Search Section */}
            <Card className="mb-8">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Search className="h-5 w-5" />
                  Search for what you need
                </CardTitle>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSearch} className="flex gap-3">
                  <div className="flex-1">
                    <Label htmlFor="search" className="sr-only">
                      Search
                    </Label>
                    <Input
                      id="search"
                      type="search"
                      placeholder="Search products, articles, or help topics..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      className="h-11"
                    />
                  </div>
                  <Button type="submit" size="lg" className="h-11 px-6">
                    <Search className="h-4 w-4 mr-2" />
                    Search
                  </Button>
                </form>
              </CardContent>
            </Card>

            <div className="grid gap-6 lg:grid-cols-2">
              {/* Popular Pages */}
              <Card>
                <CardHeader>
                  <CardTitle>Popular Pages</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  {popularPages.map((page) => {
                    const Icon = page.icon
                    return (
                      <Link
                        key={page.href}
                        to={page.href}
                        className="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors group"
                      >
                        <div className="h-10 w-10 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center group-hover:bg-gray-200 dark:group-hover:bg-gray-700 transition-colors">
                          <Icon className="h-5 w-5 text-gray-600 dark:text-gray-400" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="font-medium text-gray-900 dark:text-gray-50">
                            {page.name}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {page.description}
                          </div>
                        </div>
                        <ExternalLink className="h-4 w-4 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300" />
                      </Link>
                    )
                  })}
                </CardContent>
              </Card>

              {/* Help & Support */}
              <Card>
                <CardHeader>
                  <CardTitle>Need Help?</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <p className="text-sm text-muted-foreground mb-4">
                    Can't find what you're looking for? Our support team is here
                    to help.
                  </p>

                  {helpOptions.map((option) => {
                    const Icon = option.icon
                    return (
                      <Button
                        key={option.title}
                        variant="outline"
                        onClick={option.action}
                        className="w-full justify-start h-auto p-4"
                      >
                        <div className="flex items-center gap-3">
                          <div className="h-10 w-10 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
                            <Icon className="h-5 w-5 text-gray-600 dark:text-gray-400" />
                          </div>
                          <div className="text-left">
                            <div className="font-medium">{option.title}</div>
                            <div className="text-sm text-muted-foreground">
                              {option.description}
                            </div>
                          </div>
                        </div>
                      </Button>
                    )
                  })}

                  <Separator />

                  <div className="space-y-2">
                    <h4 className="text-sm font-medium">Quick Actions</h4>
                    <div className="space-y-2">
                      {[
                        { href: PATH_AUTH.login, label: 'Sign In' },
                        { href: PATH_AUTH.register, label: 'Create Account' },
                        { href: PATH_ROOT.about, label: 'About Us' },
                      ].map((action) => (
                        <Link
                          key={action.href}
                          to={action.href}
                          className="text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 flex items-center gap-1"
                        >
                          {action.label}
                          <ArrowLeft className="h-3 w-3 rotate-180" />
                        </Link>
                      ))}
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Back Button */}
            <div className="mt-8 text-center">
              <Button
                variant="ghost"
                onClick={() => window.history.back()}
                className="text-muted-foreground hover:text-foreground"
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                Go Back to Previous Page
              </Button>
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}
