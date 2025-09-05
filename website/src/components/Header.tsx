import { Sun, Moon, Menu, X } from '../utils/icons'
import { useState } from 'react'
import { AnimatePresence, MotionDiv } from '../utils/animations'

interface HeaderProps {
  isDark: boolean
  setIsDark: (isDark: boolean) => void
}

export default function Header({ isDark, setIsDark }: HeaderProps) {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  return (
    <MotionDiv 
      initial={{ y: -100, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.6 }}
      className="fixed top-0 left-0 right-0 z-50 glass-effect shadow-lg"
      as="header"
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <MotionDiv 
            whileHover={{ scale: 1.05 }}
            className="flex items-center space-x-2"
          >
            <div className="w-8 h-8 bg-gradient-to-br from-indigo-500 to-blue-500 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-sm">SM</span>
            </div>
            <span className="text-xl font-bold gradient-text">SecureMail</span>
          </MotionDiv>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center space-x-8">
            <a href="#features" className="text-dark-900 dark:text-white hover:text-indigo-500 transition-colors">Features</a>
            <a href="#security" className="text-dark-900 dark:text-white hover:text-indigo-500 transition-colors">Security</a>
            <a href="#comparison" className="text-dark-900 dark:text-white hover:text-indigo-500 transition-colors">Comparison</a>
            <a href="#trust" className="text-dark-900 dark:text-white hover:text-indigo-500 transition-colors">Trust</a>
          </nav>

          {/* Right side */}
          <div className="flex items-center space-x-4">
            {/* Theme Toggle */}
            <MotionDiv
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={() => setIsDark(!isDark)}
              className="p-2 rounded-lg glass-effect hover:bg-indigo-500/10 transition-colors cursor-pointer"
              as="button"
            >
              {isDark ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
            </MotionDiv>

            {/* Mobile Menu Button */}
            <MotionDiv
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className="md:hidden p-2 rounded-lg glass-effect hover:bg-indigo-500/10 transition-colors cursor-pointer"
              as="button"
            >
              {isMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </MotionDiv>
          </div>
        </div>
      </div>

      {/* Mobile Menu */}
      <AnimatePresence>
        {isMenuOpen && (
          <MotionDiv
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.3 }}
            className="md:hidden glass-effect border-t border-gray-200 dark:border-gray-700"
          >
            <div className="px-4 py-4 space-y-2">
              <a 
                href="#features" 
                className="block py-2 text-dark-900 dark:text-white hover:text-indigo-500 transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                Features
              </a>
              <a 
                href="#security" 
                className="block py-2 text-dark-900 dark:text-white hover:text-indigo-500 transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                Security
              </a>
              <a 
                href="#comparison" 
                className="block py-2 text-dark-900 dark:text-white hover:text-indigo-500 transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                Comparison
              </a>
              <a 
                href="#trust" 
                className="block py-2 text-dark-900 dark:text-white hover:text-indigo-500 transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                Trust
              </a>
            </div>
          </MotionDiv>
        )}
      </AnimatePresence>
    </MotionDiv>
  )
}