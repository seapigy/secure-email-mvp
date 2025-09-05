import { MotionDiv, useInView } from '../utils/animations'
import { useRef } from 'react'
import { 
  Globe, 
  Clock, 
  EyeOff,
  Lock, 
  Shield, 
  AlertTriangle, 
  Key,
  Server,
  Zap,
  CheckCircle,
  Timer,
  MapPin,
  Trash2,
  Eye
} from '../utils/icons'

export default function Features() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })

  const features = [
    // Core Security Features
    {
      icon: Shield,
      title: "Zero-Knowledge Architecture",
      description: "We cannot see your emails. Not content, not metadata, nothing. Complete privacy by design.",
      color: "text-indigo-500",
      category: "core"
    },
    {
      icon: Key,
      title: "AES-256-GCM Encryption",
      description: "Military-grade encryption that protects your data with the same standards used by governments. Enhanced with Post-Quantum Cryptography (PQC) key exchange.",
      color: "text-blue-500",
      category: "core"
    },
    {
      icon: Lock,
      title: "End-to-End Security",
      description: "Your messages are encrypted on your device and only decrypted by the recipient.",
      color: "text-indigo-500",
      category: "core"
    },
    {
      icon: EyeOff,
      title: "Metadata Protection",
      description: "We don't store who you're talking to, when, or how often. Complete communication privacy.",
      color: "text-blue-500",
      category: "core"
    },
    {
      icon: Globe,
      title: "Global Infrastructure",
      description: "Secure servers worldwide ensure fast, reliable access while maintaining security standards.",
      color: "text-indigo-500",
      category: "core"
    },
    {
      icon: Clock,
      title: "Real-Time Delivery",
      description: "Lightning-fast message delivery without compromising on security or privacy.",
      color: "text-blue-500",
      category: "core"
    },
    // Advanced Security Features
    {
      icon: AlertTriangle,
      title: "Advanced Threat Detection",
      description: "Real-time monitoring and protection against sophisticated cyber attacks with machine learning algorithms.",
      color: "text-red-500",
      category: "advanced"
    },
    {
      icon: Server,
      title: "Secure Infrastructure",
      description: "Hardened servers with end-to-end encryption and zero-logging policies for maximum security.",
      color: "text-purple-500",
      category: "advanced"
    },
    {
      icon: Zap,
      title: "Quantum-Resistant Keys",
      description: "Future-proof encryption ready for the quantum computing era with post-quantum cryptography.",
      color: "text-green-500",
      category: "advanced"
    },
    // Access Controls
    {
      icon: Lock,
      title: "Password-Protected Links",
      description: "Recipients must know a shared secret to open emails. No unauthorized access possible.",
      color: "text-red-500",
      category: "access"
    },
    {
      icon: CheckCircle,
      title: "2FA/TOTP Verification",
      description: "Time-based one-time codes gate email access. Military-grade authentication for every message.",
      color: "text-blue-500",
      category: "access"
    },
    {
      icon: MapPin,
      title: "Geo-Restriction Controls",
      description: "Restrict email opening to specific countries, states, or IP ranges. Location-based security.",
      color: "text-indigo-500",
      category: "access"
    },
    {
      icon: Eye,
      title: "Device Fingerprinting",
      description: "Allow access only from known recipient devices. Advanced device recognition and control.",
      color: "text-purple-500",
      category: "access"
    },
    {
      icon: Timer,
      title: "Single-Use Access Tokens",
      description: "Once viewed, the link/token expires. No replay attacks or unauthorized re-access possible.",
      color: "text-green-500",
      category: "access"
    },
    // Metadata Protection
    {
      icon: EyeOff,
      title: "Header Stripping",
      description: "Remove sender IP, device details, and routing information. Complete metadata anonymity.",
      color: "text-red-500",
      category: "metadata"
    },
    {
      icon: Shield,
      title: "Metadata Encryption",
      description: "Encrypt subject lines, timestamps, and routing info. Even metadata is protected.",
      color: "text-blue-500",
      category: "metadata"
    },
    {
      icon: Zap,
      title: "Decoy Subjects & Steganography",
      description: "Make traffic analysis harder with decoy subjects and steganographic techniques.",
      color: "text-indigo-500",
      category: "metadata"
    },
    // Ephemeral Controls
    {
      icon: Timer,
      title: "Self-Destruct Timers",
      description: "Message deletes after X minutes/hours/days. Automatic cleanup for sensitive communications.",
      color: "text-red-500",
      category: "ephemeral"
    },
    {
      icon: Eye,
      title: "View Limits",
      description: "Auto-delete after 1-3 opens. Control how many times a message can be viewed.",
      color: "text-blue-500",
      category: "ephemeral"
    },
    {
      icon: Trash2,
      title: "Auto-Delete on Failed Attempts",
      description: "Wipe message after 3 failed login attempts. Protection against brute force attacks.",
      color: "text-purple-500",
      category: "ephemeral"
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
            <span className="gradient-text">Complete Security</span>
            <br />
            <span className="text-dark-900 dark:text-white">Features</span>
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-4xl mx-auto mb-8">
            From core encryption to advanced access controls, metadata protection, and ephemeral messaging, 
            SecureMail provides comprehensive security that goes far beyond what other email providers offer. 
            Every feature applies to every email with granular control.
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

        {/* Core Features Section */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="mb-16"
        >
          <h3 className="text-2xl md:text-3xl font-bold text-center mb-8 text-dark-900 dark:text-white">
            Core Security Foundation
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.filter(f => f.category === 'core').map((feature, index) => (
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

        {/* Advanced Features Section */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.4 }}
        >
          <h3 className="text-2xl md:text-3xl font-bold text-center mb-8 text-dark-900 dark:text-white">
            Advanced Security Arsenal
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.filter(f => f.category === 'advanced').map((feature, index) => (
              <MotionDiv
                key={feature.title}
                initial={{ opacity: 0, y: 30, scale: 0.8 }}
                animate={isInView ? { opacity: 1, y: 0, scale: 1 } : {}}
                transition={{ 
                  duration: 0.6, 
                  delay: 0.6 + index * 0.1,
                  type: "spring",
                  stiffness: 100
                }}
                whileHover={{ 
                  scale: 1.05,
                  y: -5,
                  transition: { duration: 0.2 }
                }}
                className="feature-card group cursor-pointer border-2 border-transparent hover:border-gradient-to-r hover:from-red-500/20 hover:to-purple-500/20"
              >
                <MotionDiv 
                  className={`w-16 h-16 bg-gradient-to-br from-red-500/20 to-purple-500/20 rounded-2xl flex items-center justify-center mb-4 mx-auto group-hover:scale-110 transition-transform duration-300`}
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

        {/* Access Controls Section */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.6 }}
          className="mb-16"
        >
          <h3 className="text-2xl md:text-3xl font-bold text-center mb-8 text-dark-900 dark:text-white">
            Advanced Access Controls
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.filter(f => f.category === 'access').map((feature, index) => (
              <MotionDiv
                key={feature.title}
                initial={{ opacity: 0, y: 30, scale: 0.8 }}
                animate={isInView ? { opacity: 1, y: 0, scale: 1 } : {}}
                transition={{ 
                  duration: 0.6, 
                  delay: 0.8 + index * 0.1,
                  type: "spring",
                  stiffness: 100
                }}
                whileHover={{ 
                  scale: 1.05,
                  y: -5,
                  transition: { duration: 0.2 }
                }}
                className="feature-card group cursor-pointer border-2 border-transparent hover:border-gradient-to-r hover:from-red-500/20 hover:to-blue-500/20"
              >
                <MotionDiv 
                  className={`w-16 h-16 bg-gradient-to-br from-red-500/20 to-blue-500/20 rounded-2xl flex items-center justify-center mb-4 mx-auto group-hover:scale-110 transition-transform duration-300`}
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

        {/* Metadata Protection Section */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="mb-16"
        >
          <h3 className="text-2xl md:text-3xl font-bold text-center mb-8 text-dark-900 dark:text-white">
            Metadata Protection
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.filter(f => f.category === 'metadata').map((feature, index) => (
              <MotionDiv
                key={feature.title}
                initial={{ opacity: 0, y: 30, scale: 0.8 }}
                animate={isInView ? { opacity: 1, y: 0, scale: 1 } : {}}
                transition={{ 
                  duration: 0.6, 
                  delay: 1.0 + index * 0.1,
                  type: "spring",
                  stiffness: 100
                }}
                whileHover={{ 
                  scale: 1.05,
                  y: -5,
                  transition: { duration: 0.2 }
                }}
                className="feature-card group cursor-pointer border-2 border-transparent hover:border-gradient-to-r hover:from-indigo-500/20 hover:to-purple-500/20"
              >
                <MotionDiv 
                  className={`w-16 h-16 bg-gradient-to-br from-indigo-500/20 to-purple-500/20 rounded-2xl flex items-center justify-center mb-4 mx-auto group-hover:scale-110 transition-transform duration-300`}
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

        {/* Ephemeral Controls Section */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 1.0 }}
          className="mb-16"
        >
          <h3 className="text-2xl md:text-3xl font-bold text-center mb-8 text-dark-900 dark:text-white">
            Ephemeral Controls
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.filter(f => f.category === 'ephemeral').map((feature, index) => (
              <MotionDiv
                key={feature.title}
                initial={{ opacity: 0, y: 30, scale: 0.8 }}
                animate={isInView ? { opacity: 1, y: 0, scale: 1 } : {}}
                transition={{ 
                  duration: 0.6, 
                  delay: 1.2 + index * 0.1,
                  type: "spring",
                  stiffness: 100
                }}
                whileHover={{ 
                  scale: 1.05,
                  y: -5,
                  transition: { duration: 0.2 }
                }}
                className="feature-card group cursor-pointer border-2 border-transparent hover:border-gradient-to-r hover:from-purple-500/20 hover:to-green-500/20"
              >
                <MotionDiv 
                  className={`w-16 h-16 bg-gradient-to-br from-purple-500/20 to-green-500/20 rounded-2xl flex items-center justify-center mb-4 mx-auto group-hover:scale-110 transition-transform duration-300`}
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
              Military-Grade Security Controls
            </h3>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-8 max-w-3xl mx-auto">
              From access controls to ephemeral messaging, every feature applies to every email. 
              This isn't just security—it's comprehensive protection with granular control that actually works.
            </p>
            
            <MotionDiv
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-8 py-3"
              as="button"
            >
              Experience Complete Security
            </MotionDiv>
          </div>
        </MotionDiv>
      </div>
    </section>
  )
}