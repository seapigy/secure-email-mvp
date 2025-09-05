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
  Plus
} from 'lucide-react'

export default function Features() {
  const ref = useRef(null)
  const isInView = useInView(ref, { once: true, margin: "-100px" })

  const features = [
    {
      icon: Globe,
      title: "Geolocation Locks",
      description: "Restrict access to specific countries or regions",
      animation: "glow"
    },
    {
      icon: Clock,
      title: "Timed Destruction",
      description: "Emails automatically delete after set time",
      animation: "float"
    },
    {
      icon: Eye,
      title: "One-Time Read",
      description: "Emails disappear after being opened once",
      animation: "fade-in"
    },
    {
      icon: Lock,
      title: "Password Protection",
      description: "Add extra layer of security to sensitive emails",
      animation: "lock-shut"
    },
    {
      icon: Timer,
      title: "Time Locks",
      description: "Schedule when emails become accessible",
      animation: "glow"
    },
    {
      icon: MapPin,
      title: "Remote Revoke",
      description: "Instantly recall emails from anywhere",
      animation: "float"
    },
    {
      icon: Shield,
      title: "Decoy Messages",
      description: "Create fake emails to mislead attackers",
      animation: "fade-in"
    },
    {
      icon: AlertTriangle,
      title: "Metadata Stripping",
      description: "Remove all identifying information",
      animation: "lock-shut"
    },
    {
      icon: Trash2,
      title: "Tamper Alerts",
      description: "Get notified of any unauthorized access",
      animation: "glow"
    },
    {
      icon: Key,
      title: "Self-Destruct",
      description: "Auto-delete after failed access attempts",
      animation: "float"
    }
  ]

  return (
    <section id="features" ref={ref} className="section-padding bg-gradient-to-b from-gray-200 to-gray-100 dark:from-dark-900 dark:to-dark-800">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-20"
        >
                     <h2 className="text-4xl md:text-6xl font-bold mb-6">
             <span className="gradient-text">Unmatched Security</span>
             <br />
             <span className="text-dark-900 dark:text-white">Features</span>
           </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            Advanced controls that give you complete command over your email security. 
            Features that don't exist anywhere else in the world.
          </p>
        </motion.div>

        {/* Features Grid */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-6">
            {features.map((feature, index) => (
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
                  className="w-16 h-16 bg-gradient-to-br from-secure-500/20 to-primary-500/20 rounded-2xl flex items-center justify-center mb-4 mx-auto"
                  whileHover={{ 
                    scale: 1.1,
                    rotate: 5,
                    transition: { duration: 0.2 }
                  }}
                >
                  <feature.icon className="w-8 h-8 text-secure-400 group-hover:text-secure-300 transition-colors" />
                </motion.div>
                
                                 <h4 className="text-lg font-semibold mb-2 text-dark-900 dark:text-white text-center">
                   {feature.title}
                 </h4>
                 
                 <p className="text-sm text-gray-700 dark:text-gray-300 text-center leading-relaxed">
                   {feature.description}
                 </p>
              </motion.div>
            ))}
            
            {/* More Features Indicator */}
            <motion.div
              initial={{ opacity: 0, y: 30, scale: 0.8 }}
              animate={isInView ? { opacity: 1, y: 0, scale: 1 } : {}}
              transition={{ 
                duration: 0.6, 
                delay: 0.4 + features.length * 0.1,
                type: "spring",
                stiffness: 100
              }}
              whileHover={{ 
                scale: 1.05,
                y: -5,
                transition: { duration: 0.2 }
              }}
              className="feature-card group cursor-pointer border-dashed border-2 border-secure-400/50 hover:border-secure-400 transition-colors"
            >
              <div className="w-16 h-16 bg-gradient-to-br from-secure-500/10 to-primary-500/10 rounded-2xl flex items-center justify-center mb-4 mx-auto">
                <Plus className="w-8 h-8 text-secure-400 group-hover:text-secure-300 transition-colors" />
              </div>
              
              <h4 className="text-lg font-semibold mb-2 text-secure-400 text-center">
                + More
              </h4>
              
                             <p className="text-sm text-gray-600 dark:text-gray-400 text-center">
                 Continuously expanding security features
               </p>
            </motion.div>
          </div>
        </motion.div>

        {/* Feature Highlight */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="mt-20 text-center"
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
               Enterprise Security, Now for Everyone
             </h3>
            
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-6">
              These aren't just features—they're your digital armor. Every control, 
              every lock, every protection mechanism is designed to give you absolute 
              control over your privacy.
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
