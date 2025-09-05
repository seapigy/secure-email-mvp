import { MotionDiv, useInView } from '../utils/animations'
import { useRef } from 'react'
import { 
  Globe, 
  Clock, 
  Eye, 
  EyeOff,
  Lock, 
  Timer, 
  MapPin, 
  Shield, 
  AlertTriangle, 
  Trash2, 
  Key,
  Plus
} from '../utils/icons'

export default function Features() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })

  const features = [
    {
      icon: Shield,
      title: "Zero-Knowledge Architecture",
      description: "We cannot see your emails. Not content, not metadata, nothing. Complete privacy by design.",
      color: "text-indigo-500"
    },
    {
      icon: Key,
      title: "AES-256-GCM Encryption",
      description: "Military-grade encryption that protects your data with the same standards used by governments.",
      color: "text-blue-500"
    },
    {
      icon: Lock,
      title: "End-to-End Security",
      description: "Your messages are encrypted on your device and only decrypted by the recipient.",
      color: "text-indigo-500"
    },
    {
      icon: EyeOff,
      title: "Metadata Protection",
      description: "We don't store who you're talking to, when, or how often. Complete communication privacy.",
      color: "text-blue-500"
    },
    {
      icon: Globe,
      title: "Global Infrastructure",
      description: "Secure servers worldwide ensure fast, reliable access while maintaining security standards.",
      color: "text-indigo-500"
    },
    {
      icon: Clock,
      title: "Real-Time Delivery",
      description: "Lightning-fast message delivery without compromising on security or privacy.",
      color: "text-blue-500"
    }
  ]

  return (
    <section ref={ref} id="features" className="section-padding bg-white dark:bg-dark-800">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-20"
        >
          <h2 className="text-4xl md:text-6xl font-bold mb-6">
            <span className="gradient-text">Core Security</span>
            <br />
            <span className="text-dark-900 dark:text-white">Features</span>
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-4xl mx-auto mb-8">
            Essential security features that form the foundation of truly secure email communication. 
            These aren't optional extras—they're the minimum standard for real privacy.
          </p>
          
          {/* Marketing Highlight */}
          <div className="glass-effect rounded-2xl p-6 max-w-3xl mx-auto">
            <p className="text-2xl md:text-3xl font-bold text-dark-900 dark:text-white mb-2">
              "Security that actually works."
            </p>
            <p className="text-lg text-gray-600 dark:text-gray-300">
              No compromises. No backdoors. No surveillance.
            </p>
          </div>
        </MotionDiv>

        {/* Features Grid */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature, index) => (
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
              The Foundation of Trust
            </h3>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-8 max-w-3xl mx-auto">
              These core features aren't just checkboxes—they're the essential building blocks 
              of truly secure communication. Without them, you don't have real security.
            </p>
            
            <MotionDiv
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-8 py-3"
              as="button"
            >
              Learn More About Security
            </MotionDiv>
          </div>
        </MotionDiv>
      </div>
    </section>
  )
}