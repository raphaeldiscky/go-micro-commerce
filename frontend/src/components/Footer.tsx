import { Link } from '@tanstack/react-router'
import { Github, Heart, Linkedin } from 'lucide-react'

export default function Footer() {
  const currentYear = new Date().getFullYear()

  const technologies = [
    { name: 'Go', url: 'https://golang.org' },
    { name: 'React', url: 'https://reactjs.org' },
    { name: 'PostgreSQL', url: 'https://postgresql.org' },
    { name: 'Redis', url: 'https://redis.io' },
    { name: 'Kafka', url: 'https://kafka.apache.org' },
    { name: 'Docker', url: 'https://docker.com' },
  ]

  const quickLinks = [
    { name: 'Home', path: '/' },
    { name: 'Services', path: '/services' },
    { name: 'Chat Demo', path: '/chat' },
    { name: 'About', path: '/about' },
  ]

  return (
    <footer className="bg-gray-900 text-white">
      <div className="container mx-auto px-4 py-12">
        <div className="grid gap-8 md:grid-cols-4">
          {/* Brand Section */}
          <div className="md:col-span-2">
            <div className="flex items-center space-x-2 mb-4">
              <div className="h-8 w-8 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 flex items-center justify-center text-white font-bold text-sm">
                GM
              </div>
              <span className="font-bold text-xl">Go Micro Commerce</span>
            </div>
            <p className="text-gray-300 mb-4 max-w-md">
              A modern distributed systems architecture built with Go
              microservices, demonstrating advanced patterns and technologies
              for educational purposes.
            </p>
            <div className="flex space-x-4">
              <a
                href="https://github.com/yourusername/go-micro-commerce"
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-300 hover:text-white transition-colors"
                aria-label="GitHub"
              >
                <Github className="h-6 w-6" />
              </a>
              <a
                href="https://linkedin.com/in/yourprofile"
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-300 hover:text-white transition-colors"
                aria-label="LinkedIn"
              >
                <Linkedin className="h-6 w-6" />
              </a>
            </div>
          </div>

          {/* Quick Links */}
          <div>
            <h3 className="font-semibold mb-4">Quick Links</h3>
            <ul className="space-y-2">
              {quickLinks.map((link) => (
                <li key={link.path}>
                  <Link
                    to={link.path}
                    className="text-gray-300 hover:text-white transition-colors text-sm"
                  >
                    {link.name}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Technologies */}
          <div>
            <h3 className="font-semibold mb-4">Built With</h3>
            <ul className="space-y-2">
              {technologies.map((tech) => (
                <li key={tech.name}>
                  <a
                    href={tech.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-gray-300 hover:text-white transition-colors text-sm"
                  >
                    {tech.name}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Bottom Section */}
        <div className="border-t border-gray-800 mt-8 pt-8 flex flex-col md:flex-row justify-between items-center">
          <div className="text-gray-400 text-sm">
            © {currentYear} Go Micro Commerce. Built for educational purposes.
          </div>
          <div className="text-gray-400 text-sm mt-4 md:mt-0">
            <span className="flex items-center space-x-1">
              <span>Made with</span>
              <Heart className="h-4 w-4 text-red-500 fill-current" />
              <span>by Raphael Discky</span>
            </span>
          </div>
        </div>
      </div>
    </footer>
  )
}
