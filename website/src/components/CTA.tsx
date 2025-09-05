import { motion } from 'framer-motion'
import { useInView } from 'framer-motion'
import { useRef } from 'react'
import { Shield, ArrowRight, Lock, EyeOff } from 'lucide-react'

export default function CTA() {
  const ref = useRef(null)
  const isInView = useInView(ref, { once: true, margin: "-100px" })

  return (
    <section ref={ref} className="section-padding bg-gradient-to-b from-gray-100 to-gray-200 dark:from-dark-800 dark:to-dark-900 relative overflow-hidden">
      {/* Background Elements */}
      <div className="absolute inset-0">
        <motion.div
          animate={{ 
            rotate: [0, 360],
            scale: [1, 1.1, 1]
          }}
          transition={{ 
            duration: 20,
            repeat: Infinity,
            ease: "linear"
          }}
          className="absolute top-20 right-20 w-64 h-64 bg-gradient-to-br from-secure-500/10 to-primary-500/10 rounded-full blur-3xl"
        />
        <motion.div
          animate={{ 
            rotate: [360, 0],
            scale: [1, 1.2, 1]
          }}
          transition={{ 
            duration: 25,
            repeat: Infinity,
            ease: "linear"
          }}
          className="absolute bottom-20 left-20 w-80 h-80 bg-gradient-to-br from-primary-500/10 to-secure-500/10 rounded-full blur-3xl"
        />
      </div>

      <div className="relative z-10 max-w-7xl mx-auto text-center">
        {/* Main CTA Content */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="mb-16"
        >
          {/* Bold Headline Restatement */}
                     <motion.h2
             initial={{ opacity: 0, y: 30 }}
             animate={isInView ? { opacity: 1, y: 0 } : {}}
             transition={{ duration: 0.8, delay: 0.2 }}
             className="text-5xl md:text-7xl lg:text-8xl font-black leading-tight mb-8"
           >
             <span className="gradient-text">Email Reinvented.</span>
             <br />
             <span className="text-dark-900 dark:text-white">Security Without</span>
             <br />
             <span className="gradient-text">Compromise.</span>
           </motion.h2>

          {/* Supporting Text */}
          <motion.p
            initial={{ opacity: 0, y: 30 }}
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            transition={{ duration: 0.8, delay: 0.4 }}
            className="text-xl md:text-2xl text-gray-600 dark:text-gray-300 max-w-4xl mx-auto mb-12"
          >
            Join thousands of security-conscious individuals and organizations who have 
            already made the switch to the world's most secure email platform.
          </motion.p>

          {/* Security Icons Row */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            transition={{ duration: 0.8, delay: 0.6 }}
            className="flex flex-wrap justify-center items-center gap-8 mb-12 text-secure-400"
          >
            <motion.div
              animate={{ float: true }}
              className="flex items-center space-x-2"
            >
              <Shield className="w-6 h-6" />
              <span className="text-sm font-medium">Zero-Knowledge</span>
            </motion.div>
            <motion.div
              animate={{ float: true }}
              transition={{ delay: 0.5 }}
              className="flex items-center space-x-2"
            >
              <Lock className="w-6 h-6" />
              <span className="text-sm font-medium">Military-Grade</span>
            </motion.div>
            <motion.div
              animate={{ float: true }}
              transition={{ delay: 1 }}
              className="flex items-center space-x-2"
            >
              <EyeOff className="w-6 h-6" />
              <span className="text-sm font-medium">No Visibility</span>
            </motion.div>
          </motion.div>
        </motion.div>

        {/* CTA Buttons */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="flex flex-col sm:flex-row gap-6 justify-center items-center mb-16"
        >
                     <motion.button
             whileHover={{ scale: 1.05 }}
             whileTap={{ scale: 0.95 }}
             className="btn-primary text-xl px-12 py-4 flex items-center space-x-3"
           >
             <span>Join the Waitlist</span>
             <ArrowRight className="w-6 h-6" />
           </motion.button>
          
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="btn-secondary text-xl px-12 py-4"
          >
            Schedule Demo
          </motion.button>
        </motion.div>

        {/* Trust Indicators */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 1.0 }}
          className="glass-effect rounded-2xl p-8 max-w-4xl mx-auto"
        >
                     <h3 className="text-2xl font-bold mb-6 text-dark-900 dark:text-white">
             Why Security Experts Choose SecureMail
           </h3>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-sm">
            <div className="text-center">
              <div className="text-3xl font-bold text-secure-400 mb-2">99.99%</div>
              <div className="text-gray-600 dark:text-gray-300">Uptime Guarantee</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-primary-400 mb-2">256-bit</div>
              <div className="text-gray-600 dark:text-gray-300">Encryption Standard</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-secure-400 mb-2">24/7</div>
              <div className="text-gray-600 dark:text-gray-300">Security Monitoring</div>
            </div>
          </div>
        </motion.div>

        {/* Final Statement */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 1.2 }}
          className="mt-16"
        >
                     <p className="text-lg text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
             <span className="font-semibold text-secure-400">No compromises.</span>{' '}
             <span className="font-semibold text-primary-400">No exceptions.</span>{' '}
             <span className="font-semibold text-secure-400">Just security.</span>
           </p>
        </motion.div>
      </div>
    </section>
  )
}
