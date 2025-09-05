import { motion } from 'framer-motion'
import { useInView } from 'framer-motion'
import { useRef } from 'react'
import { Shield, Lock, EyeOff, Server, Key, Globe } from 'lucide-react'

export default function TrustBreaker() {
  const ref = useRef(null)
  const isInView = useInView(ref, { once: true, margin: "-100px" })

  const features = [
    {
      icon: Shield,
      title: "AES-256-GCM Encryption",
      description: "Military-grade encryption that's virtually unbreakable",
      color: "text-secure-400"
    },
    {
      icon: Lock,
      title: "Argon2id Hashing",
      description: "State-of-the-art password hashing with memory-hard functions",
      color: "text-primary-400"
    },
    {
      icon: Key,
      title: "TLS 1.3 + PQC Ready",
      description: "Quantum-resistant encryption for future-proof security",
      color: "text-secure-400"
    },
    {
      icon: Server,
      title: "Zero-Knowledge Architecture",
      description: "We cannot see your emails, not even metadata",
      color: "text-primary-400"
    },
    {
      icon: EyeOff,
      title: "No Visibility Model",
      description: "Complete privacy with no server-side access",
      color: "text-secure-400"
    },
    {
      icon: Globe,
      title: "Global Compliance",
      description: "GDPR, HIPAA, SOC2 ready for enterprise use",
      color: "text-primary-400"
    }
  ]

  return (
    <section id="security" ref={ref} className="section-padding bg-gradient-to-b from-gray-100 to-gray-200 dark:from-dark-800 dark:to-dark-900">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-20"
        >
                       <h2 className="text-4xl md:text-6xl font-bold mb-6">
               <span className="gradient-text">Why SecureMail</span>
               <br />
               <span className="text-dark-900 dark:text-white">is Different</span>
             </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            While other providers claim security, we deliver true zero-knowledge privacy. 
            Our servers are designed to be completely blind to your data.
          </p>
        </motion.div>

        {/* Zero Knowledge Explanation */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="mb-20"
        >
          <div className="glass-effect rounded-3xl p-8 md:p-12 text-center">
            <motion.div
              initial={{ scale: 0 }}
              animate={isInView ? { scale: 1 } : {}}
              transition={{ duration: 0.6, delay: 0.4 }}
              className="w-24 h-24 bg-gradient-to-br from-secure-500 to-primary-500 rounded-full flex items-center justify-center mx-auto mb-6"
            >
              <EyeOff className="w-12 h-12 text-white" />
            </motion.div>
            
                         <h3 className="text-3xl md:text-4xl font-bold mb-4 text-dark-900 dark:text-white">
               We Cannot See Your Emails
             </h3>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-6 max-w-2xl mx-auto">
              <span className="font-semibold text-secure-400">Not content.</span>{' '}
              <span className="font-semibold text-primary-400">Not sender.</span>{' '}
              <span className="font-semibold text-secure-400">Not recipient.</span>{' '}
              <span className="font-semibold text-primary-400">Nothing.</span>
            </p>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-gray-500 dark:text-gray-400">
              <div className="flex items-center justify-center space-x-2">
                <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                <span>No Email Content</span>
              </div>
              <div className="flex items-center justify-center space-x-2">
                <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                <span>No Metadata</span>
              </div>
              <div className="flex items-center justify-center space-x-2">
                <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                <span>No Server Access</span>
              </div>
            </div>
          </div>
        </motion.div>

        {/* Technical Features Grid */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.4 }}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, index) => (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 30 }}
                animate={isInView ? { opacity: 1, y: 0 } : {}}
                transition={{ duration: 0.6, delay: 0.6 + index * 0.1 }}
                className="feature-card group"
              >
                <div className={`w-16 h-16 bg-gradient-to-br from-secure-500/20 to-primary-500/20 rounded-2xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform duration-300`}>
                  <feature.icon className={`w-8 h-8 ${feature.color}`} />
                </div>
                
                                 <h4 className="text-xl font-semibold mb-2 text-dark-900 dark:text-white">
                   {feature.title}
                 </h4>
                 
                 <p className="text-gray-700 dark:text-gray-300">
                   {feature.description}
                 </p>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Trust Statement */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="text-center mt-20"
        >
          <div className="glass-effect rounded-2xl p-8 inline-block">
                         <p className="text-2xl md:text-3xl font-bold text-dark-900 dark:text-white mb-4">
               "If it's not SecureMail, it's not secure."
             </p>
             <p className="text-lg text-gray-700 dark:text-gray-300">
               Join the revolution in email privacy
             </p>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
