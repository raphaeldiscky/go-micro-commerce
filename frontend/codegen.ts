import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  overwrite: true,
  schema: '../graphql-gateway/supergraph-schema.graphql',
  documents: ['src/**/*.{ts,tsx}'],
  ignoreNoDocuments: true,
  hooks: {
    afterOneFileWrite: [
      // $1 will be the file path
      'bash -c \'if grep -q "__generated__" <<< "$1"; then sed -i "1i/* eslint-disable */" "$1"; fi\' _',
    ],
  },
  generates: {
    // Base types file
    './src/types/__generated__/graphql.ts': {
      plugins: ['typescript'],
      config: {
        scalars: {
          UUID: 'string',
          Time: 'string',
          Decimal: 'string',
        },
      },
    },
    // Operation types near each component file
    './src/': {
      preset: 'near-operation-file',
      presetConfig: {
        baseTypesPath: './types/__generated__/graphql.ts',
      },
      plugins: ['typescript-operations'],
      config: {
        avoidOptionals: {
          field: true,
          inputValue: false,
        },
        defaultScalarType: 'unknown',
        nonOptionalTypename: true,
        skipTypeNameForRoot: true,
        scalars: {
          UUID: 'string',
          Time: 'string',
          Decimal: 'string',
        },
      },
    },
  },
}

export default config
