import {
  HERO_CONTENT,
  PAGE_CONTENT,
  PATH,
  PATH_ROOT,
  SERVICES,
  TECHNOLOGIES,
} from '@/constants'
import type { FileRoutesByPath } from '@tanstack/react-router'
import { createFileRoute, Link } from '@tanstack/react-router'
import { BookOpen, Building, MessageCircle, Wrench } from 'lucide-react'
import { Button } from '../components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../components/ui/card'

export const Route = createFileRoute(PATH_ROOT.home as keyof FileRoutesByPath)({
  component: HomePage,
})

function HomePage() {
  // Data imported from constants

  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
        <div className="container mx-auto px-4 py-24 sm:py-32">
          <div className="text-center">
            <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-6xl">
              {HERO_CONTENT.TITLE}
            </h1>
            <p className="mt-6 text-lg leading-8 text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
              {HERO_CONTENT.DESCRIPTION}
            </p>
            <div className="mt-10 flex items-center justify-center gap-x-6">
              <Button asChild size="lg">
                <Link
                  className="flex items-center gap-2"
                  to={PATH_ROOT.services}
                >
                  <Wrench className="h-5 w-5" />
                  {HERO_CONTENT.PRIMARY_CTA}
                </Link>
              </Button>
              <Button asChild size="lg" variant="outline">
                <Link
                  className="flex items-center gap-2"
                  to={PATH.products.root}
                >
                  <MessageCircle className="h-5 w-5" />
                  {HERO_CONTENT.SECONDARY_CTA}
                </Link>
              </Button>
            </div>
          </div>
        </div>
        {/* Decorative background elements */}
        <div className="absolute inset-0 -z-10 h-full w-full bg-white dark:bg-gray-900 [background:radial-gradient(125%_125%_at_50%_10%,#fff_40%,#63e_100%)] dark:[background:radial-gradient(125%_125%_at_50%_10%,#000_40%,#63e_100%)] opacity-20"></div>
      </section>

      {/* Services Grid */}
      <section className="py-16 bg-white dark:bg-gray-800">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
              {PAGE_CONTENT.HOME.SERVICES_SECTION.TITLE}
            </h2>
            <p className="text-lg text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
              {PAGE_CONTENT.HOME.SERVICES_SECTION.DESCRIPTION}
            </p>
          </div>

          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {SERVICES.map((service) => (
              <Card
                className="hover:shadow-lg transition-shadow"
                key={service.name}
              >
                <CardHeader>
                  <div className="flex items-center space-x-2">
                    <service.icon className="h-6 w-6 text-blue-600 dark:text-blue-400" />
                    <CardTitle className="text-xl">{service.name}</CardTitle>
                  </div>
                  <CardDescription>{service.description}</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex flex-wrap gap-2">
                    {service.features.map((feature) => (
                      <span
                        className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                        key={feature}
                      >
                        {feature}
                      </span>
                    ))}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Technology Stack */}
      <section className="py-16 bg-gray-50 dark:bg-gray-900">
        <div className="container mx-auto px-4">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
              {PAGE_CONTENT.HOME.TECH_STACK.TITLE}
            </h2>
            <p className="text-lg text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
              {PAGE_CONTENT.HOME.TECH_STACK.DESCRIPTION}
            </p>
          </div>

          <Card className="p-8 max-w-5xl mx-auto">
            <div className="grid gap-8 md:grid-cols-2">
              <div>
                <div className="grid gap-3 grid-cols-2">
                  {TECHNOLOGIES.slice(0, 4).map((tech) => (
                    <div
                      className="flex items-center space-x-3 p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors"
                      key={tech.name}
                    >
                      <tech.icon className="h-8 w-8 text-foreground flex-shrink-0" />
                      <div>
                        <h4 className="text-sm font-medium text-foreground">
                          {tech.name}
                        </h4>
                        <p className="text-xs text-muted-foreground">
                          {tech.description}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
              <div>
                <div className="grid gap-3 grid-cols-2">
                  {TECHNOLOGIES.slice(4, 8).map((tech) => (
                    <div
                      className="flex items-center space-x-3 p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors"
                      key={tech.name}
                    >
                      <tech.icon className="h-8 w-8 text-foreground flex-shrink-0" />
                      <div>
                        <h4 className="text-sm font-medium text-foreground">
                          {tech.name}
                        </h4>
                        <p className="text-xs text-muted-foreground">
                          {tech.description}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </Card>
        </div>
      </section>

      {/* Quick Links */}
      <section className="py-16 bg-white dark:bg-gray-800">
        <div className="container mx-auto px-4">
          <div className="text-center">
            <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-8">
              {PAGE_CONTENT.HOME.GET_STARTED.TITLE}
            </h2>
            <div className="grid gap-6 md:grid-cols-3 max-w-4xl mx-auto">
              <Card className="text-center hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="mb-2">
                    <Building className="h-10 w-10 mx-auto" />
                  </div>
                  <CardTitle>
                    {PAGE_CONTENT.HOME.GET_STARTED.ARCHITECTURE.TITLE}
                  </CardTitle>
                  <CardDescription>
                    {PAGE_CONTENT.HOME.GET_STARTED.ARCHITECTURE.DESCRIPTION}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button asChild className="w-full" variant="outline">
                    <Link to={PATH_ROOT.services}>
                      {PAGE_CONTENT.HOME.GET_STARTED.ARCHITECTURE.CTA}
                    </Link>
                  </Button>
                </CardContent>
              </Card>

              <Card className="text-center hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="mb-2">
                    <MessageCircle className="h-10 w-10 mx-auto" />
                  </div>
                  <CardTitle>
                    {PAGE_CONTENT.HOME.GET_STARTED.PRODUCT.TITLE}
                  </CardTitle>
                  <CardDescription>
                    {PAGE_CONTENT.HOME.GET_STARTED.PRODUCT.DESCRIPTION}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button asChild className="w-full">
                    <Link to={PATH.chat.root}>
                      {PAGE_CONTENT.HOME.GET_STARTED.PRODUCT.CTA}
                    </Link>
                  </Button>
                </CardContent>
              </Card>

              <Card className="text-center hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="mb-2">
                    <BookOpen className="h-10 w-10 mx-auto" />
                  </div>
                  <CardTitle>
                    {PAGE_CONTENT.HOME.GET_STARTED.LEARN_MORE.TITLE}
                  </CardTitle>
                  <CardDescription>
                    {PAGE_CONTENT.HOME.GET_STARTED.LEARN_MORE.DESCRIPTION}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button asChild className="w-full" variant="outline">
                    <Link to={PATH_ROOT.about}>
                      {PAGE_CONTENT.HOME.GET_STARTED.LEARN_MORE.CTA}
                    </Link>
                  </Button>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}
