import { MotionDiv, useInView } from '../utils/animations'
import { useRef } from 'react'
import { Shield, Clock, AlertTriangle, CheckCircle } from '../utils/icons'

export default function QuantumResistant() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })

  const timelineItems = [
    {
      year: "1994",
      title: "Shor's Algorithm Discovery",
      description: "Peter Shor proves quantum computers can break RSA and ECC encryption, establishing the theoretical foundation of quantum threat",
      status: "discovery",
      icon: AlertTriangle
    },
    {
      year: "2000s",
      title: "Early Quantum Research",
      description: "IBM, Google, and others begin serious quantum computing research. First small-scale quantum computers built with 2-5 qubits",
      status: "research",
      icon: Clock
    },
    {
      year: "2010s",
      title: "Quantum Supremacy Race",
      description: "Google achieves quantum supremacy in 2019 with 53-qubit processor. NIST begins Post-Quantum Cryptography standardization process",
      status: "breakthrough",
      icon: Shield
    },
    {
      year: "2020-2024",
      title: "PQC Standardization",
      description: "NIST selects final PQC algorithms (CRYSTALS-Kyber, CRYSTALS-Dilithium). Real-world PQC implementations begin in government and enterprise",
      status: "standardization",
      icon: CheckCircle
    },
    {
      year: "2025",
      title: "Today - SecureMail Leads",
      description: "Current encryption (AES-256) is secure against classical computers, but SecureMail already implements PQC hybrid encryption for future-proof security",
      status: "secure",
      icon: CheckCircle
    },
    {
      year: "2026-2030",
      title: "Quantum Threat Emerges",
      description: "Quantum computers with 1000+ qubits begin breaking current encryption standards. Organizations without PQC face data exposure",
      status: "warning",
      icon: AlertTriangle
    },
    {
      year: "2030+",
      title: "Post-Quantum Era",
      description: "Only quantum-resistant encryption remains secure. Legacy systems become vulnerable to quantum attacks. PQC becomes mandatory for security",
      status: "critical",
      icon: Shield
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
            <span className="gradient-text">Why Quantum-Resistant</span>
            <br />
            <span className="text-dark-900 dark:text-white">Matters</span>
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-4xl mx-auto mb-8">
            The quantum computing revolution is coming. While others scramble to catch up, 
            SecureMail is already quantum-ready with Post-Quantum Cryptography (PQC).
          </p>
          
          {/* Technical Context */}
          <div className="bg-white dark:bg-dark-800 rounded-2xl p-6 max-w-5xl mx-auto mb-8 border border-gray-200 dark:border-gray-700">
            <h3 className="text-2xl font-bold text-dark-900 dark:text-white mb-4 text-center">
              The Quantum Threat Explained
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 text-left">
              <div>
                <h4 className="text-lg font-semibold text-indigo-600 dark:text-indigo-400 mb-2">
                  Why Current Encryption Fails
                </h4>
                <p className="text-gray-600 dark:text-gray-300 text-sm">
                  Shor's algorithm allows quantum computers to factor large numbers exponentially faster than classical computers, 
                  breaking RSA and ECC encryption that secures most of today's internet.
                </p>
              </div>
              <div>
                <h4 className="text-lg font-semibold text-blue-600 dark:text-blue-400 mb-2">
                  The PQC Solution
                </h4>
                <p className="text-gray-600 dark:text-gray-300 text-sm">
                  Post-Quantum Cryptography uses mathematical problems that are hard for both classical and quantum computers, 
                  ensuring security in the post-quantum era.
                </p>
              </div>
            </div>
          </div>
        </MotionDiv>

        {/* Timeline Visual */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="mb-16"
        >
          <div className="relative">
            {/* Timeline Line - Hidden on mobile, visible on desktop */}
            <div className="hidden md:block absolute left-1/2 transform -translate-x-1/2 w-1 h-full bg-gradient-to-b from-green-500 via-yellow-500 to-red-500 rounded-full"></div>
            
            {/* Timeline Items */}
            <div className="space-y-8 md:space-y-12">
              {timelineItems.map((item, index) => (
                <MotionDiv
                  key={item.year}
                  initial={{ opacity: 0, x: index % 2 === 0 ? -50 : 50 }}
                  animate={isInView ? { opacity: 1, x: 0 } : {}}
                  transition={{ duration: 0.6, delay: 0.4 + index * 0.2 }}
                  className={`flex items-center ${
                    // Mobile: always vertical layout
                    'flex-col md:flex-row' + 
                    // Desktop: alternating layout
                    (index % 2 === 0 ? ' md:flex-row' : ' md:flex-row-reverse')
                  }`}
                >
                  {/* Mobile: Full width, Desktop: Half width */}
                  <div className={`w-full md:w-1/2 ${
                    // Mobile: always left-aligned with padding
                    'pl-4 md:pl-0' +
                    // Desktop: alternating alignment
                    (index % 2 === 0 ? ' md:pr-8 md:text-right' : ' md:pl-8 md:text-left')
                  }`}>
                    <div className="bg-white dark:bg-dark-800 rounded-2xl p-4 md:p-6 shadow-lg border border-gray-200 dark:border-gray-700">
                      <div className="flex items-center space-x-3 mb-3">
                        <item.icon className={`w-5 h-5 md:w-6 md:h-6 ${
                          item.status === 'secure' ? 'text-green-500' :
                          item.status === 'warning' ? 'text-yellow-500' :
                          item.status === 'critical' ? 'text-red-500' :
                          item.status === 'discovery' ? 'text-purple-500' :
                          item.status === 'research' ? 'text-blue-500' :
                          item.status === 'breakthrough' ? 'text-orange-500' :
                          item.status === 'standardization' ? 'text-indigo-500' :
                          'text-gray-500'
                        }`} />
                        <span className="text-xl md:text-2xl font-bold text-indigo-600 dark:text-indigo-400">{item.year}</span>
                      </div>
                      <h3 className="text-lg md:text-xl font-semibold text-dark-900 dark:text-white mb-2">{item.title}</h3>
                      <p className="text-sm md:text-base text-gray-600 dark:text-gray-300 leading-relaxed">{item.description}</p>
                    </div>
                  </div>
                  
                  {/* Timeline Dot - Positioned differently for mobile vs desktop */}
                  <div className={`w-3 h-3 md:w-4 md:h-4 rounded-full border-2 md:border-4 border-white dark:border-dark-800 ${
                    item.status === 'secure' ? 'bg-green-500' :
                    item.status === 'warning' ? 'bg-yellow-500' :
                    item.status === 'critical' ? 'bg-red-500' :
                    item.status === 'discovery' ? 'bg-purple-500' :
                    item.status === 'research' ? 'bg-blue-500' :
                    item.status === 'breakthrough' ? 'bg-orange-500' :
                    item.status === 'standardization' ? 'bg-indigo-500' :
                    'bg-gray-500'
                  } ${
                    // Mobile: positioned to the left of content
                    'absolute left-0 top-6 md:relative md:top-0'
                  }`}></div>
                  
                  {/* Desktop spacer - hidden on mobile */}
                  <div className="hidden md:block w-1/2"></div>
                </MotionDiv>
              ))}
            </div>
          </div>
        </MotionDiv>

        {/* Bottom CTA */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="text-center"
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
              Future-Proof Security Today
            </h3>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-8 max-w-3xl mx-auto">
              Don't wait for quantum computers to break your encryption. 
              SecureMail's PQC implementation ensures your data remains secure 
              in the post-quantum world.
            </p>
            
            <MotionDiv
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-8 py-3"
              as="button"
            >
              Learn About PQC
            </MotionDiv>
          </div>
        </MotionDiv>
      </div>
    </section>
  )
}
