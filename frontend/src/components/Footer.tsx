import { Link } from '@tanstack/react-router'
import { Heart } from 'lucide-react'
import {
  APP_CONFIG,
  FOOTER_CONTENT,
  PROFILE_IMAGE_URL,
  QUICK_LINKS,
  SOCIAL_LINKS,
  TECHNOLOGY_LINKS,
} from '@/constants'

export default function Footer() {
  const currentYear = new Date().getFullYear()

  // Data imported from constants

  return (
    <footer className="bg-gray-900 text-white">
      <div className="container mx-auto px-4 py-12">
        <div className="grid gap-8 md:grid-cols-4">
          {/* Brand Section */}
          <div className="md:col-span-2">
            <div className="flex items-center space-x-2 mb-4">
              <img
                src={PROFILE_IMAGE_URL}
                alt={APP_CONFIG.BRAND.LOGO_ALT}
                className="h-8 w-8 rounded-lg object-cover"
              />
              <span className="font-bold text-xl">{APP_CONFIG.NAME}</span>
            </div>
            <p className="text-gray-300 mb-4 max-w-md">
              {FOOTER_CONTENT.DESCRIPTION}
            </p>
            <div className="flex space-x-4">
              {SOCIAL_LINKS.map((social) => (
                <a
                  key={social.name}
                  href={social.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-gray-300 hover:text-white transition-colors"
                  aria-label={social.ariaLabel}
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
              {TECHNOLOGY_LINKS.map((tech) => (
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
