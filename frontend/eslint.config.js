//  @ts-check

import { tanstackConfig } from '@tanstack/eslint-config'

export default [
  ...tanstackConfig,
  {
    rules: {
      'import/order': [
        'error',
        {
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
          groups: [
            'builtin', // fs, path, http…
            'external', // react, lodash…
            'internal', // your aliases like @/*
            'parent', // ../*
            'sibling', // ./sibling
            'index', // ./ (index file)
          ],
          'newlines-between': 'never', // change to 'always' if you want spacing between groups
        },
      ],
    },
  },
]
