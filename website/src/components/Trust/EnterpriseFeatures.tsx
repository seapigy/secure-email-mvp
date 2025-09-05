import { MotionDiv } from '../../utils/animations'
import { Users, Lock, Shield } from '../../utils/icons'

export default function EnterpriseFeatures() {
  const enterpriseFeatures = [
    {
      icon: Users,
      title: "Team Management",
      description: "Advanced user management and role-based access controls",
      color: "text-indigo-500"
    },
    {
      icon: Lock,
      title: "Admin Controls",
      description: "Comprehensive administrative controls and monitoring",
      color: "text-blue-500"
    },
    {
      icon: Shield,
      title: "Audit Logs",
      description: "Complete audit trail for compliance and security monitoring",
      color: "text-indigo-500"
    }
  ]

  return (
    <MotionDiv
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.8, delay: 0.4 }}
      className="text-center"
    >
      <h3 className="text-3xl md:text-4xl font-bold mb-12 text-dark-900 dark:text-white">
        Enterprise Features
      </h3>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
        {enterpriseFeatures.map((feature, index) => (
          <MotionDiv
            key={feature.title}
            initial={{ opacity: 0, y: 30, scale: 0.8 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
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
      
      {/* Enterprise CTA */}
      <MotionDiv
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.8, delay: 0.8 }}
        className="glass-effect rounded-3xl p-8 md:p-12 max-w-4xl mx-auto"
      >
        <h4 className="text-2xl md:text-3xl font-bold mb-4 text-dark-900 dark:text-white">
          Ready for Enterprise?
        </h4>
        
        <p className="text-lg text-gray-600 dark:text-gray-300 mb-8 max-w-2xl mx-auto">
          Contact our enterprise team to learn about custom solutions, 
          dedicated support, and advanced security features.
        </p>
        
        <MotionDiv
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          className="btn-primary text-lg px-8 py-3"
          as="button"
        >
          Contact Enterprise Sales
        </MotionDiv>
      </MotionDiv>
    </MotionDiv>
  )
}