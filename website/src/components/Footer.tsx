import { motion } from 'framer-motion'
import { Shield, Lock, EyeOff } from 'lucide-react'

export default function Footer() {
  return (
    <footer className="bg-gray-100 dark:bg-dark-900 py-12 border-t border-gray-300 dark:border-dark-600/50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center">
          {/* Logo and Brand */}
          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="flex items-center justify-center space-x-3 mb-6"
          >
            <div className="w-10 h-10 bg-gradient-to-br from-secure-500 to-primary-500 rounded-xl flex items-center justify-center">
              <span className="text-white font-bold text-lg">SM</span>
            </div>
            <span className="text-2xl font-bold gradient-text">SecureMail</span>
          </motion.div>

          {/* Tagline */}
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="text-lg text-gray-600 dark:text-gray-300 mb-8"
          >
            The World's Most Secure Email
          </motion.p>

          {/* Security Icons */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.4 }}
            className="flex justify-center items-center space-x-6 text-secure-400 mb-8"
          >
            <div className="flex items-center space-x-2">
              <Shield className="w-5 h-5" />
              <span className="text-sm font-medium">Zero-Knowledge</span>
            </div>
            <div className="flex items-center space-x-2">
              <Lock className="w-5 h-5" />
              <span className="text-sm font-medium">Military-Grade</span>
            </div>
            <div className="flex items-center space-x-2">
              <EyeOff className="w-5 h-5" />
              <span className="text-sm font-medium">No Visibility</span>
            </div>
          </motion.div>

          {/* Copyright */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.6 }}
            className="text-gray-500 dark:text-gray-400 text-sm"
          >
            © 2024 SecureMail. All rights reserved. Privacy by design.
          </motion.div>
        </div>
      </div>
    </footer>
  )
}
