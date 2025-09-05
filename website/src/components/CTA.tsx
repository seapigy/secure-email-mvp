import { MotionDiv, useInView } from '../utils/animations'
import { useRef } from 'react'
import { Shield, ArrowRight, Lock, EyeOff } from '../utils/icons'

export default function CTA() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })

  return (
    <section ref={ref} className="section-padding bg-gradient-to-b from-gray-100 to-gray-200 dark:from-dark-800 dark:to-dark-900 relative overflow-hidden">
      {/* Background Elements */}
      <div className="absolute inset-0">
        <MotionDiv
          animate={{ 
            rotate: [0, 360],
            scale: [1, 1.1, 1]
          }}
          transition={{ 
            duration: 20,
            repeat: Infinity,
            ease: "linear"
          }}
          className="absolute top-20 left-20 w-32 h-32 bg-gradient-to-br from-indigo-500/20 to-blue-500/20 rounded-full blur-xl"
        />
        <MotionDiv
          animate={{ 
            rotate: [360, 0],
            scale: [1.1, 1, 1.1]
          }}
          transition={{ 
            duration: 25,
            repeat: Infinity,
            ease: "linear"
          }}
          className="absolute bottom-20 right-20 w-40 h-40 bg-gradient-to-br from-blue-500/20 to-indigo-500/20 rounded-full blur-xl"
        />
      </div>

      <div className="max-w-7xl mx-auto relative z-10">
        <div className="text-center">
          {/* Main CTA */}
          <MotionDiv
            initial={{ opacity: 0, y: 30 }}
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            transition={{ duration: 0.8 }}
            className="mb-12"
          >
            <h2 className="text-4xl md:text-6xl font-bold mb-6">
              <span className="gradient-text">Ready to Experience</span>
              <br />
              <span className="text-dark-900 dark:text-white">True Email Security?</span>
            </h2>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto mb-8">
              Join the revolution in email security. Experience zero-knowledge architecture, 
              quantum-resistant encryption, and privacy that actually works.
            </p>

            <MotionDiv
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="inline-flex items-center space-x-3 btn-primary text-lg px-8 py-4"
              as="button"
            >
              <span>Get Started Now</span>
              <ArrowRight className="w-5 h-5" />
            </MotionDiv>
          </MotionDiv>

          {/* Statistics */}
          <MotionDiv
            initial={{ opacity: 0, y: 30 }}
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12"
          >
            <div className="text-center">
              <div className="text-3xl md:text-4xl font-bold text-indigo-500 mb-2">256-bit</div>
              <div className="text-gray-600 dark:text-gray-300">AES Encryption</div>
            </div>
            <div className="text-center">
              <div className="text-3xl md:text-4xl font-bold text-blue-500 mb-2">Zero</div>
              <div className="text-gray-600 dark:text-gray-300">Knowledge Architecture</div>
            </div>
            <div className="text-center">
              <div className="text-3xl md:text-4xl font-bold text-indigo-500 mb-2">Quantum</div>
              <div className="text-gray-600 dark:text-gray-300">Resistant Security</div>
            </div>
          </MotionDiv>

          {/* Security Features */}
          <MotionDiv
            initial={{ opacity: 0, y: 30 }}
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            transition={{ duration: 0.8, delay: 0.4 }}
            className="flex flex-wrap items-center justify-center gap-8 text-sm text-gray-600 dark:text-gray-300"
          >
            <div className="flex items-center space-x-2">
              <Shield className="w-4 h-4 text-indigo-500" />
              <span>Military-Grade Security</span>
            </div>
            <div className="flex items-center space-x-2">
              <Lock className="w-4 h-4 text-blue-500" />
              <span>End-to-End Encryption</span>
            </div>
            <div className="flex items-center space-x-2">
              <EyeOff className="w-4 h-4 text-indigo-500" />
              <span>Complete Privacy</span>
            </div>
          </MotionDiv>
        </div>
      </div>
    </section>
  )
}