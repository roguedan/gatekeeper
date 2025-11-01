# Gatekeeper Frontend

React + TypeScript frontend for Gatekeeper wallet-native authentication gateway.

## Features

- **SIWE Authentication**: Sign-In with Ethereum (EIP-4361) wallet authentication
- **Wallet Integration**: Support for MetaMask, WalletConnect, Coinbase Wallet, and more via RainbowKit
- **JWT Token Management**: Secure token storage and automatic refresh
- **Protected Routes**: Route-based authentication with AuthGuard
- **API Key Management**: Create, view, and revoke API keys
- **Token Gating**: Demo of blockchain-based access control
- **Responsive Design**: Mobile-first design with Tailwind CSS
- **Dark Mode**: Full dark mode support
- **TypeScript**: Type-safe development experience

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Styling
- **wagmi** - React hooks for Ethereum
- **RainbowKit** - Wallet connection UI
- **React Query** - Data fetching and caching
- **React Router** - Client-side routing
- **Axios** - HTTP client
- **SIWE** - Sign-In with Ethereum library

## Quick Start

### Prerequisites

- Node.js 18+ and npm
- MetaMask or another Web3 wallet
- Backend API running on `http://localhost:8080` (see main README)

### Installation

```bash
cd web
npm install
```

### Environment Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Edit `.env`:

```env
VITE_API_URL=http://localhost:8080
VITE_CHAIN_ID=1
VITE_ETHEREUM_RPC_URL=https://eth-mainnet.g.alchemy.com/v2/your-api-key
VITE_WALLETCONNECT_PROJECT_ID=your-walletconnect-project-id
VITE_APP_NAME=Gatekeeper
VITE_APP_DOMAIN=localhost:3000
```

### Development

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Build for Production

```bash
npm run build
npm run preview
```

## Project Structure

```
web/
├── src/
│   ├── components/          # React components
│   │   ├── auth/           # Authentication components
│   │   ├── common/         # Reusable UI components
│   │   └── layout/         # Layout components
│   ├── config/             # Configuration files
│   ├── contexts/           # React contexts
│   ├── hooks/              # Custom React hooks
│   ├── pages/              # Page components
│   ├── services/           # API and service layer
│   ├── types/              # TypeScript type definitions
│   ├── test/               # Test utilities
│   ├── App.tsx             # Main App component
│   ├── main.tsx            # Entry point
│   └── index.css           # Global styles
├── public/                 # Static assets
├── index.html              # HTML template
├── package.json            # Dependencies
├── tsconfig.json           # TypeScript config
├── vite.config.ts          # Vite config
├── tailwind.config.js      # Tailwind config
└── vitest.config.ts        # Test config
```

## Testing

### Run Tests

```bash
npm test
```

### Run Tests with UI

```bash
npm run test:ui
```

### Generate Coverage Report

```bash
npm run test:coverage
```

Coverage threshold is set to 80% for lines, functions, branches, and statements.

## Key Components

### Authentication Flow

1. **WalletConnect**: User connects wallet via RainbowKit
2. **SignInFlow**: User signs SIWE message
3. **Token Storage**: JWT token stored in localStorage
4. **AuthGuard**: Protects routes requiring authentication

### Custom Hooks

- `useAuth`: Authentication state management
- `useSIWE`: Sign-In with Ethereum flow
- `useAPIKeys`: API key CRUD operations
- `useProtectedData`: Fetch protected resources

### Services

- `apiClient`: Axios HTTP client with JWT interceptors
- `authService`: SIWE authentication
- `apiKeyService`: API key management
- `protectedService`: Protected resource access
- `storage`: LocalStorage utilities

## API Integration

The frontend integrates with the Gatekeeper backend API:

### Authentication Endpoints

- `GET /auth/siwe/nonce` - Get nonce for SIWE
- `POST /auth/siwe/verify` - Verify signature and get JWT

### Protected Endpoints

- `GET /api/data` - Example protected endpoint
- `POST /api/keys` - Create API key
- `GET /api/keys` - List API keys
- `DELETE /api/keys/:id` - Revoke API key

All protected endpoints require `Authorization: Bearer <token>` header.

## Troubleshooting

### Wallet Connection Issues

- Ensure MetaMask or another wallet is installed
- Check that you're on a supported network
- Clear browser cache and localStorage

### API Connection Issues

- Verify backend is running on correct port
- Check CORS configuration in backend
- Verify API_URL in `.env` file

### Build Issues

- Clear `node_modules` and reinstall: `rm -rf node_modules package-lock.json && npm install`
- Clear Vite cache: `rm -rf node_modules/.vite`

## Performance

- Code splitting by route and vendor chunks
- Optimized bundle size with tree-shaking
- Lazy loading for non-critical components
- Target production build: <500KB gzipped

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

## Contributing

1. Create a feature branch
2. Make your changes
3. Add tests for new functionality
4. Ensure all tests pass: `npm test`
5. Run type checking: `npm run type-check`
6. Submit a pull request

## License

MIT License - see LICENSE file for details
