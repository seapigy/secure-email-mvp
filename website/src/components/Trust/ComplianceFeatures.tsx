import { MotionDiv } from '../../utils/animations'
import { Shield, Building2, Award, Globe } from '../../utils/icons'

export default function ComplianceFeatures() {
  const complianceFeatures = [
    {
      icon: Shield,
      title: "GDPR Compliant",
      description: "Full compliance with European data protection regulations",
      color: "text-indigo-500"
    },
    {
      icon: Building2,
      title: "SOC 2 Type II",
      description: "Audited security controls and operational procedures",
      color: "text-blue-500"
    },
    {
      icon: Award,
      title: "ISO 27001",
      description: "International standard for information security management",
      color: "text-indigo-500"
    },
    {
      icon: Globe,
      title: "CCPA Compliant",
      description: "California Consumer Privacy Act compliance",
      color: "text-blue-500"
    }
  ]

  return (
    <MotionDiv
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.8, delay: 0.2 }}
      className="mb-20"
    >
      <h3 className="text-3xl md:text-4xl font-bold text-center mb-12 text-dark-900 dark:text-white">
        Compliance & Certifications
      </h3>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
        {complianceFeatures.map((feature, index) => (
          <MotionDiv
            key={feature.title}
            initial={{ opacity: 0, y: 30, scale: 0.8 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
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
  )
}