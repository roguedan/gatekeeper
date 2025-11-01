import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-router-dom',
      'wagmi',
      'viem',
      '@wagmi/core',
      '@wagmi/connectors',
      '@rainbow-me/rainbowkit',
      '@tanstack/react-query',
      'siwe',
    ],
    esbuildOptions: {
      // Use target that matches container capabilities
      target: 'es2020',
    },
  },
  server: {
    port: 3000,
    watch: {
      usePolling: true,
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/auth': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          'vendor-web3': ['wagmi', 'viem', '@rainbow-me/rainbowkit', 'siwe'],
          'vendor-query': ['@tanstack/react-query'],
        },
      },
    },
  },
})
