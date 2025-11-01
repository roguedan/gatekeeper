import { getDefaultConfig } from '@rainbow-me/rainbowkit'
import { supportedChains } from './chains'
import { env } from './env'

export const wagmiConfig = getDefaultConfig({
  appName: env.appName,
  projectId: env.walletConnectProjectId || 'demo-project-id',
  chains: supportedChains as any,
  ssr: false,
})
