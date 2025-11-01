import { Chain } from 'wagmi/chains'
import { mainnet, sepolia, polygon, optimism, arbitrum } from 'wagmi/chains'

export const supportedChains: Chain[] = [
  mainnet,
  sepolia,
  polygon,
  optimism,
  arbitrum,
]

export const defaultChain = mainnet
