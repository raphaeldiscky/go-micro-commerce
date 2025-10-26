import tailwindcss from '@tailwindcss/vite'
import tanstackRouter from '@tanstack/router-plugin/vite'
import viteReact from '@vitejs/plugin-react'
import fs from 'node:fs'
import path, { resolve } from 'node:path'
import { defineConfig } from 'vite'

const isLocal = process.env.VITE_ENVIRONMENT === 'dev'
const keyPath = path.resolve(__dirname, 'go.micro.commerce-key.pem')
const certPath = path.resolve(__dirname, 'go.micro.commerce.pem')

// Only enable HTTPS if certs exist and we're running locally
const httpsConfig =
  isLocal && fs.existsSync(keyPath) && fs.existsSync(certPath)
    ? {
        key: fs.readFileSync(keyPath),
        cert: fs.readFileSync(certPath),
      }
    : undefined

export default defineConfig({
  server: {
    allowedHosts: ['go.micro.commerce'],
    https: httpsConfig,
    port: 3031,
    proxy: {
      '/graph': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
        ws: true,
      },
      '/product.v1.ProductService': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
  plugins: [
    tanstackRouter({ autoCodeSplitting: true }),
    viteReact(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
})
