import { motion } from 'framer-motion'
import { Sun, Moon, Menu, X } from 'lucide-react'
import { useState } from 'react'
import { AnimatePresence } from 'framer-motion'

interface HeaderProps {
  isDark: boolean
  setIsDark: (isDark: boolean) => void
}

export default function Header({ isDark, setIsDark }: HeaderProps) {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  return (
    <motion.header 
      initial={{ y: -100, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.6 }}
      className="fixed top-0 left-0 right-0 z-50 glass-effect shadow-lg"
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <motion.div 
            whileHover={{ scale: 1.05 }}
            className="flex items-center space-x-2"
          >
            <div className="w-8 h-8 bg-gradient-to-br from-secure-500 to-primary-500 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-sm">SM</span>
            </div>
            <span className="text-xl font-bold gradient-text">SecureMail</span>
          </motion.div>

          {/* Desktop Navigation */}
                     <nav className="hidden md:flex items-center space-x-8">
             <a href="#features" className="text-dark-900 dark:text-white hover:text-secure-400 transition-colors">Features</a>
             <a href="#security" className="text-dark-900 dark:text-white hover:text-secure-400 transition-colors">Security</a>
             <a href="#comparison" className="text-dark-900 dark:text-white hover:text-secure-400 transition-colors">Comparison</a>
             <a href="#trust" className="text-dark-900 dark:text-white hover:text-secure-400 transition-colors">Trust</a>
           </nav>

          {/* Right side */}
          <div className="flex items-center space-x-4">
            {/* Theme Toggle */}
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={() => setIsDark(!isDark)}
              className="p-2 rounded-lg glass-effect hover:bg-secure-500/10 transition-colors"
            >
              {isDark ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
            </motion.button>

            {/* CTA Button */}
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary hidden sm:block"
            >
              Get Started
            </motion.button>

            {/* Mobile Menu Button */}
            <button
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className="md:hidden p-2 rounded-lg glass-effect"
            >
              {isMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {/* Mobile Menu */}
        <AnimatePresence>
          {isMenuOpen && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              transition={{ duration: 0.3 }}
              className="md:hidden py-4 border-t border-white/20 dark:border-dark-600/50"
            >
                             <nav className="flex flex-col space-y-4">
                 <a href="#features" className="text-dark-900 dark:text-white hover:text-secure-400 transition-colors">Features</a>
                 <a href="#security" className="text-dark-900 dark:text-white hover:text-secure-400 transition-colors">Security</a>
                 <a href="#comparison" className="text-dark-900 dark:text-white hover:text-secure-400 transition-colors">Comparison</a>
                 <a href="#trust" className="text-dark-900 dark:text-white hover:text-secure-400 transition-colors">Trust</a>
                 <button className="btn-primary w-full mt-4">Get Started</button>
               </nav>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </motion.header>
  )
}
