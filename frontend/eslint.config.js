//  @ts-check

import { tanstackConfig } from '@tanstack/eslint-config'
import perfectionist from 'eslint-plugin-perfectionist'

export default [
  ...tanstackConfig,
  {
    plugins: {
      perfectionist,
    },
    rules: {
      'import/order': 'off',
      'perfectionist/sort-imports': [
        'error',
        {
          type: 'alphabetical',
          order: 'asc',
          newlinesBetween: 'never',
          internalPattern: ['^@/'],
          groups: [
            'type-import',
            'internal',
            ['builtin', 'external'],
            ['parent', 'sibling', 'index'],
            'unknown',
          ],
        },
      ],
      'sort-imports': 'off',
    },
  },
]
