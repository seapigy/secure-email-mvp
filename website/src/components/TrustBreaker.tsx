import { MotionDiv, useInView } from '../utils/animations'
import { useRef } from 'react'
import { Shield, Lock, EyeOff, Server, Key, Globe } from '../utils/icons'

export default function TrustBreaker() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })

  const features = [
    {
      icon: Shield,
      title: "AES-256-GCM Encryption",
      description: "Military-grade encryption that's virtually unbreakable, enhanced with Post-Quantum Cryptography (PQC) key exchange",
      color: "text-indigo-500"
    },
    {
      icon: Lock,
      title: "Argon2id Hashing",
      description: "State-of-the-art password hashing with memory-hard functions",
      color: "text-blue-500"
    },
    {
      icon: EyeOff,
      title: "Zero-Knowledge Architecture",
      description: "We cannot see your emails. Not content, not metadata, nothing.",
      color: "text-indigo-500"
    },
    {
      icon: Server,
      title: "Secure Infrastructure",
      description: "Hardened servers with end-to-end encryption and zero-logging policies",
      color: "text-blue-500"
    },
    {
      icon: Key,
      title: "Quantum-Resistant Keys",
      description: "Future-proof encryption ready for the quantum computing era",
      color: "text-indigo-500"
    },
    {
      icon: Globe,
      title: "Global Privacy",
      description: "Your data is protected by the strongest privacy laws worldwide",
      color: "text-blue-500"
    }
  ]

  return (
    <section ref={ref} className="section-padding bg-gradient-to-b from-gray-100 to-gray-200 dark:from-dark-800 dark:to-dark-900 relative overflow-hidden">
      {/* Background Elements */}
      <div className="absolute inset-0">
        <MotionDiv
          animate={{ 
            rotate: [0, 360],
            scale: [1, 1.2, 1]
          }}
          transition={{ 
            duration: 30,
            repeat: Infinity,
            ease: "linear"
          }}
          className="absolute top-10 left-10 w-64 h-64 bg-gradient-to-br from-indigo-500/10 to-blue-500/10 rounded-full blur-3xl"
        />
        <MotionDiv
          animate={{ 
            rotate: [360, 0],
            scale: [1.2, 1, 1.2]
          }}
          transition={{ 
            duration: 35,
            repeat: Infinity,
            ease: "linear"
          }}
          className="absolute bottom-10 right-10 w-80 h-80 bg-gradient-to-br from-blue-500/10 to-indigo-500/10 rounded-full blur-3xl"
        />
      </div>

      <div className="max-w-7xl mx-auto relative z-10">
        {/* Section Header */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-20"
        >
          <h2 className="text-4xl md:text-6xl font-bold mb-6">
            <span className="gradient-text">Why Trust</span>
            <br />
            <span className="text-dark-900 dark:text-white">SecureMail?</span>
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-4xl mx-auto mb-8">
            Because we've built the most secure email system on Earth. Here's how we achieve 
            zero-knowledge architecture and military-grade security.
          </p>
          
          {/* Trust Statement */}
          <div className="glass-effect rounded-2xl p-8 max-w-4xl mx-auto">
            <MotionDiv
              animate={{ 
                scale: [1, 1.02, 1],
                rotate: [0, 1, -1, 0]
              }}
              transition={{ 
                duration: 6,
                repeat: Infinity,
                repeatType: "reverse"
              }}
              className="w-16 h-16 bg-gradient-to-br from-indigo-500 to-blue-500 rounded-full flex items-center justify-center mx-auto mb-4"
            >
              <Shield className="w-8 h-8 text-white" />
            </MotionDiv>
            
            <h3 className="text-2xl md:text-3xl font-bold text-dark-900 dark:text-white mb-4">
              "We Cannot See Your Emails"
            </h3>
            <p className="text-lg text-gray-600 dark:text-gray-300">
              This isn't marketing speak. It's a technical reality. Our zero-knowledge architecture 
              means we literally cannot access your data, even if we wanted to.
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
            <h3 className="text-3xl md:text-4xl font-bold mb-4 text-dark-900 dark:text-white">
              The Technical Reality
            </h3>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-8 max-w-3xl mx-auto">
              This isn't just about features—it's about building a system where privacy 
              is mathematically guaranteed, not just promised.
            </p>
            
            <MotionDiv
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-8 py-3"
              as="button"
            >
              See How It Works
            </MotionDiv>
          </div>
        </MotionDiv>
      </div>
    </section>
  )
}