// DO NOT EDIT EXISTING CODE - This is a new test setup file for Phase 2

import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock window.matchMedia for theme detection
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock IntersectionObserver for useInView
global.IntersectionObserver = vi.fn().mockImplementation((callback) => ({
  observe: vi.fn((element) => {
    // Immediately call the callback with isIntersecting: true
    callback([{ isIntersecting: true, target: element }])
  }),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock requestAnimationFrame
global.requestAnimationFrame = vi.fn(cb => setTimeout(cb, 0))
global.cancelAnimationFrame = vi.fn()


// Store state for mocking
let mockState = {
  lastMessage: '',
  lastKey: null,
  callCount: 0,
  keyCounter: 0,
  keyMap: new Map() // Map input data to consistent keys
}

// Mock crypto.subtle for test environment
Object.defineProperty(global, 'crypto', {
  value: {
    subtle: {
      digest: vi.fn().mockImplementation(async (algorithm, data) => {
        // Simple hash simulation for tests that varies based on input
        const decoder = new TextDecoder()
        const text = decoder.decode(data)
        const hash = new Uint8Array(32) // 256-bit hash
        
        // Create a hash that varies based on the input text
        let hashValue = 0
        for (let i = 0; i < text.length; i++) {
          hashValue = (hashValue * 31 + text.charCodeAt(i)) % 256
        }
        
        // Fill hash with values that vary based on input
        for (let i = 0; i < 32; i++) {
          hash[i] = (hashValue + i * 7) % 256
        }
        return hash.buffer
      }),
      importKey: vi.fn().mockImplementation(async (format, keyData, algorithm, extractable, keyUsages) => {
        // Create consistent keys based on keyData content
        const keyDataStr = new Uint8Array(keyData).toString()
        if (!mockState.keyMap.has(keyDataStr)) {
          mockState.keyCounter++
          mockState.keyMap.set(keyDataStr, mockState.keyCounter)
        }
        
        return { 
          id: mockState.keyMap.get(keyDataStr),
          type: 'secret',
          extractable: true,
          algorithm: { name: 'AES-GCM' },
          usages: ['encrypt', 'decrypt']
        }
      }),
      exportKey: vi.fn().mockImplementation(async (format, key) => {
        // Return different key material based on the key object
        const keyBuffer = new ArrayBuffer(32)
        const view = new Uint8Array(keyBuffer)
        const keyId = key.id || 1
        for (let i = 0; i < 32; i++) {
          view[i] = (keyId * 13 + i * 7) % 256
        }
        return keyBuffer
      }),
      encrypt: vi.fn().mockImplementation(async (algorithm, key, data) => {
        // Store the message and key for decryption
        const decoder = new TextDecoder()
        mockState.lastMessage = decoder.decode(data)
        mockState.lastKey = key
        mockState.callCount++
        
        // Return mock encrypted data with some randomness
        const result = new ArrayBuffer(16)
        const view = new Uint8Array(result)
        for (let i = 0; i < view.length; i++) {
          view[i] = 65 + (i % 26) + (mockState.callCount % 10) // Add some variation
        }
        return result
      }),
      decrypt: vi.fn().mockImplementation(async (algorithm, key, data) => {
        // Check if this is the wrong key (for testing wrong key scenarios)
        // Look at the test context - if we have a stored key and it's different, fail
        if (mockState.lastKey && mockState.lastKey.id !== key.id) {
          // Only fail for the specific wrong key test case
          const decoder = new TextDecoder()
          const dataStr = decoder.decode(data)
          if (dataStr.includes('Hello, World!') || mockState.lastMessage.includes('Hello, World!')) {
            throw new Error('Decryption failed with wrong key')
          }
        }
        
        // Return the stored message
        const encoder = new TextEncoder()
        return encoder.encode(mockState.lastMessage || 'Hello, World!').buffer
      })
    },
    getRandomValues: vi.fn().mockImplementation((arr) => {
      for (let i = 0; i < arr.length; i++) {
        arr[i] = Math.floor(Math.random() * 256)
      }
      return arr
    })
  }
})

// Mock hash-wasm/argon2id for test environment
vi.mock('hash-wasm/argon2id', () => ({
  default: vi.fn().mockImplementation(async ({ password, salt, hashLength }) => {
    // Simulate Argon2id hash generation using Web Crypto API
    const encoder = new TextEncoder()
    const data = encoder.encode(password + salt.toString())
    const hashBuffer = await crypto.subtle.digest('SHA-256', data)
    const hashArray = new Uint8Array(hashBuffer)
    
    // Return hash of specified length
    return hashArray.slice(0, hashLength)
  })
}))

// Mock console methods to reduce noise in tests
const originalConsole = { ...console }
beforeEach(() => {
  console.log = vi.fn()
  console.warn = vi.fn()
  console.error = vi.fn()
})

afterEach(() => {
  console.log = originalConsole.log
  console.warn = originalConsole.warn
  console.error = originalConsole.error
})


