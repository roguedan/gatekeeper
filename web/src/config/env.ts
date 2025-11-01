interface Env {
  apiUrl: string
  chainId: number
  ethereumRpcUrl: string
  walletConnectProjectId: string
  appName: string
  appDomain: string
}

const getEnv = (): Env => {
  return {
    apiUrl: import.meta.env.VITE_API_URL || 'http://localhost:8080',
    chainId: parseInt(import.meta.env.VITE_CHAIN_ID || '1'),
    ethereumRpcUrl: import.meta.env.VITE_ETHEREUM_RPC_URL || '',
    walletConnectProjectId: import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || '',
    appName: import.meta.env.VITE_APP_NAME || 'Gatekeeper',
    appDomain: import.meta.env.VITE_APP_DOMAIN || 'localhost:3000',
  }
}

export const env = getEnv()
