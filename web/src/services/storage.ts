const TOKEN_KEY = 'gatekeeper_auth_token'
const ADDRESS_KEY = 'gatekeeper_wallet_address'

export const storage = {
  getToken: (): string | null => {
    return localStorage.getItem(TOKEN_KEY)
  },

  setToken: (token: string): void => {
    localStorage.setItem(TOKEN_KEY, token)
  },

  removeToken: (): void => {
    localStorage.removeItem(TOKEN_KEY)
  },

  getAddress: (): string | null => {
    return localStorage.getItem(ADDRESS_KEY)
  },

  setAddress: (address: string): void => {
    localStorage.setItem(ADDRESS_KEY, address)
  },

  removeAddress: (): void => {
    localStorage.removeItem(ADDRESS_KEY)
  },

  clear: (): void => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(ADDRESS_KEY)
  },
}
