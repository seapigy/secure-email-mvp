import { motion } from 'framer-motion'
import { useInView } from 'framer-motion'
import { useRef } from 'react'
import { 
  Globe, 
  Clock, 
  Eye, 
  Lock, 
  Timer, 
  MapPin, 
  Shield, 
  AlertTriangle, 
  Trash2, 
  Key,
  Zap,
  Fingerprint
} from 'lucide-react'

// DO NOT EDIT EXISTING CODE - This is a new component for Phase 2

export default function SecurityFeaturesGrid() {
  const ref = useRef(null)
  const isInView = useInView(ref, { once: true, margin: "-100px" })

  const securityFeatures = [
    {
      icon: Globe,
      title: "Geolocation Lock",
      description: "Restrict access to specific countries or regions. Perfect for compliance and security.",
      color: "text-secure-400",
      animation: "glow"
    },
    {
      icon: Clock,
      title: "Timed Destruction",
      description: "Emails automatically self-destruct after a set time. No traces left behind.",
      color: "text-primary-400",
      animation: "float"
    },
    {
      icon: Eye,
      title: "One-Time Read",
      description: "Emails disappear after being opened once. Perfect for sensitive information.",
      color: "text-secure-400",
      animation: "fade-in"
    },
    {
      icon: Lock,
      title: "Password Protection",
      description: "Add military-grade password protection to individual emails.",
      color: "text-primary-400",
      animation: "lock-shut"
    },
    {
      icon: Timer,
      title: "Time Lock",
      description: "Schedule when emails become accessible. Future-proof your communications.",
      color: "text-secure-400",
      animation: "glow"
    },
    {
      icon: MapPin,
      title: "Remote Revoke",
      description: "Instantly recall emails from anywhere in the world. Ultimate control.",
      color: "text-primary-400",
      animation: "float"
    },
    {
      icon: Shield,
      title: "Strip Metadata",
      description: "Remove all identifying information. Complete anonymity guaranteed.",
      color: "text-secure-400",
      animation: "fade-in"
    },
    {
      icon: AlertTriangle,
      title: "Tamper Alerts",
      description: "Get notified instantly if anyone attempts unauthorized access.",
      color: "text-primary-400",
      animation: "lock-shut"
    },
    {
      icon: Trash2,
      title: "Self-Destruct After Failed Attempts",
      description: "Auto-delete emails after multiple failed access attempts.",
      color: "text-secure-400",
      animation: "glow"
    },
    {
      icon: Key,
      title: "Quantum-Resistant Encryption",
      description: "PQC hybrid encryption ready for the quantum computing era.",
      color: "text-primary-400",
      animation: "float"
    },
    {
      icon: Zap,
      title: "Zero-Knowledge Architecture",
      description: "We cannot see your emails. Not content, not metadata, nothing.",
      color: "text-secure-400",
      animation: "fade-in"
    },
    {
      icon: Fingerprint,
      title: "Biometric Access",
      description: "Secure access using your fingerprint or face ID.",
      color: "text-primary-400",
      animation: "lock-shut"
    },
    {
      icon: Lock,
      title: "Closed-Source Fortress",
      description: "No leaks. No exposure. No attack surface. Our code is sealed from prying eyes, making SecureMail immune to the risks of open-source exploitation.",
      color: "text-secure-400",
      animation: "glow"
    }
  ]

  return (
    <section ref={ref} className="section-padding bg-gradient-to-b from-gray-100 to-gray-200 dark:from-dark-800 dark:to-dark-900">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-20"
        >
          <h2 className="text-4xl md:text-6xl font-bold mb-6">
            <span className="gradient-text">Security Beyond</span>
            <br />
            <span className="text-dark-900 dark:text-white">Imagination</span>
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-4xl mx-auto mb-8">
            No other email system on Earth offers this level of security. Every feature is designed 
            to give you absolute control over your privacy and data.
          </p>
          
          {/* Marketing Highlight */}
          <div className="glass-effect rounded-2xl p-6 max-w-3xl mx-auto">
            <p className="text-2xl md:text-3xl font-bold text-dark-900 dark:text-white mb-2">
              "Security beyond imagination."
            </p>
            <p className="text-lg text-gray-600 dark:text-gray-300">
              No other email system on Earth offers this.
            </p>
          </div>
        </motion.div>

        {/* Features Grid */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {securityFeatures.map((feature, index) => (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 30, scale: 0.8 }}
                animate={isInView ? { opacity: 1, y: 0, scale: 1 } : {}}
                transition={{ 
                  duration: 0.6, 
                  delay: 0.4 + index * 0.1,
                  type: "spring",
                  stiffness: 100
                }}
                whileHover={{ 
                  scale: 1.05,
                  y: -5,
                  transition: { duration: 0.2 }
                }}
                className="feature-card group cursor-pointer"
              >
                <motion.div 
                  className={`w-16 h-16 bg-gradient-to-br from-secure-500/20 to-primary-500/20 rounded-2xl flex items-center justify-center mb-4 mx-auto group-hover:scale-110 transition-transform duration-300`}
                  whileHover={{ 
                    rotate: 5,
                    transition: { duration: 0.2 }
                  }}
                >
                  <feature.icon className={`w-8 h-8 ${feature.color}`} />
                </motion.div>
                
                <h4 className="text-lg font-semibold mb-3 text-dark-900 dark:text-white text-center">
                  {feature.title}
                </h4>
                
                <p className="text-sm text-gray-700 dark:text-gray-300 text-center leading-relaxed">
                  {feature.description}
                </p>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Bottom CTA */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="text-center mt-20"
        >
          <div className="glass-effect rounded-3xl p-8 md:p-12 max-w-4xl mx-auto">
            <motion.div
              animate={{ 
                scale: [1, 1.05, 1],
                rotate: [0, 5, -5, 0]
              }}
              transition={{ 
                duration: 4,
                repeat: Infinity,
                repeatType: "reverse"
              }}
              className="w-20 h-20 bg-gradient-to-br from-secure-500 to-primary-500 rounded-full flex items-center justify-center mx-auto mb-6"
            >
              <Shield className="w-10 h-10 text-white" />
            </motion.div>
            
            <h3 className="text-3xl md:text-4xl font-bold mb-4 text-dark-900 dark:text-white">
              This is Just the Beginning
            </h3>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-8 max-w-3xl mx-auto">
              SecureMail is constantly evolving with new security features. What you see here is just 
              the foundation of what's possible when security is the priority, not an afterthought.
            </p>
            
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-8 py-3"
            >
              Explore All Features
            </motion.button>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
