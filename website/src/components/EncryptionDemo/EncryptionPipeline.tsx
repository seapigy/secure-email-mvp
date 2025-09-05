import { MotionDiv } from '../../utils/animations'
import { Shield, Zap, Key, Globe, Cpu, AlertTriangle } from '../../utils/icons'

interface EncryptionPipelineProps {
  currentStep?: number
  isEncrypting?: boolean
}

export default function EncryptionPipeline({ currentStep = 0, isEncrypting = false }: EncryptionPipelineProps) {
  const steps = [
    {
      icon: Key,
      title: "Key Derivation",
      description: "Argon2id password hashing",
      color: "text-gray-500",
      completedColor: "#10b981" // Green for key derivation
    },
    {
      icon: Shield,
      title: "AES-256-GCM",
      description: "Military-grade encryption",
      color: "text-gray-500",
      completedColor: "#3b82f6" // Blue for encryption
    },
    {
      icon: Globe,
      title: "Transport",
      description: "Secure transmission",
      color: "text-gray-500",
      completedColor: "#8b5cf6" // Purple for transport
    },
    {
      icon: Cpu,
      title: "PQC Hybrid",
      description: "Quantum-resistant layer",
      color: "text-gray-500",
      completedColor: "#f59e0b" // Orange for quantum computing
    },
    {
      icon: AlertTriangle,
      title: "Complete",
      description: "Message secured",
      color: "text-gray-500",
      completedColor: "#ef4444" // Red for completion/security
    }
  ]

  return (
    <div className="mb-8">
      <h3 className="text-xl font-semibold mb-4 text-dark-900 dark:text-white">
        Real-Time Encryption Pipeline
      </h3>
      
      <div className="relative">
        {/* Progress Line */}
        <div className="absolute top-6 left-1/2 right-1/2 h-1 bg-gray-200 dark:bg-gray-700 rounded-full transform -translate-x-1/2" style={{ width: 'calc(100% - 3rem)' }}>
          <MotionDiv
            initial={{ width: 0 }}
            animate={{ width: isEncrypting ? `${(currentStep / (steps.length - 1)) * 100}%` : currentStep >= steps.length - 1 ? "100%" : currentStep > 0 ? `${(currentStep / (steps.length - 1)) * 100}%` : "0%" }}
            transition={{ duration: 0.5, ease: "easeInOut" }}
            className="h-full bg-gradient-to-r from-indigo-500 to-blue-500 rounded-full"
          />
        </div>
        
        {/* Steps */}
        <div className="flex justify-between items-start">
          {steps.map((step, index) => (
            <div key={step.title} className="flex flex-col items-center text-center flex-1">
              <MotionDiv
                initial={{ scale: 0.8, opacity: 0 }}
                animate={{ 
                  scale: 1, 
                  opacity: 1,
                  backgroundColor: index <= currentStep && currentStep > 0 ? step.completedColor : "#6b7280"
                }}
                transition={{ duration: 0.5, delay: index * 0.2 }}
                className="w-12 h-12 rounded-full flex items-center justify-center mb-2 text-white relative z-10"
              >
                <step.icon className="w-5 h-5" />
              </MotionDiv>
              
              <h4 className="text-sm font-medium mb-1 text-gray-500">
                {step.title}
              </h4>
              
              <p className="text-xs text-gray-500 dark:text-gray-400 max-w-20">
                {step.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}