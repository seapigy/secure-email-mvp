import { MotionDiv } from '../../utils/animations'
import { Shield, Lock, CheckCircle } from '../../utils/icons'

export default function TrustSection() {
  return (
    <MotionDiv
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.8 }}
      className="text-center mb-20"
    >
      <h2 className="text-4xl md:text-6xl font-bold mb-6">
        <span className="gradient-text">Trust & Compliance</span>
        <br />
        <span className="text-dark-900 dark:text-white">Built In</span>
      </h2>
      <p className="text-xl text-gray-600 dark:text-gray-300 max-w-4xl mx-auto mb-8">
        Enterprise-grade security and compliance features that meet the highest standards 
        for privacy, security, and regulatory requirements.
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
          "Verified by Security Experts"
        </h3>
        <p className="text-lg text-gray-600 dark:text-gray-300">
          Our security architecture has been independently verified by leading cybersecurity 
          experts and meets the highest industry standards.
        </p>
      </div>
    </MotionDiv>
  )
}