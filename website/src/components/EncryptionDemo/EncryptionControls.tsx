import { MotionDiv } from '../../utils/animations'
import { Copy, Check, RefreshCw } from '../../utils/icons'
import { useState } from 'react'

interface EncryptionControlsProps {
  message: string
  setMessage: (message: string) => void
  ciphertext: string
  setCiphertext: (ciphertext: string) => void
  decryptedMessage: string
  onEncrypt: () => void
  onDecrypt: () => void
  onReset: () => void
}

export default function EncryptionControls({
  message,
  setMessage,
  ciphertext,
  setCiphertext,
  decryptedMessage,
  onEncrypt,
  onDecrypt,
  onReset
}: EncryptionControlsProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (ciphertext) {
      await navigator.clipboard.writeText(ciphertext)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="space-y-6">
      {/* Message Input */}
      <div>
        <label className="block text-sm font-medium text-dark-900 dark:text-white mb-2">
          Enter your message:
        </label>
        <textarea
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Type your secret message here..."
          className="w-full h-32 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-dark-800 text-dark-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
        />
      </div>

      {/* Ciphertext Output */}
      {ciphertext && (
        <div>
          <label className="block text-sm font-medium text-dark-900 dark:text-white mb-2">
            Encrypted message:
          </label>
          <div className="relative">
            <textarea
              value={ciphertext}
              onChange={(e) => setCiphertext(e.target.value)}
              className="w-full h-32 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-dark-700 text-dark-900 dark:text-white font-mono text-sm resize-none"
              placeholder="Encrypted message will appear here..."
            />
            <MotionDiv
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={handleCopy}
              className="absolute top-2 right-2 p-2 bg-indigo-500 text-white rounded-lg cursor-pointer hover:bg-indigo-600 transition-colors"
            >
              {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
            </MotionDiv>
          </div>
        </div>
      )}

      {/* Decrypted Message Output */}
      {decryptedMessage && (
        <div>
          <label className="block text-sm font-medium text-dark-900 dark:text-white mb-2">
            Decrypted message:
          </label>
          <textarea
            value={decryptedMessage}
            readOnly
            className="w-full h-32 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-green-50 dark:bg-green-900/20 text-dark-900 dark:text-white font-mono text-sm resize-none"
            placeholder="Decrypted message will appear here..."
          />
        </div>
      )}

      {/* Control Buttons */}
      <div className="flex flex-wrap gap-4">
        <MotionDiv
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          onClick={onEncrypt}
          disabled={!message.trim()}
          className="flex-1 min-w-32 px-6 py-3 bg-gradient-to-r from-indigo-500 to-blue-500 text-white rounded-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:from-indigo-600 hover:to-blue-600 transition-all duration-200 cursor-pointer text-center"
          as="button"
        >
          Encrypt Message
        </MotionDiv>
        
        <MotionDiv
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          onClick={onDecrypt}
          disabled={!ciphertext.trim()}
          className="flex-1 min-w-32 px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:from-blue-600 hover:to-indigo-600 transition-all duration-200 cursor-pointer text-center"
          as="button"
        >
          Decrypt Message
        </MotionDiv>
        
        <MotionDiv
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          onClick={onReset}
          className="px-6 py-3 bg-gray-500 text-white rounded-lg font-medium hover:bg-gray-600 transition-colors duration-200 cursor-pointer flex items-center space-x-2"
          as="button"
        >
          <RefreshCw className="w-4 h-4" />
          <span>Reset</span>
        </MotionDiv>
      </div>
    </div>
  )
}