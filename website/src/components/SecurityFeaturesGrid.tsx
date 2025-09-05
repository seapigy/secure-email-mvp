import { MotionDiv, useInView } from '../utils/animations'
import { useRef } from 'react'
import { 
  Globe, 
  Clock, 
  Eye, 
  Lock, 
  Shield, 
  AlertTriangle, 
  Key,
  Zap
} from '../utils/icons'

// DO NOT EDIT EXISTING CODE - This is a new component for Phase 2

export default function SecurityFeaturesGrid() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })

  const securityFeatures = [
    {
      icon: Zap,
      title: "Zero-Knowledge Architecture",
      description: "We cannot see your emails. Not content, not metadata, nothing. Complete privacy by design.",
      color: "text-indigo-500",
      animation: "fade-in"
    },
    {
      icon: Key,
      title: "Quantum-Resistant Encryption",
      description: "PQC hybrid encryption ready for the quantum computing era. Future-proof security.",
      color: "text-blue-500",
      animation: "float"
    },
    {
      icon: Shield,
      title: "Decoy Messages",
      description: "Create fake emails to mislead attackers. Advanced deception technology.",
      color: "text-indigo-500",
      animation: "glow"
    },
    {
      icon: Fingerprint,
      title: "Biometric Access",
      description: "Secure access using your fingerprint or face ID. Next-gen authentication.",
      color: "text-blue-500",
      animation: "lock-shut"
    },
    {
      icon: Lock,
      title: "Closed-Source Fortress",
      description: "No leaks. No exposure. No attack surface. Our code is sealed from prying eyes.",
      color: "text-indigo-500",
      animation: "glow"
    },
    {
      icon: Globe,
      title: "Advanced Geolocation",
      description: "Multi-layered location-based security with city-level precision and compliance.",
      color: "text-blue-500",
      animation: "float"
    },
    {
      icon: Clock,
      title: "Smart Destruction",
      description: "AI-powered email lifecycle management with intelligent deletion patterns.",
      color: "text-indigo-500",
      animation: "fade-in"
    },
    {
      icon: AlertTriangle,
      title: "Threat Intelligence",
      description: "Real-time threat detection and response with machine learning algorithms.",
      color: "text-blue-500",
      animation: "lock-shut"
    },
    {
      icon: Trash2,
      title: "Quantum Erasure",
      description: "Physically impossible data recovery using quantum mechanics principles.",
      color: "text-indigo-500",
      animation: "glow"
    },
    {
      icon: MapPin,
      title: "Stealth Mode",
      description: "Invisible email delivery with no digital footprint or traceability.",
      color: "text-blue-500",
      animation: "float"
    },
    {
      icon: Eye,
      title: "Holographic Security",
      description: "Multi-dimensional access control with impossible-to-replicate security layers.",
      color: "text-indigo-500",
      animation: "fade-in"
    },
    {
      icon: Timer,
      title: "Temporal Encryption",
      description: "Time-based encryption keys that change every millisecond for ultimate security.",
      color: "text-blue-500",
      animation: "lock-shut"
    }
  ]

  return (
    <section ref={ref} className="section-padding bg-gradient-to-b from-gray-100 to-gray-200 dark:from-dark-800 dark:to-dark-900">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <MotionDiv
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
            Revolutionary security features that push the boundaries of what's possible. 
            These advanced capabilities exist nowhere else in the email world.
          </p>
          
          {/* Marketing Highlight */}
          <div className="glass-effect rounded-2xl p-6 max-w-3xl mx-auto">
            <p className="text-2xl md:text-3xl font-bold text-dark-900 dark:text-white mb-2">
              "Security beyond imagination."
            </p>
            <p className="text-lg text-gray-600 dark:text-gray-300">
              Cutting-edge technology that redefines email security.
            </p>
          </div>
        </MotionDiv>

        {/* Features Grid */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {securityFeatures.map((feature, index) => (
              <MotionDiv
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
                <MotionDiv 
                  className={`w-16 h-16 bg-gradient-to-br from-indigo-500/20 to-blue-500/20 rounded-2xl flex items-center justify-center mb-4 mx-auto group-hover:scale-110 transition-transform duration-300`}
                  whileHover={{ 
                    rotate: 5,
                    transition: { duration: 0.2 }
                  }}
                >
                  <feature.icon className={`w-8 h-8 ${feature.color}`} />
                </MotionDiv>
                
                <h4 className="text-lg font-semibold mb-3 text-dark-900 dark:text-white text-center">
                  {feature.title}
                </h4>
                
                <p className="text-sm text-gray-700 dark:text-gray-300 text-center leading-relaxed">
                  {feature.description}
                </p>
              </MotionDiv>
            ))}
          </div>
        </MotionDiv>

        {/* Bottom CTA */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="text-center mt-20"
        >
          <div className="glass-effect rounded-3xl p-8 md:p-12 max-w-4xl mx-auto">
            <MotionDiv
              animate={{ 
                scale: [1, 1.05, 1],
                rotate: [0, 5, -5, 0]
              }}
              transition={{ 
                duration: 4,
                repeat: Infinity,
                repeatType: "reverse"
              }}
              className="w-20 h-20 bg-gradient-to-br from-indigo-500 to-blue-500 rounded-full flex items-center justify-center mx-auto mb-6"
            >
              <Shield className="w-10 h-10 text-white" />
            </MotionDiv>
            
            <h3 className="text-3xl md:text-4xl font-bold mb-4 text-dark-900 dark:text-white">
              The Future of Email Security
            </h3>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-8 max-w-3xl mx-auto">
              These revolutionary features represent the cutting edge of email security technology. 
              We're not just building secure email—we're reimagining what's possible.
            </p>
            
            <MotionDiv
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-8 py-3"
              as="button"
            >
              Explore All Features
            </MotionDiv>
          </div>
        </MotionDiv>
      </div>
    </section>
  )
}