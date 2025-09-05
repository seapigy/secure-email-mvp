import { MotionDiv, useInView } from '../utils/animations'
import { useRef } from 'react'
import { Check, X, Shield, Lock, EyeOff, Zap, Globe, Clock } from '../utils/icons'

export default function Comparison() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })

  const features = [
    {
      name: "End-to-End Encryption",
      securemail: true,
      gmail: false,
      outlook: false,
      protonmail: true
    },
    {
      name: "Zero-Knowledge Architecture",
      securemail: true,
      gmail: false,
      outlook: false,
      protonmail: false
    },
    {
      name: "Quantum-Resistant Encryption",
      securemail: true,
      gmail: false,
      outlook: false,
      protonmail: false
    },
    {
      name: "Metadata Protection",
      securemail: true,
      gmail: false,
      outlook: false,
      protonmail: true
    },
    {
      name: "No Data Collection",
      securemail: true,
      gmail: false,
      outlook: false,
      protonmail: true
    },
    {
      name: "Open Source",
      securemail: false,
      gmail: false,
      outlook: false,
      protonmail: true
    },
    {
      name: "Free Tier",
      securemail: true,
      gmail: true,
      outlook: true,
      protonmail: true
    },
    {
      name: "Mobile Apps",
      securemail: true,
      gmail: true,
      outlook: true,
      protonmail: true
    }
  ]

  return (
    <section ref={ref} id="comparison" className="section-padding bg-white dark:bg-dark-800">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-20"
        >
          <h2 className="text-4xl md:text-6xl font-bold mb-6">
            <span className="gradient-text">How We Compare</span>
            <br />
            <span className="text-dark-900 dark:text-white">to the Competition</span>
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-4xl mx-auto mb-8">
            See how SecureMail stacks up against the biggest names in email. 
            Spoiler: we're in a league of our own.
          </p>
        </MotionDiv>

        {/* Comparison Table */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="overflow-x-auto"
        >
          <div className="min-w-full bg-white dark:bg-dark-800 rounded-2xl shadow-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
            {/* Table Header */}
            <div className="bg-gradient-to-r from-indigo-500 to-blue-500 text-white">
              <div className="grid grid-cols-5 gap-4 p-6">
                <div className="font-semibold text-lg">Feature</div>
                <div className="text-center font-semibold text-lg">SecureMail</div>
                <div className="text-center font-semibold text-lg">Gmail</div>
                <div className="text-center font-semibold text-lg">Outlook</div>
                <div className="text-center font-semibold text-lg">ProtonMail</div>
              </div>
            </div>

            {/* Table Body */}
            <div className="divide-y divide-gray-200 dark:divide-gray-700">
              {features.map((feature, index) => (
                <MotionDiv
                  key={feature.name}
                  initial={{ opacity: 0, x: -20 }}
                  animate={isInView ? { opacity: 1, x: 0 } : {}}
                  transition={{ duration: 0.5, delay: 0.4 + index * 0.05 }}
                  className="grid grid-cols-5 gap-4 p-6 hover:bg-gray-50 dark:hover:bg-dark-700 transition-colors"
                >
                  <div className="font-medium text-dark-900 dark:text-white">
                    {feature.name}
                  </div>
                  <div className="text-center">
                    {feature.securemail ? (
                      <Check className="w-6 h-6 text-green-500 mx-auto" />
                    ) : (
                      <X className="w-6 h-6 text-red-500 mx-auto" />
                    )}
                  </div>
                  <div className="text-center">
                    {feature.gmail ? (
                      <Check className="w-6 h-6 text-green-500 mx-auto" />
                    ) : (
                      <X className="w-6 h-6 text-red-500 mx-auto" />
                    )}
                  </div>
                  <div className="text-center">
                    {feature.outlook ? (
                      <Check className="w-6 h-6 text-green-500 mx-auto" />
                    ) : (
                      <X className="w-6 h-6 text-red-500 mx-auto" />
                    )}
                  </div>
                  <div className="text-center">
                    {feature.protonmail ? (
                      <Check className="w-6 h-6 text-green-500 mx-auto" />
                    ) : (
                      <X className="w-6 h-6 text-red-500 mx-auto" />
                    )}
                  </div>
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
          className="text-center mt-20"
        >
          <div className="glass-effect rounded-3xl p-8 md:p-12 max-w-4xl mx-auto">
            <h3 className="text-3xl md:text-4xl font-bold mb-4 text-dark-900 dark:text-white">
              The Clear Winner
            </h3>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-8 max-w-3xl mx-auto">
              While others compromise on privacy and security, SecureMail delivers 
              uncompromising protection without sacrificing usability.
            </p>
            
            <MotionDiv
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-8 py-3"
              as="button"
            >
              Make the Switch Today
            </MotionDiv>
          </div>
        </MotionDiv>
      </div>
    </section>
  )
}