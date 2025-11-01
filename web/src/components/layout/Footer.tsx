import { Github, Twitter, Globe } from 'lucide-react'

export const Footer = () => {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 mt-auto">
      <div className="max-w-7xl mx-auto px-3 sm:px-6 lg:px-8 py-6 sm:py-8">
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6 sm:gap-8">
          {/* About */}
          <div>
            <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-2 sm:mb-3">Gatekeeper</h3>
            <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
              Wallet-native authentication gateway using Sign-In with Ethereum (SIWE) and blockchain-based access control.
            </p>
          </div>

          {/* Links */}
          <div>
            <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-2 sm:mb-3">Resources</h3>
            <ul className="space-y-1.5 sm:space-y-2">
              <li>
                <a href="#" className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 transition-colors">
                  Documentation
                </a>
              </li>
              <li>
                <a href="#" className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 transition-colors">
                  API Reference
                </a>
              </li>
              <li>
                <a href="#" className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 transition-colors">
                  GitHub
                </a>
              </li>
            </ul>
          </div>

          {/* Social */}
          <div className="sm:col-span-2 md:col-span-1">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-2 sm:mb-3">Connect</h3>
            <div className="flex space-x-3 sm:space-x-4">
              <a href="#" className="text-gray-600 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 transition-colors p-1 min-h-[44px] min-w-[44px] flex items-center justify-center">
                <Github className="h-5 w-5" />
              </a>
              <a href="#" className="text-gray-600 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 transition-colors p-1 min-h-[44px] min-w-[44px] flex items-center justify-center">
                <Twitter className="h-5 w-5" />
              </a>
              <a href="#" className="text-gray-600 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 transition-colors p-1 min-h-[44px] min-w-[44px] flex items-center justify-center">
                <Globe className="h-5 w-5" />
              </a>
            </div>
          </div>
        </div>

        <div className="mt-6 sm:mt-8 pt-6 sm:pt-8 border-t border-gray-200 dark:border-gray-700">
          <p className="text-xs sm:text-sm text-center text-gray-600 dark:text-gray-400 leading-relaxed px-2">
            {currentYear} Gatekeeper. All rights reserved. Built with React, TypeScript, and wagmi.
          </p>
        </div>
      </div>
    </footer>
  )
}
