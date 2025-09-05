import { motion } from 'framer-motion';
import { Eye, EyeOff, Shield, Lock, Zap } from 'lucide-react';

interface ZeroKnowledgeMeterProps {
  isActive: boolean;
  encryptionLevel: 'none' | 'aes' | 'hybrid' | 'complete';
  serverVisibility: number; // 0-100, where 0 is completely private
}

export default function ZeroKnowledgeMeter({ 
  isActive, 
  encryptionLevel, 
  serverVisibility 
}: ZeroKnowledgeMeterProps) {
  const getVisibilityColor = (level: number) => {
    if (level === 0) return 'text-secure-400';
    if (level < 30) return 'text-yellow-400';
    if (level < 70) return 'text-orange-400';
    return 'text-red-400';
  };

  const getVisibilityIcon = (level: number) => {
    if (level === 0) return <EyeOff className="w-6 h-6" />;
    if (level < 30) return <Shield className="w-6 h-6" />;
    if (level < 70) return <Lock className="w-6 h-6" />;
    return <Eye className="w-6 h-6" />;
  };

  const getEncryptionDescription = (level: string) => {
    switch (level) {
      case 'none':
        return 'No encryption - Server sees everything';
      case 'aes':
        return 'AES-256-GCM - Server sees encrypted data only';
      case 'hybrid':
        return 'Hybrid AES + PQC - Server sees nothing';
      case 'complete':
        return 'Zero-Knowledge - Complete privacy guaranteed';
      default:
        return 'Encryption status unknown';
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: isActive ? 1 : 0.7, scale: isActive ? 1 : 0.95 }}
      transition={{ duration: 0.5 }}
      className="bg-gradient-to-r from-dark-800 to-dark-900 rounded-2xl p-6 border border-secure-400/30 shadow-2xl"
    >
      <div className="text-center mb-6">
        <motion.div
          animate={{ 
            scale: isActive ? [1, 1.05, 1] : 1,
            rotate: isActive ? [0, 5, -5, 0] : 0
          }}
          transition={{ 
            duration: 2, 
            repeat: isActive ? Infinity : 0,
            ease: "easeInOut"
          }}
          className="w-16 h-16 bg-gradient-to-br from-secure-500 to-primary-500 rounded-full flex items-center justify-center mx-auto mb-4"
        >
          <Zap className="w-8 h-8 text-white" />
        </motion.div>
        
        <h3 className="text-2xl font-bold text-white mb-2">
          Zero-Knowledge Meter
        </h3>
        <p className="text-gray-300 text-sm">
          Real-time privacy protection visualization
        </p>
      </div>

      {/* Encryption Level Indicator */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-3">
          <span className="text-white font-medium">Encryption Level:</span>
          <span className={`font-bold ${getVisibilityColor(serverVisibility)}`}>
            {encryptionLevel.toUpperCase()}
          </span>
        </div>
        
        <div className="bg-dark-700 rounded-full h-3 overflow-hidden">
          <motion.div
            initial={{ width: 0 }}
            animate={{ width: `${(encryptionLevel === 'none' ? 0 : encryptionLevel === 'aes' ? 25 : encryptionLevel === 'hybrid' ? 75 : 100)}%` }}
            transition={{ duration: 1, ease: "easeOut" }}
            className="h-full bg-gradient-to-r from-secure-500 to-primary-500"
          />
        </div>
        
        <p className="text-gray-400 text-xs mt-2 text-center">
          {getEncryptionDescription(encryptionLevel)}
        </p>
      </div>

      {/* Server Visibility Meter */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-3">
          <span className="text-white font-medium">Server Visibility:</span>
          <div className="flex items-center space-x-2">
            {getVisibilityIcon(serverVisibility)}
            <span className={`font-bold ${getVisibilityColor(serverVisibility)}`}>
              {serverVisibility}%
            </span>
          </div>
        </div>
        
        <div className="bg-dark-700 rounded-full h-3 overflow-hidden">
          <motion.div
            initial={{ width: 0 }}
            animate={{ width: `${serverVisibility}%` }}
            transition={{ duration: 1, ease: "easeOut" }}
            className={`h-full ${serverVisibility === 0 ? 'bg-secure-500' : serverVisibility < 30 ? 'bg-yellow-500' : serverVisibility < 70 ? 'bg-orange-500' : 'bg-red-500'}`}
          />
        </div>
        
        <div className="flex justify-between text-xs text-gray-400 mt-1">
          <span>Private</span>
          <span>Exposed</span>
        </div>
      </div>

      {/* Privacy Features */}
      <div className="space-y-3">
        <div className="flex items-center space-x-3">
          <div className={`w-3 h-3 rounded-full ${encryptionLevel !== 'none' ? 'bg-secure-400' : 'bg-gray-600'}`} />
          <span className={`text-sm ${encryptionLevel !== 'none' ? 'text-white' : 'text-gray-500'}`}>
            AES-256-GCM Encryption
          </span>
        </div>
        
        <div className="flex items-center space-x-3">
          <div className={`w-3 h-3 rounded-full ${encryptionLevel === 'hybrid' || encryptionLevel === 'complete' ? 'bg-secure-400' : 'bg-gray-600'}`} />
          <span className={`text-sm ${encryptionLevel === 'hybrid' || encryptionLevel === 'complete' ? 'text-white' : 'text-gray-500'}`}>
            PQC Hybrid Key Exchange
          </span>
        </div>
        
        <div className="flex items-center space-x-3">
          <div className={`w-3 h-3 rounded-full ${encryptionLevel === 'complete' ? 'bg-secure-400' : 'bg-gray-600'}`} />
          <span className={`text-sm ${encryptionLevel === 'complete' ? 'text-white' : 'text-gray-500'}`}>
            Zero-Knowledge Architecture
          </span>
        </div>
        
        <div className="flex items-center space-x-3">
          <div className={`w-3 h-3 rounded-full ${encryptionLevel !== 'none' ? 'bg-secure-400' : 'bg-gray-600'}`} />
          <span className={`text-sm ${encryptionLevel !== 'none' ? 'text-white' : 'text-gray-500'}`}>
            Client-Side Only
          </span>
        </div>
      </div>

      {/* Status Message */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5 }}
        className="mt-6 p-4 bg-dark-700 rounded-xl border-l-4 border-secure-400"
      >
        <div className="flex items-start space-x-3">
          <Shield className="w-5 h-5 text-secure-400 mt-0.5 flex-shrink-0" />
          <div>
            <p className="text-white font-medium text-sm">
              {serverVisibility === 0 ? 'Complete Privacy Achieved' : 'Privacy Protection Active'}
            </p>
            <p className="text-gray-400 text-xs mt-1">
              {serverVisibility === 0 
                ? 'Your data is completely invisible to servers. Zero-knowledge encryption active.'
                : 'Your data is protected with multiple encryption layers. Server visibility minimized.'
              }
            </p>
          </div>
        </div>
      </motion.div>
    </motion.div>
  );
}


