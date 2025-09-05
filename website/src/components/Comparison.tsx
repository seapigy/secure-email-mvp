import { motion } from 'framer-motion'
import { useInView } from 'framer-motion'
import { useRef } from 'react'
import { Check, X, Shield, Lock, EyeOff, Zap, Globe, Clock } from 'lucide-react'

export default function Comparison() {
  const ref = useRef(null)
  const isInView = useInView(ref, { once: true, margin: "-100px" })

  const comparisonData = [
    {
      feature: "Basic TLS Encryption",
      other: true,
      securemail: true,
      icon: Lock
    },
    {
      feature: "Advanced Security Controls",
      other: false,
      securemail: true,
      icon: Shield
    },
    {
      feature: "Provider Visibility",
      other: true,
      securemail: false,
      icon: EyeOff
    },
    {
      feature: "Zero-Knowledge Architecture",
      other: false,
      securemail: true,
      icon: Lock
    },
    {
      feature: "PQC Ready",
      other: false,
      securemail: true,
      icon: Zap
    },
    {
      feature: "Self-Destruct Messages",
      other: false,
      securemail: true,
      icon: Clock
    },
    {
      feature: "Enterprise Compliance",
      other: false,
      securemail: true,
      icon: Globe
    },
    {
      feature: "Geolocation Locks",
      other: false,
      securemail: true,
      icon: Globe
    }
  ]

  return (
    <section id="comparison" ref={ref} className="section-padding bg-gradient-to-b from-gray-100 to-gray-200 dark:from-dark-800 dark:to-dark-900">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-20"
        >
                     <h2 className="text-4xl md:text-6xl font-bold mb-6">
             <span className="gradient-text">Unlike Anything</span>
             <br />
             <span className="text-dark-900 dark:text-white">in the World</span>
           </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            See how SecureMail's revolutionary approach to email security 
            leaves traditional providers in the dust.
          </p>
        </motion.div>

        {/* Comparison Table */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="mb-20"
        >
          <div className="glass-effect rounded-3xl p-8 overflow-hidden">
            {/* Table Header */}
                         <div className="grid grid-cols-3 gap-6 mb-8 pb-6 border-b border-gray-300 dark:border-dark-600/50">
                             <div className="text-left">
                 <h3 className="text-xl font-semibold text-dark-900 dark:text-white mb-2">Security Features</h3>
                 <p className="text-gray-600 dark:text-gray-400 text-sm">What matters for your privacy</p>
               </div>
               <div className="text-center">
                 <h3 className="text-xl font-semibold text-gray-600 dark:text-gray-400 mb-2">Other Email Providers</h3>
                 <p className="text-gray-600 dark:text-gray-500 text-sm">Basic protection</p>
               </div>
              <div className="text-center">
                <h3 className="text-xl font-semibold gradient-text mb-2">SecureMail</h3>
                <p className="text-secure-400 text-sm">Revolutionary security</p>
              </div>
            </div>

            {/* Comparison Rows */}
            <div className="space-y-4">
              {comparisonData.map((row, index) => (
                <motion.div
                  key={row.feature}
                  initial={{ opacity: 0, x: -30 }}
                  animate={isInView ? { opacity: 1, x: 0 } : {}}
                  transition={{ duration: 0.6, delay: 0.4 + index * 0.1 }}
                  className="grid grid-cols-3 gap-6 items-center py-4 hover:bg-gray-100 dark:hover:bg-dark-700/50 rounded-xl px-4 transition-colors"
                >
                                     <div className="flex items-center space-x-3">
                     <row.icon className="w-5 h-5 text-secure-400" />
                     <span className="text-dark-900 dark:text-white font-medium">{row.feature}</span>
                   </div>
                  
                  <div className="text-center">
                    {row.other ? (
                      <motion.div
                        initial={{ scale: 0 }}
                        animate={isInView ? { scale: 1 } : {}}
                        transition={{ duration: 0.3, delay: 0.6 + index * 0.1 }}
                        className="inline-flex items-center justify-center w-8 h-8 bg-green-500/20 rounded-full"
                      >
                        <Check className="w-5 h-5 text-green-400" />
                      </motion.div>
                    ) : (
                      <motion.div
                        initial={{ scale: 0 }}
                        animate={isInView ? { scale: 1 } : {}}
                        transition={{ duration: 0.3, delay: 0.6 + index * 0.1 }}
                        className="inline-flex items-center justify-center w-8 h-8 bg-red-500/20 rounded-full"
                      >
                        <X className="w-5 h-5 text-red-400" />
                      </motion.div>
                    )}
                  </div>
                  
                  <div className="text-center">
                    {row.securemail ? (
                      <motion.div
                        initial={{ scale: 0 }}
                        animate={isInView ? { scale: 1 } : {}}
                        transition={{ duration: 0.3, delay: 0.6 + index * 0.1 }}
                        className="inline-flex items-center justify-center w-8 h-8 bg-secure-500/20 rounded-full"
                      >
                        <Check className="w-5 h-5 text-secure-400" />
                      </motion.div>
                    ) : (
                      <motion.div
                        initial={{ scale: 0 }}
                        animate={isInView ? { scale: 1 } : {}}
                        transition={{ duration: 0.3, delay: 0.6 + index * 0.1 }}
                        className="inline-flex items-center justify-center w-8 h-8 bg-red-500/20 rounded-full"
                      >
                        <X className="w-5 h-5 text-red-400" />
                      </motion.div>
                    )}
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </motion.div>

        {/* Bottom CTA */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="text-center"
        >
          <div className="glass-effect rounded-2xl p-8 max-w-2xl mx-auto">
                         <h3 className="text-2xl md:text-3xl font-bold mb-4 text-dark-900 dark:text-white">
               The Choice is Clear
             </h3>
            <p className="text-lg text-gray-600 dark:text-gray-300 mb-6">
              While others offer basic protection, SecureMail delivers 
              military-grade security with zero compromises.
            </p>
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-8 py-3"
            >
              Start Securing Your Emails
            </motion.button>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
