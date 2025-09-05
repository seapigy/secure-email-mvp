import React from 'react'
import { MotionDiv } from '../../utils/animations'

interface EncryptionLogsProps {
  logs: string[]
  showLogs: boolean
}

export default function EncryptionLogs({ logs, showLogs }: EncryptionLogsProps) {
  return (
    <div className="mt-8">
      <h3 className="text-xl font-semibold mb-4 text-dark-900 dark:text-white">
        Real Encryption Logs
      </h3>
      
      <div className="bg-dark-900 dark:bg-dark-800 rounded-xl p-4 max-h-64 overflow-y-auto">
        <div className="space-y-2">
          {logs.map((log, index) => (
            <MotionDiv
              key={index}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.3, delay: index * 0.05 }}
              className="text-sm font-mono text-green-500"
            >
              {log}
            </MotionDiv>
          ))}
        </div>
        
        {logs.length === 0 && (
          <div className="text-gray-500 text-sm font-mono">
            Waiting for encryption operations...
          </div>
        )}
      </div>
    </div>
  )
}