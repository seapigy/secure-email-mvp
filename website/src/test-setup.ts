import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock window.matchMedia for theme detection
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock IntersectionObserver for useInView
global.IntersectionObserver = vi.fn().mockImplementation((callback) => ({
  observe: vi.fn((element) => {
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

// Enhanced crypto mocks for tests
let mockCounter = 0;
let keyCounter = 0;

// Store state for mocking
let mockState = {
  lastMessage: '',
  lastKey: null,
  callCount: 0,
  keyCounter: 0,
  keyMap: new Map(),
  encryptionMap: new Map(), // Store encryption results for consistent decryption
  keyMap: new Map() // Store keys by input for consistency
}

const mockCrypto = {
  getRandomValues: vi.fn((arr) => {
    // Generate different random values each time to ensure different ciphertexts
    for (let i = 0; i < arr.length; i++) {
      arr[i] = (mockCounter + i) % 256
    }
    mockCounter++
    return arr
  }),
  subtle: {
    encrypt: vi.fn().mockImplementation(async (algorithm, key, data) => {
      // Store the encryption for later decryption
      const keyId = JSON.stringify(key)
      const dataStr = new TextDecoder().decode(data)
      const ciphertext = `encrypted_${dataStr}_${mockCounter}`
      
      mockState.encryptionMap.set(keyId, {
        ciphertext,
        originalData: dataStr
      })
      
      // Return different ciphertexts each time
      const buffer = new ArrayBuffer(16)
      const view = new Uint8Array(buffer)
      for (let i = 0; i < 16; i++) {
        view[i] = (mockCounter + i) % 256
      }
      mockCounter++
      return buffer
    }),
    decrypt: vi.fn().mockImplementation(async (algorithm, key, data) => {
      const keyId = JSON.stringify(key)
      const stored = mockState.encryptionMap.get(keyId)
      
      if (!stored) {
        throw new Error('Decryption failed with wrong key')
      }
      
      return new TextEncoder().encode(stored.originalData)
    }),
    importKey: vi.fn().mockImplementation(async (format, keyData, algorithm, extractable, keyUsages) => {
      // Create a consistent key based on the input data
      const keyId = JSON.stringify(keyData)
      if (!mockState.keyMap.has(keyId)) {
        mockState.keyMap.set(keyId, {
          type: 'secret',
          algorithm,
          extractable,
          usages: keyUsages,
          id: keyId
        })
      }
      
      return mockState.keyMap.get(keyId)
    }),
    exportKey: vi.fn().mockImplementation(async (format, key) => {
      // Return consistent keys based on the key object
      const keyId = key.id || JSON.stringify(key)
      
      // Generate consistent key data based on the key ID
      const buffer = new ArrayBuffer(32)
      const view = new Uint8Array(buffer)
      const hash = keyId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
      
      for (let i = 0; i < 32; i++) {
        view[i] = (hash + i) % 256
      }
      
      return buffer
    }),
    digest: vi.fn().mockImplementation(async (algorithm, data) => {
      // Return consistent hashes based on input data
      const dataStr = new TextDecoder().decode(data)
      const hash = dataStr.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
      
      const buffer = new ArrayBuffer(32)
      const view = new Uint8Array(buffer)
      for (let i = 0; i < 32; i++) {
        view[i] = (hash + i) % 256
      }
      return buffer
    }),
  }
}

Object.defineProperty(global, 'crypto', {
  value: mockCrypto,
  writable: true
})
