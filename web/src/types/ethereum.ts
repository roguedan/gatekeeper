export interface ChainConfig {
  id: number
  name: string
  network: string
  nativeCurrency: {
    name: string
    symbol: string
    decimals: number
  }
  rpcUrls: {
    default: {
      http: string[]
    }
    public: {
      http: string[]
    }
  }
  blockExplorers: {
    default: {
      name: string
      url: string
    }
  }
}

export interface ContractConfig {
  address: `0x${string}`
  abi: unknown[]
  chainId: number
}
