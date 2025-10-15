import {
  APP_CONFIG,
  FOOTER_CONTENT,
  QUICK_LINKS,
  SOCIAL_LINKS,
  TECHNOLOGY_LINKS,
} from '@/constants'
import { getCurrentYear } from '@/lib/utils/date'
import { Link } from '@tanstack/react-router'
import { Heart } from 'lucide-react'

export default function Footer() {
  const currentYear = getCurrentYear()

  // Data imported from constants

  return (
    <footer className="bg-gray-900 text-white">
      <div className="container mx-auto px-4 py-12">
        <div className="grid gap-8 md:grid-cols-4">
          {/* Brand Section */}
          <div className="md:col-span-2">
            <div className="flex items-center space-x-2 mb-4">
              <span className="font-bold text-xl">{APP_CONFIG.NAME}</span>
            </div>
            <p className="text-gray-300 mb-4 max-w-md">
              {FOOTER_CONTENT.DESCRIPTION}
            </p>
            <div className="flex space-x-4">
              {SOCIAL_LINKS.map((social) => (
                <a
                  aria-label={social.ariaLabel}
                  className="text-gray-300 hover:text-white transition-colors"
                  href={social.url}
                  key={social.name}
                  rel="noopener noreferrer"
                  target="_blank"
                >
                  <social.icon className="h-6 w-6" />
                </a>
              ))}
            </div>
          </div>

          {/* Quick Links */}
          <div>
            <h3 className="font-semibold mb-4">Quick Links</h3>
            <ul className="space-y-2">
              {QUICK_LINKS.map((link) => (
                <li key={link.path}>
                  <Link
                    className="text-gray-300 hover:text-white transition-colors text-sm"
                    to={link.path}
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
              {TECHNOLOGY_LINKS.map((tech) => (
                <li key={tech.name}>
                  <a
                    className="text-gray-300 hover:text-white transition-colors text-sm"
                    href={tech.url}
                    rel="noopener noreferrer"
                    target="_blank"
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
            © {currentYear} {FOOTER_CONTENT.COPYRIGHT}
          </div>
          <div className="text-gray-400 text-sm mt-4 md:mt-0">
            <span className="flex items-center space-x-1">
              <span>{FOOTER_CONTENT.MADE_WITH_LOVE}</span>
              <Heart className="h-4 w-4 text-red-500 fill-current" />
              <span>{FOOTER_CONTENT.BY_AUTHOR}</span>
            </span>
          </div>
        </div>
      </div>
    </footer>
  )
}
