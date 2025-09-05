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

// Mock the crypto utils
vi.mock('../utils/cryptoUtils', () => ({
  hybridEncrypt: vi.fn().mockResolvedValue({
    ciphertext: 'encrypted-data-here',
    iv: new Uint8Array([1, 2, 3, 4]),
    salt: new Uint8Array([5, 6, 7, 8])
  }),
  hybridDecrypt: vi.fn().mockResolvedValue('decrypted message'),
  generatePQCKeyPair: vi.fn().mockResolvedValue({
    publicKey: new Uint8Array([1, 2, 3, 4, 5]),
    privateKey: new Uint8Array([6, 7, 8, 9, 10])
  })
}))

describe('Real Hybrid Encryption Demo', () => {
  beforeEach(() => {
    // Mock console.log to avoid noise in tests
    vi.spyOn(console, 'log').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders the encryption demo component', () => {
    render(<EncryptionDemo />)
    
    expect(screen.getByText('Real Hybrid Encryption Demo')).toBeInTheDocument()
    expect(screen.getByText(/Experience military-grade encryption in action/)).toBeInTheDocument()
  })

  it('displays encryption pipeline steps', () => {
    render(<EncryptionDemo />)
    
    expect(screen.getByText('Real-Time Encryption Pipeline')).toBeInTheDocument()
    expect(screen.getByText('Key Derivation')).toBeInTheDocument()
    expect(screen.getByText('AES-256-GCM')).toBeInTheDocument()
    expect(screen.getByText('Transport')).toBeInTheDocument()
    expect(screen.getByText('PQC Hybrid')).toBeInTheDocument()
    expect(screen.getByText('Complete')).toBeInTheDocument()
  })

  it('shows encryption controls', () => {
    render(<EncryptionDemo />)
    
    expect(screen.getByText('Enter your message:')).toBeInTheDocument()
    expect(screen.getByText('Encrypt Message')).toBeInTheDocument()
    expect(screen.getByText('Decrypt Message')).toBeInTheDocument()
    expect(screen.getByText('Reset')).toBeInTheDocument()
  })

  it('displays real encryption logs section', () => {
    render(<EncryptionDemo />)
    
    expect(screen.getByText('Real Encryption Logs')).toBeInTheDocument()
    // The logs section exists but may not show the waiting text if showLogs is false
  })

  it('allows user to enter a message', () => {
    render(<EncryptionDemo />)
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    expect(textarea).toBeInTheDocument()
    
    fireEvent.change(textarea, { target: { value: 'Test message' } })
    expect(textarea.value).toBe('Test message')
  })

  it('shows encrypt button is disabled when no message', () => {
    render(<EncryptionDemo />)
    
    const encryptButton = screen.getByText('Encrypt Message')
    expect(encryptButton).toBeDisabled()
  })

  it('enables encrypt button when message is entered', () => {
    render(<EncryptionDemo />)
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    const encryptButton = screen.getByText('Encrypt Message')
    
    fireEvent.change(textarea, { target: { value: 'Test message' } })
    expect(encryptButton).not.toBeDisabled()
  })

  it('shows decrypt button is disabled when no ciphertext', () => {
    render(<EncryptionDemo />)
    
    const decryptButton = screen.getByText('Decrypt Message')
    expect(decryptButton).toBeDisabled()
  })

  it('allows reset functionality', () => {
    render(<EncryptionDemo />)
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    const resetButton = screen.getByText('Reset')
    
    fireEvent.change(textarea, { target: { value: 'Test message' } })
    expect(textarea.value).toBe('Test message')
    
    fireEvent.click(resetButton)
    expect(textarea.value).toBe('')
  })
})