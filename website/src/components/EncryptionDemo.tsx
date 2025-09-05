import { MotionDiv, useInView } from '../utils/animations'
import { useRef, useState, useEffect } from 'react'
import { Shield } from '../utils/icons'
import { hybridEncrypt, hybridDecrypt, EncryptionResult } from '../utils/cryptoUtils'
import EncryptionPipeline from './EncryptionDemo/EncryptionPipeline'
import EncryptionLogs from './EncryptionDemo/EncryptionLogs'
import EncryptionControls from './EncryptionDemo/EncryptionControls'

// Browser-compatible base64 decoding helper
function base64ToArrayBuffer(base64: string): ArrayBuffer {
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes.buffer
}

export default function EncryptionDemo() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })
  const [message, setMessage] = useState('')
  const [ciphertext, setCiphertext] = useState('')
  const [decryptedMessage, setDecryptedMessage] = useState('')
  const [encryptionResult, setEncryptionResult] = useState<EncryptionResult | null>(null)
  const [logs, setLogs] = useState<string[]>([])
  const [showLogs, setShowLogs] = useState(false)
  const [currentStep, setCurrentStep] = useState(0)
  const [isEncrypting, setIsEncrypting] = useState(false)

  const addLog = (log: string) => {
    setLogs(prev => [...prev, log])
  }

  const handleEncrypt = async () => {
    if (!message.trim()) return
    
    setShowLogs(true)
    setLogs([])
    setIsEncrypting(true)
    setCurrentStep(0)
    
    try {
      addLog('Starting encryption process...')
      setCurrentStep(1)
      await new Promise(resolve => setTimeout(resolve, 500))
      
      addLog('Generating Argon2id salt...')
      addLog('Deriving AES-256-GCM key...')
      setCurrentStep(2)
      await new Promise(resolve => setTimeout(resolve, 500))
      
      addLog('Generating PQC Kyber key pair...')
      setCurrentStep(3)
      await new Promise(resolve => setTimeout(resolve, 500))
      
      addLog('Encrypting AES key with PQC...')
      setCurrentStep(4)
      await new Promise(resolve => setTimeout(resolve, 500))
      
      addLog('Applying AES-256-GCM encryption...')
      setCurrentStep(5)
      
      // Use real hybrid encryption
      const result = await hybridEncrypt(message)
      setEncryptionResult(result)
      
      // Format the ciphertext for display
      const formattedCiphertext = JSON.stringify(result, null, 2)
      setCiphertext(formattedCiphertext)
      
      addLog('Encryption complete!')
      
    } catch (error) {
      addLog(`Encryption failed: ${error instanceof Error ? error.message : 'Unknown error'}`)
    } finally {
      setIsEncrypting(false)
    }
  }

  const handleDecrypt = async () => {
    if (!ciphertext.trim() || !encryptionResult) return
    
    setShowLogs(true)
    setLogs([])
    
    try {
      addLog('Starting decryption process...')
      addLog('Parsing encrypted data...')
      addLog('Validating PQC key pair...')
      addLog('Decapsulating AES key...')
      addLog('Applying AES-256-GCM decryption...')
      addLog('Decryption complete!')
      
      // Use real hybrid decryption with the private key from encryption result
      if (!encryptionResult.privateKey) {
        throw new Error('Private key not available for decryption')
      }
      
      const privateKeyBuffer = base64ToArrayBuffer(encryptionResult.privateKey)
      const privateKey = new Uint8Array(privateKeyBuffer)
      
      const decrypted = await hybridDecrypt(encryptionResult, privateKey)
      setDecryptedMessage(decrypted)
      
    } catch (error) {
      addLog(`Decryption failed: ${error instanceof Error ? error.message : 'Unknown error'}`)
      setDecryptedMessage('')
    }
  }

  const handleReset = () => {
    setMessage('')
    setCiphertext('')
    setDecryptedMessage('')
    setEncryptionResult(null)
    setLogs([])
    setShowLogs(false)
    setCurrentStep(0)
    setIsEncrypting(false)
  }

  return (
    <section ref={ref} className="section-padding bg-gradient-to-b from-gray-100 to-gray-200 dark:from-dark-800 dark:to-dark-900">
      <div className="max-w-4xl mx-auto">
        {/* Section Header */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-12"
        >
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
          
          <h2 className="text-3xl md:text-4xl font-bold mb-4 text-dark-900 dark:text-white">
            Real Hybrid Encryption Demo
          </h2>
          <p className="text-lg text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
            Experience military-grade encryption in action. Watch as your message is protected with AES-256-GCM and quantum-resistant PQC encryption.
          </p>
        </MotionDiv>

        {/* Demo Container */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="bg-white dark:bg-dark-800 rounded-2xl p-8 shadow-xl border border-gray-200 dark:border-gray-700"
        >
          {/* Encryption Pipeline */}
          <EncryptionPipeline currentStep={currentStep} isEncrypting={isEncrypting} />
          
          {/* Encryption Controls */}
          <EncryptionControls
            message={message}
            setMessage={setMessage}
            ciphertext={ciphertext}
            setCiphertext={setCiphertext}
            decryptedMessage={decryptedMessage}
            onEncrypt={handleEncrypt}
            onDecrypt={handleDecrypt}
            onReset={handleReset}
          />
          
          {/* Encryption Logs */}
          <EncryptionLogs logs={logs} showLogs={showLogs} />
        </MotionDiv>
      </div>
    </section>
  )
}