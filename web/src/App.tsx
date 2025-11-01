import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { WagmiProvider } from 'wagmi'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RainbowKitProvider } from '@rainbow-me/rainbowkit'
import { AuthProvider } from './contexts'
import { MainLayout } from './components/layout'
import { AuthGuard } from './components/auth'
import { Home, Dashboard, APIKeys, TokenGating } from './pages'
import { wagmiConfig } from './config'

import '@rainbow-me/rainbowkit/styles.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
})

function App() {
  return (
    <WagmiProvider config={wagmiConfig}>
      <QueryClientProvider client={queryClient}>
        <RainbowKitProvider>
          <AuthProvider>
            <BrowserRouter>
              <MainLayout>
                <Routes>
                  <Route path="/" element={<Home />} />
                  <Route
                    path="/dashboard"
                    element={
                      <AuthGuard>
                        <Dashboard />
                      </AuthGuard>
                    }
                  />
                  <Route
                    path="/api-keys"
                    element={
                      <AuthGuard>
                        <APIKeys />
                      </AuthGuard>
                    }
                  />
                  <Route
                    path="/token-gating"
                    element={
                      <AuthGuard>
                        <TokenGating />
                      </AuthGuard>
                    }
                  />
                </Routes>
              </MainLayout>
            </BrowserRouter>
          </AuthProvider>
        </RainbowKitProvider>
      </QueryClientProvider>
    </WagmiProvider>
  )
}

export default App
