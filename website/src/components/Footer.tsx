import { MotionDiv } from '../utils/animations'
import { Shield, Lock, EyeOff } from '../utils/icons'

export default function Footer() {
  return (
    <footer className="bg-gray-100 dark:bg-dark-900 py-12 border-t border-gray-300 dark:border-dark-600/50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center">
          {/* Logo and Brand */}
          <MotionDiv 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="flex items-center justify-center space-x-3 mb-6"
          >
            <div className="w-10 h-10 bg-gradient-to-br from-indigo-500 to-blue-500 rounded-xl flex items-center justify-center">
              <span className="text-white font-bold text-lg">SM</span>
            </div>
            <span className="text-2xl font-bold gradient-text">SecureMail</span>
          </MotionDiv>

          {/* Tagline */}
          <MotionDiv
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="text-lg text-gray-600 dark:text-gray-300 mb-8"
            as="p"
          >
            The World's Most Secure Email
          </MotionDiv>

          {/* Security Icons */}
          <MotionDiv
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.4 }}
            className="flex items-center justify-center space-x-6 mb-8"
          >
            <div className="flex items-center space-x-2 text-indigo-500">
              <Shield className="w-5 h-5" />
              <span className="text-sm font-medium">Zero-Knowledge</span>
            </div>
            <div className="flex items-center space-x-2 text-indigo-500">
              <Lock className="w-5 h-5" />
              <span className="text-sm font-medium">End-to-End</span>
            </div>
            <div className="flex items-center space-x-2 text-indigo-500">
              <EyeOff className="w-5 h-5" />
              <span className="text-sm font-medium">Private</span>
            </div>
          </MotionDiv>

          {/* Copyright */}
          <MotionDiv
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.6, delay: 0.6 }}
            className="text-sm text-gray-500 dark:text-gray-400"
            as="p"
          >
            © 2024 SecureMail. All rights reserved. Privacy by design.
          </MotionDiv>
        </div>
      </div>
    </footer>
  )
}