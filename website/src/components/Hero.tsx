import { motion } from 'framer-motion'
import { Shield, Lock, EyeOff, ArrowRight } from 'lucide-react'

export default function Hero() {
  return (
    <section className="relative min-h-screen flex items-center justify-center overflow-hidden pt-16">
      {/* Static Background */}
      <div className="absolute inset-0 bg-gradient-to-br from-white via-gray-50 to-white dark:from-dark-900 dark:via-dark-800 dark:to-dark-900">
      </div>

      {/* Content */}
      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <motion.div
          initial={{ opacity: 0, y: 50 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1.2, ease: "easeOut" }}
          className="space-y-8"
        >
          {/* Main Headline */}
          <motion.h1 
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="text-5xl md:text-7xl lg:text-8xl font-black leading-tight"
          >
            <span className="gradient-text">The World's</span>
            <br />
            <span className="text-dark-900 dark:text-white">Most Secure</span>
            <br />
            <span className="gradient-text">Email.</span>
          </motion.h1>

          {/* Subheadline */}
          <motion.p
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.4 }}
            className="text-xl md:text-2xl lg:text-3xl text-gray-600 dark:text-gray-300 max-w-5xl mx-auto font-light"
          >
            <span className="font-semibold text-secure-400">The Most Secure Email in the World</span> — powered by{' '}
            <span className="font-semibold text-primary-400">AES-256-GCM</span>,{' '}
            <span className="font-semibold text-secure-400">Argon2id</span>,{' '}
            <span className="font-semibold text-primary-400">TLS 1.3</span>, and{' '}
            <span className="font-semibold text-secure-400">PQC hybrid encryption</span>.{' '}
            <span className="font-semibold text-primary-400">Zero-knowledge.</span>{' '}
            <span className="font-semibold text-secure-400">Zero visibility.</span>{' '}
            <span className="font-semibold text-primary-400">Absolute privacy.</span>
          </motion.p>

          {/* CTA Buttons */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.6 }}
            className="flex flex-col sm:flex-row gap-4 justify-center items-center"
          >
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-10 py-4 flex items-center space-x-2"
            >
              <span>Get Early Access</span>
              <ArrowRight className="w-5 h-5" />
            </motion.button>
            
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-secondary text-lg px-10 py-4"
            >
              Learn More
            </motion.button>
          </motion.div>

          {/* Trust Indicators */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.8 }}
            className="flex flex-wrap justify-center items-center gap-8 pt-8 text-gray-500 dark:text-gray-400"
          >
            <div className="flex items-center space-x-2">
              <Shield className="w-5 h-5 text-secure-400" />
              <span className="text-sm">AES-256-GCM</span>
            </div>
            <div className="flex items-center space-x-2">
              <Lock className="w-5 h-5 text-primary-400" />
              <span className="text-sm">Zero-Knowledge</span>
            </div>
            <div className="flex items-center space-x-2">
              <EyeOff className="w-5 h-5 text-secure-400" />
              <span className="text-sm">No Visibility</span>
            </div>
          </motion.div>
        </motion.div>
      </div>

      {/* Scroll Indicator */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 1, delay: 1.2 }}
        className="absolute bottom-8 left-1/2 transform -translate-x-1/2"
      >
        <motion.div
          animate={{ y: [0, 10, 0] }}
          transition={{ duration: 2, repeat: Infinity }}
          className="w-6 h-10 border-2 border-gray-400 rounded-full flex justify-center"
        >
          <motion.div
            animate={{ y: [0, 12, 0] }}
            transition={{ duration: 2, repeat: Infinity }}
            className="w-1 h-3 bg-gray-400 rounded-full mt-2"
          />
        </motion.div>
      </motion.div>
    </section>
  )
}
