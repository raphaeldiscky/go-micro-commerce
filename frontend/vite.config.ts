import tailwindcss from '@tailwindcss/vite'
import tanstackRouter from '@tanstack/router-plugin/vite'
import viteReact from '@vitejs/plugin-react'
import fs from 'node:fs'
import path, { resolve } from 'node:path'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    allowedHosts: ['go.micro.commerce'],
    https: {
      key: fs.readFileSync(
        path.resolve(__dirname, 'go.micro.commerce-key.pem'),
      ),
      cert: fs.readFileSync(path.resolve(__dirname, 'go.micro.commerce.pem')),
    },
    port: 3031, // your Vite dev port
    proxy: {
      '/graph': {
        target: 'http://localhost:8080', // your GraphQL backend
        changeOrigin: true,
        secure: false,
        ws: true, // Enable WebSocket proxy
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
