import {
  GITHUB_REPO_URL,
  KEY_FEATURES,
  PAGE_CONTENT,
  PATH_ROOT,
  TECHNICAL_GOALS,
} from '@/constants'
import type { FileRoutesByPath } from '@tanstack/react-router'
import { createFileRoute } from '@tanstack/react-router'
import {
  BookOpen,
  Building2,
  CheckCircle,
  FileText,
  Rocket,
  Sparkles,
  Target,
} from 'lucide-react'
import { Button } from '../components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../components/ui/card'

export const Route = createFileRoute(PATH_ROOT.about as keyof FileRoutesByPath)(
  {
    component: AboutPage,
  },
)

function AboutPage() {
  // Data imported from constants

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-white dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-12">
        {/* Header */}
        <div className="text-center mb-16">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-4">
            <BookOpen className="inline-block mr-2" size={28} />{' '}
            {PAGE_CONTENT.ABOUT.TITLE}
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            {PAGE_CONTENT.ABOUT.HERO_DESCRIPTION}
          </p>
        </div>

        {/* Project Overview */}
        <section className="mb-16">
          <Card className="max-w-4xl mx-auto">
            <CardHeader>
              <CardTitle className="text-2xl text-center">
                <Target className="inline-block mr-2" size={24} />{' '}
                {PAGE_CONTENT.ABOUT.PROJECT_MISSION.TITLE}
              </CardTitle>
            </CardHeader>
            <CardContent className="prose dark:prose-invert max-w-none">
              <p className="text-gray-700 dark:text-gray-300 text-center text-lg leading-relaxed">
                {PAGE_CONTENT.ABOUT.PROJECT_MISSION.DESCRIPTION}
              </p>
              <div className="mt-8 bg-blue-50 dark:bg-blue-900/20 p-6 rounded-lg border-l-4 border-blue-500">
                <p className="text-gray-700 dark:text-gray-300 mb-0">
                  <strong>Note:</strong>{' '}
                  {PAGE_CONTENT.ABOUT.PROJECT_MISSION.NOTE}
                </p>
              </div>
            </CardContent>
          </Card>
        </section>

        {/* Key Features */}
        <section className="mb-16">
          <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-8 text-center">
            <Sparkles className="inline-block mr-2" size={28} />{' '}
            {PAGE_CONTENT.ABOUT.KEY_FEATURES_TITLE}
          </h2>
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {KEY_FEATURES.map((feature) => (
              <Card
                className="hover:shadow-lg transition-shadow"
                key={feature.title}
              >
                <CardHeader>
                  <div className="flex items-center space-x-2 mb-2">
                    <feature.icon className="text-blue-500" size={24} />
                    <CardTitle className="text-lg">{feature.title}</CardTitle>
                  </div>
                  <CardDescription>{feature.description}</CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        </section>

        {/* Technical Goals */}
        <section className="mb-16">
          <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-8 text-center">
            <Target className="inline-block mr-2" size={28} />{' '}
            {PAGE_CONTENT.ABOUT.TECHNICAL_GOALS_TITLE}
          </h2>
          <div className="grid gap-6 md:grid-cols-2">
            {TECHNICAL_GOALS.map((section) => (
              <Card className="h-fit" key={section.category}>
                <CardHeader>
                  <CardTitle className="text-xl">{section.category}</CardTitle>
                </CardHeader>
                <CardContent>
                  <ul className="space-y-2">
                    {section.goals.map((goal, index) => (
                      <li className="flex items-start space-x-2" key={index}>
                        <CheckCircle
                          className="text-green-500 mt-1"
                          size={16}
                        />
                        <span className="text-gray-700 dark:text-gray-300 text-sm">
                          {goal}
                        </span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        {/* Architecture Highlights */}
        <section className="mb-16">
          <Card className="max-w-5xl mx-auto bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20">
            <CardHeader>
              <CardTitle className="text-2xl text-center">
                <Building2 className="inline-block mr-2" size={24} />{' '}
                {PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS.TITLE}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid gap-6 md:grid-cols-2">
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                    {
                      PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS
                        .SERVICE_ARCHITECTURE
                    }
                  </h3>
                  <ul className="space-y-2 text-sm text-gray-700 dark:text-gray-300">
                    {PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS.SERVICE_ITEMS.map(
                      (item, index) => (
                        <li key={index}>• {item}</li>
                      ),
                    )}
                  </ul>
                </div>
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                    {
                      PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS
                        .ADVANCED_PATTERNS
                    }
                  </h3>
                  <ul className="space-y-2 text-sm text-gray-700 dark:text-gray-300">
                    {PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS.ADVANCED_ITEMS.map(
                      (item, index) => (
                        <li key={index}>• {item}</li>
                      ),
                    )}
                  </ul>
                </div>
              </div>

              <div className="pt-6 border-t border-gray-200 dark:border-gray-700">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  {
                    PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS
                      .INFRASTRUCTURE_TITLE
                  }
                </h3>
                <div className="grid gap-4 md:grid-cols-3">
                  <div className="space-y-2">
                    <h4 className="font-medium text-gray-900 dark:text-white">
                      {
                        PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS.DATA_LAYER
                          .TITLE
                      }
                    </h4>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {
                        PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS.DATA_LAYER
                          .DESCRIPTION
                      }
                    </p>
                  </div>
                  <div className="space-y-2">
                    <h4 className="font-medium text-gray-900 dark:text-white">
                      {
                        PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS.MESSAGING
                          .TITLE
                      }
                    </h4>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {
                        PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS.MESSAGING
                          .DESCRIPTION
                      }
                    </p>
                  </div>
                  <div className="space-y-2">
                    <h4 className="font-medium text-gray-900 dark:text-white">
                      {
                        PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS.MONITORING
                          .TITLE
                      }
                    </h4>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {
                        PAGE_CONTENT.ABOUT.ARCHITECTURE_HIGHLIGHTS.MONITORING
                          .DESCRIPTION
                      }
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </section>

        {/* Call to Action */}
        <section className="text-center">
          <Card className="max-w-2xl mx-auto bg-gradient-to-r from-green-50 to-blue-50 dark:from-green-900/20 dark:to-blue-900/20">
            <CardContent className="p-8">
              <h3 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
                <Rocket className="inline-block mr-2" size={24} />{' '}
                {PAGE_CONTENT.ABOUT.CTA.TITLE}
              </h3>
              <p className="text-gray-700 dark:text-gray-300 mb-6">
                {PAGE_CONTENT.ABOUT.CTA.DESCRIPTION}
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Button asChild size="lg">
                  <a
                    className="flex items-center space-x-2"
                    href={GITHUB_REPO_URL}
                    rel="noopener noreferrer"
                    target="_blank"
                  >
                    <svg
                      className="h-5 w-5"
                      fill="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                    </svg>
                    <span>{PAGE_CONTENT.ABOUT.CTA.VIEW_SOURCE}</span>
                  </a>
                </Button>
                <Button asChild size="lg" variant="outline">
                  <a className="flex items-center space-x-2" href="/services">
                    <FileText size={16} />
                    <span>{PAGE_CONTENT.ABOUT.CTA.BROWSE_SERVICES}</span>
                  </a>
                </Button>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
    </div>
  )
}
