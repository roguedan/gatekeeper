/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string
  readonly VITE_CHAIN_ID: string
  readonly VITE_ETHEREUM_RPC_URL: string
  readonly VITE_WALLETCONNECT_PROJECT_ID: string
  readonly VITE_APP_NAME: string
  readonly VITE_APP_DOMAIN: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
