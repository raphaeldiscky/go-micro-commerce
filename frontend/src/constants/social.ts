import { Github, Linkedin } from 'lucide-react'
import type { SocialLink } from './types'
import { GITHUB_REPO_URL, LINKEDIN_PROFILE_URL } from './urls'

// Social media links
export const SOCIAL_LINKS: Array<SocialLink> = [
  {
    name: 'GitHub',
    url: GITHUB_REPO_URL,
    icon: Github,
    ariaLabel: 'GitHub',
  },
  {
    name: 'LinkedIn',
    url: LINKEDIN_PROFILE_URL,
    icon: Linkedin,
    ariaLabel: 'LinkedIn',
  },
]
