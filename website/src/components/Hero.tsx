import { MotionDiv, MotionButton, fadeIn, slideUp, slideInLeft, slideInRight } from '../utils/animations'
import { Shield, Lock, EyeOff, ArrowRight } from '../utils/icons'

export default function Hero() {
  return (
    <section className="relative min-h-screen flex items-center justify-center overflow-hidden pt-16">
      {/* Static Background */}
      <div className="absolute inset-0 bg-gradient-to-br from-white via-gray-50 to-white dark:from-dark-900 dark:via-dark-800 dark:to-dark-900">
      </div>

      {/* Content */}
      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <MotionDiv
          initial={{ opacity: 0, transform: 'translateY(50px)' }}
          animate={{ opacity: 1, transform: 'translateY(0)' }}
          transition={{ duration: 1.2 }}
          className="space-y-8"
        >
          {/* Main Headline */}
          <h1 
            className="text-5xl md:text-7xl lg:text-8xl font-black leading-tight animate-slide-up"
            style={{ animationDelay: '0.2s' }}
          >
            <span className="gradient-text">The World's</span>
            <br />
            <span className="text-dark-900 dark:text-white">Most Secure</span>
            <br />
            <span className="gradient-text">Email.</span>
          </h1>

          {/* Subheadline */}
          <p
            className="text-xl md:text-2xl lg:text-3xl text-gray-600 dark:text-gray-300 max-w-5xl mx-auto font-light animate-slide-up"
            style={{ animationDelay: '0.4s' }}
          >
            <span className="font-semibold text-secure-400">The Most Secure Email in the World</span> — powered by{' '}
            <span className="font-semibold text-primary-400">AES-256-GCM</span>,{' '}
            <span className="font-semibold text-secure-400">Argon2id</span>,{' '}
            <span className="font-semibold text-primary-400">TLS 1.3</span>, and{' '}
            <span className="font-semibold text-secure-400">PQC hybrid encryption</span>.{' '}
            <span className="font-semibold text-primary-400">Zero-knowledge.</span>{' '}
            <span className="font-semibold text-secure-400">Zero visibility.</span>{' '}
            <span className="font-semibold text-primary-400">Absolute privacy.</span>
          </p>

          {/* CTA Buttons */}
          <div
            className="flex flex-col sm:flex-row gap-4 justify-center items-center animate-slide-up"
            style={{ animationDelay: '0.6s' }}
          >
            <MotionButton
              whileHover={{ transform: 'scale(1.05)' }}
              whileTap={{ transform: 'scale(0.95)' }}
              className="btn-primary text-lg px-10 py-4 flex items-center space-x-2"
            >
              <span>Get Early Access</span>
              <ArrowRight className="w-5 h-5" />
            </MotionButton>
            
            <MotionButton
              whileHover={{ transform: 'scale(1.05)' }}
              whileTap={{ transform: 'scale(0.95)' }}
              className="btn-secondary text-lg px-10 py-4"
            >
              Learn More
            </MotionButton>
          </div>

          {/* Trust Indicators */}
          <div
            className="flex flex-wrap justify-center items-center gap-8 pt-8 text-gray-500 dark:text-gray-400 animate-slide-up"
            style={{ animationDelay: '0.8s' }}
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
          </div>
        </MotionDiv>
      </div>

      {/* Scroll Indicator */}
      <div
        className="absolute bottom-8 left-1/2 transform -translate-x-1/2 animate-fade-in"
        style={{ animationDelay: '1.2s' }}
      >
        <div
          className="w-6 h-10 border-2 border-gray-400 rounded-full flex justify-center animate-float"
        >
          <div
            className="w-1 h-3 bg-gray-400 rounded-full mt-2 animate-float"
            style={{ animationDelay: '0.5s' }}
          />
        </div>
      </div>
    </section>
  )
}
