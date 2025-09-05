import { motion } from 'framer-motion'
import { useInView } from 'framer-motion'
import { useRef } from 'react'
import { Shield, CheckCircle, Building2, Users, Award, Globe, Lock } from 'lucide-react'

export default function Trust() {
  const ref = useRef(null)
  const isInView = useInView(ref, { once: true, margin: "-100px" })

  const complianceFeatures = [
    {
      icon: Shield,
      title: "GDPR Compliant",
      description: "Full compliance with European data protection regulations",
      color: "text-secure-400"
    },
    {
      icon: Building2,
      title: "HIPAA Ready",
      description: "Healthcare industry compliance for sensitive medical data",
      color: "text-primary-400"
    },
    {
      icon: Award,
      title: "SOC2 Type II",
      description: "Enterprise-grade security certification and auditing",
      color: "text-secure-400"
    },
    {
      icon: Globe,
      title: "Global Standards",
      description: "Meets international security and privacy standards",
      color: "text-primary-400"
    }
  ]

  const enterpriseFeatures = [
    {
      icon: Users,
      title: "Team Management",
      description: "Secure collaboration with role-based access controls"
    },
    {
      icon: Lock,
      title: "Admin Controls",
      description: "Centralized security policies and monitoring"
    },
    {
      icon: Shield,
      title: "Audit Logs",
      description: "Complete visibility into security events and access"
    },
    {
      icon: CheckCircle,
      title: "Compliance Reporting",
      description: "Automated reports for regulatory requirements"
    }
  ]

  return (
    <section id="trust" ref={ref} className="section-padding bg-gradient-to-b from-gray-200 to-gray-100 dark:from-dark-900 dark:to-dark-800">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-20"
        >
                     <h2 className="text-4xl md:text-6xl font-bold mb-6">
             <span className="gradient-text">Trust &</span>
             <br />
             <span className="text-dark-900 dark:text-white">Compliance</span>
           </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            Built for enterprise, accessible to everyone. SecureMail meets the highest 
            standards of security and compliance in the industry.
          </p>
        </motion.div>

        {/* Compliance Features */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="mb-20"
        >
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {complianceFeatures.map((feature, index) => (
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
                className="feature-card text-center group"
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
                
                                 <h4 className="text-xl font-semibold mb-2 text-dark-900 dark:text-white">
                   {feature.title}
                 </h4>
                 
                 <p className="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                   {feature.description}
                 </p>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Enterprise Section */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.6 }}
          className="mb-20"
        >
          <div className="glass-effect rounded-3xl p-8 md:p-12">
            <div className="text-center mb-12">
              <motion.div
                initial={{ scale: 0 }}
                animate={isInView ? { scale: 1 } : {}}
                transition={{ duration: 0.6, delay: 0.8 }}
                className="w-20 h-20 bg-gradient-to-br from-secure-500 to-primary-500 rounded-full flex items-center justify-center mx-auto mb-6"
              >
                <Building2 className="w-10 h-10 text-white" />
              </motion.div>
              
                             <h3 className="text-3xl md:text-4xl font-bold mb-4 text-dark-900 dark:text-white">
                 Enterprise Security, Now for Everyone
               </h3>
              
              <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
                SecureMail brings enterprise-grade security features to individuals and teams, 
                making military-level protection accessible to everyone.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {enterpriseFeatures.map((feature, index) => (
                <motion.div
                  key={feature.title}
                  initial={{ opacity: 0, y: 30 }}
                  animate={isInView ? { opacity: 1, y: 0 } : {}}
                  transition={{ duration: 0.6, delay: 1.0 + index * 0.1 }}
                  className="text-center"
                >
                  <div className="w-12 h-12 bg-secure-500/20 rounded-xl flex items-center justify-center mx-auto mb-4">
                    <feature.icon className="w-6 h-6 text-secure-400" />
                  </div>
                  
                                   <h4 className="text-lg font-semibold mb-2 text-dark-900 dark:text-white">
                   {feature.title}
                 </h4>
                 
                 <p className="text-sm text-gray-700 dark:text-gray-300">
                   {feature.description}
                 </p>
                </motion.div>
              ))}
            </div>
          </div>
        </motion.div>

        {/* Trust Statement */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 1.2 }}
          className="text-center"
        >
          <div className="glass-effect rounded-2xl p-8 max-w-3xl mx-auto">
                         <h3 className="text-2xl md:text-3xl font-bold mb-4 text-dark-900 dark:text-white">
               Trusted by Security Experts Worldwide
             </h3>
                         <p className="text-lg text-gray-700 dark:text-gray-300 mb-6">
              Unlike others, SecureMail is not open-source. Our encryption architecture is proprietary and locked down. No one outside our core team can see, study, or exploit our code — eliminating the open-source attack surface while still being built on peer-reviewed cryptographic standards.
            </p>
            <div className="flex flex-wrap justify-center items-center gap-6 text-sm text-gray-600 dark:text-gray-400">
              <div className="flex items-center space-x-2">
                <Lock className="w-4 h-4 text-secure-400" />
                <span>Closed-Source Fortress</span>
              </div>
              <div className="flex items-center space-x-2">
                <CheckCircle className="w-4 h-4 text-secure-400" />
                <span>Independent Audits</span>
              </div>
              <div className="flex items-center space-x-2">
                <CheckCircle className="w-4 h-4 text-secure-400" />
                <span>Proprietary Architecture</span>
              </div>
            </div>
          </div>
        </motion.div>

        {/* Enhanced Privacy Statement - Phase 2 Addition */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 1.4 }}
          className="text-center mt-12"
        >
          <div className="glass-effect rounded-2xl p-8 max-w-4xl mx-auto border-2 border-secure-400/20">
            <motion.div
              animate={{ 
                scale: [1, 1.05, 1],
                rotate: [0, 2, -2, 0]
              }}
              transition={{ 
                duration: 6,
                repeat: Infinity,
                repeatType: "reverse"
              }}
              className="w-16 h-16 bg-gradient-to-br from-secure-500 to-primary-500 rounded-full flex items-center justify-center mx-auto mb-6"
            >
              <Shield className="w-8 h-8 text-white" />
            </motion.div>
            
            <h3 className="text-2xl md:text-3xl font-bold mb-4 text-dark-900 dark:text-white">
              Even Our Website Doesn't Track You
            </h3>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
              <div className="text-center">
                <div className="w-12 h-12 bg-red-500/20 rounded-full flex items-center justify-center mx-auto mb-3">
                  <span className="text-2xl">🚫</span>
                </div>
                <p className="text-lg font-semibold text-dark-900 dark:text-white">No Cookies</p>
                <p className="text-sm text-gray-600 dark:text-gray-400">Zero tracking cookies</p>
              </div>
              <div className="text-center">
                <div className="w-12 h-12 bg-red-500/20 rounded-full flex items-center justify-center mx-auto mb-3">
                  <span className="text-2xl">👁️</span>
                </div>
                <p className="text-lg font-semibold text-dark-900 dark:text-white">No Tracking</p>
                <p className="text-sm text-gray-600 dark:text-gray-400">No analytics or tracking</p>
              </div>
              <div className="text-center">
                <div className="w-12 h-12 bg-red-500/20 rounded-full flex items-center justify-center mx-auto mb-3">
                  <span className="text-2xl">📊</span>
                </div>
                <p className="text-lg font-semibold text-dark-900 dark:text-white">No Analytics</p>
                <p className="text-sm text-gray-600 dark:text-gray-400">Privacy by design</p>
              </div>
            </div>
            
            <p className="text-lg text-gray-700 dark:text-gray-300 mb-6">
              We practice what we preach. This website uses zero tracking, zero cookies, and zero analytics. 
              Your privacy starts from the moment you visit us.
            </p>
            
            {/* Code Integrity Verification */}
            <div className="bg-dark-900 dark:bg-dark-800 rounded-xl p-4 max-w-2xl mx-auto">
              <p className="text-sm text-gray-400 dark:text-gray-400 mb-2">
                Verify this page's authenticity:
              </p>
              <p className="text-xs text-secure-400 font-mono">
                Check our published signed hash: <span className="text-primary-400">securemail.com/verify</span>
              </p>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
