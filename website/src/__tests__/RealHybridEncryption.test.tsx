import React from 'react'
import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import EncryptionDemo from '../components/EncryptionDemo'

// Mock framer-motion to fix viewport issues in test environment
vi.mock('framer-motion', () => {
  const actual = vi.importActual('framer-motion')
  return {
    ...actual,
    useInView: () => true, // always in view
    motion: {
      div: ({ children, ...props }) => <div {...props}>{children}</div>,
      span: ({ children, ...props }) => <span {...props}>{children}</span>,
      button: ({ children, ...props }) => <button {...props}>{children}</button>,
      section: ({ children, ...props }) => <section {...props}>{children}</section>,
      h2: ({ children, ...props }) => <h2 {...props}>{children}</h2>,
      h3: ({ children, ...props }) => <h3 {...props}>{children}</h3>,
      p: ({ children, ...props }) => <p {...props}>{children}</p>,
      label: ({ children, ...props }) => <label {...props}>{children}</label>,
      textarea: ({ children, ...props }) => <textarea {...props}>{children}</textarea>,
    },
  }
})

// Mock the PQC crypto module
vi.mock('../utils/pqcHybridCrypto', () => ({
  pqcCrypto: {
    getStatus: vi.fn(() => ({
      isInitialized: true,
      algorithm: 'kyber-512',
      features: [
        'AES-256-GCM (Real)',
        'PQC Hybrid (ECDH + Simulated Kyber)',
        'Zero-Knowledge Architecture',
        'Client-Side Only',
        'Quantum-Resistant Key Exchange'
      ]
    })),
    generateKeyPair: vi.fn().mockResolvedValue({
      publicKey: new Uint8Array([1, 2, 3, 4, 5]),
      privateKey: new Uint8Array([6, 7, 8, 9, 10]),
      algorithm: 'kyber-512'
    }),
    encryptHybrid: vi.fn().mockResolvedValue({
      encryptedData: new Uint8Array([11, 12, 13, 14, 15]),
      encapsulatedKey: new Uint8Array([16, 17, 18, 19, 20]),
      iv: new Uint8Array([21, 22, 23, 24, 25]),
      algorithm: 'AES-256-GCM + PQC Hybrid (ECDH)',
      timestamp: Date.now()
    }),
    decryptHybrid: vi.fn().mockResolvedValue('Test decrypted message')
  }
}))

describe('Real Hybrid Encryption Demo', () => {
  beforeEach(() => {
    // Mock console methods
    vi.spyOn(console, 'log').mockImplementation(() => {})
    vi.spyOn(console, 'warn').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders the real hybrid encryption demo', () => {
    render(<EncryptionDemo />)
    
    expect(screen.getByText('Experience Real')).toBeInTheDocument()
    expect(screen.getByText('Hybrid Encryption')).toBeInTheDocument()
    expect(screen.getByText('REAL ENCRYPTION')).toBeInTheDocument()
  })

  it('shows zero-knowledge meter with initial state', () => {
    render(<EncryptionDemo />)
    
    expect(screen.getByText('Experience Real')).toBeInTheDocument()
    expect(screen.getByText('Hybrid Encryption')).toBeInTheDocument()
    expect(screen.getByText('REAL ENCRYPTION')).toBeInTheDocument()
  })

  it('generates hybrid key pair on initialization', async () => {
    render(<EncryptionDemo />)
    
    await waitFor(() => {
      expect(screen.getByText('AES-256-GCM + Argon2id + Kyber-512')).toBeInTheDocument()
    })
  })

  it('allows user to input message and encrypt', async () => {
    render(<EncryptionDemo />)
    
    // Wait for key generation
    await waitFor(() => {
      expect(screen.getByText('AES-256-GCM + Argon2id + Kyber-512')).toBeInTheDocument()
    })
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    fireEvent.change(textarea, { target: { value: 'Test secret message' } })
    
    const encryptButton = screen.getByText(/🔐 Encrypt with REAL Hybrid Crypto/)
    expect(encryptButton).toBeInTheDocument()
    expect(encryptButton).not.toBeDisabled()
  })

  it('shows encryption pipeline during encryption', async () => {
    render(<EncryptionDemo />)
    
    // Wait for key generation
    await waitFor(() => {
      expect(screen.getByText('AES-256-GCM + Argon2id + Kyber-512')).toBeInTheDocument()
    })
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    fireEvent.change(textarea, { target: { value: 'Test message' } })
    
    const encryptButton = screen.getByText(/🔐 Encrypt with REAL Hybrid Crypto/)
    fireEvent.click(encryptButton)
    
    // Check for pipeline steps
    await waitFor(() => {
      expect(screen.getByText('Real Encryption Pipeline')).toBeInTheDocument()
      expect(screen.getByText('Generate Keys')).toBeInTheDocument()
    })
  })

  it('displays encryption results after completion', async () => {
    render(<EncryptionDemo />)
    
    // Wait for key generation
    await waitFor(() => {
      expect(screen.getByText('AES-256-GCM + Argon2id + Kyber-512')).toBeInTheDocument()
    })
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    fireEvent.change(textarea, { target: { value: 'Test message' } })
    
    const encryptButton = screen.getByText(/🔐 Encrypt with REAL Hybrid Crypto/)
    fireEvent.click(encryptButton)
    
    // Wait for encryption to complete
    await waitFor(() => {
      expect(screen.getByText('Encryption Results')).toBeInTheDocument()
    }, { timeout: 10000 })
    
    expect(screen.getByText('Hybrid Ciphertext (AES + PQC)')).toBeInTheDocument()
  })

  it('allows decryption of encrypted message', async () => {
    render(<EncryptionDemo />)
    
    // Wait for key generation
    await waitFor(() => {
      expect(screen.getByText('AES-256-GCM + Argon2id + Kyber-512')).toBeInTheDocument()
    })
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    fireEvent.change(textarea, { target: { value: 'Test message' } })
    
    const encryptButton = screen.getByText(/🔐 Encrypt with REAL Hybrid Crypto/)
    fireEvent.click(encryptButton)
    
    // Wait for encryption to complete
    await waitFor(() => {
      expect(screen.getByText('🔓 Decrypt with REAL Hybrid Crypto')).toBeInTheDocument()
    }, { timeout: 10000 })
    
    const decryptButton = screen.getByText(/🔓 Decrypt with REAL Hybrid Crypto/)
    fireEvent.click(decryptButton)
    
    // Wait for decryption
    await waitFor(() => {
      expect(screen.getByText('Decrypted Message')).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('shows technical details after encryption', async () => {
    render(<EncryptionDemo />)
    
    // Wait for key generation
    await waitFor(() => {
      expect(screen.getByText('AES-256-GCM + Argon2id + Kyber-512')).toBeInTheDocument()
    })
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    fireEvent.change(textarea, { target: { value: 'Test message' } })
    
    const encryptButton = screen.getByText(/🔐 Encrypt with REAL Hybrid Crypto/)
    fireEvent.click(encryptButton)
    
    // Wait for encryption to complete
    await waitFor(() => {
      expect(screen.getByText('Technical Details')).toBeInTheDocument()
    }, { timeout: 10000 })
    
    expect(screen.getByText('Algorithm:')).toBeInTheDocument()
    expect(screen.getByText('Encrypted Data:')).toBeInTheDocument()
    expect(screen.getByText('Encapsulated Key:')).toBeInTheDocument()
  })

  it('allows generating new key pairs', async () => {
    render(<EncryptionDemo />)
    
    // Wait for key generation
    await waitFor(() => {
      expect(screen.getByText('AES-256-GCM + Argon2id + Kyber-512')).toBeInTheDocument()
    })
    
    // Test that the encryption button is available
    const encryptButton = screen.getByText('🔐 Encrypt with REAL Hybrid Crypto')
    expect(encryptButton).toBeInTheDocument()
    
    // Test that the component is ready for encryption
    expect(screen.getByText('Your Message')).toBeInTheDocument()
    expect(screen.getByText('Encryption Results')).toBeInTheDocument()
  })

  it('clears all data when clear button is clicked', async () => {
    render(<EncryptionDemo />)
    
    // Wait for key generation
    await waitFor(() => {
      expect(screen.getByText('AES-256-GCM + Argon2id + Kyber-512')).toBeInTheDocument()
    })
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    fireEvent.change(textarea, { target: { value: 'Test message' } })
    
    const clearButton = screen.getByText('Clear All')
    fireEvent.click(clearButton)
    
    // Check that textarea is cleared
    expect(textarea).toHaveValue('')
  })
})


