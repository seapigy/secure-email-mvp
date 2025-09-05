import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import EncryptionDemo from '../components/EncryptionDemo'

// Mock the crypto utils to avoid WASM loading issues in test environment
vi.mock('../utils/cryptoUtils', () => ({
  hybridEncrypt: vi.fn().mockResolvedValue({
    encryptedMessage: 'mock-ciphertext',
    encryptedAESKey: 'mock-encrypted-key',
    iv: 'mock-iv',
    salt: 'mock-salt',
    publicKey: 'mock-public-key',
    privateKey: 'mock-private-key',
    algorithm: 'AES-256-GCM + Argon2id + Kyber-512',
    timestamp: Date.now()
  }),
  hybridDecrypt: vi.fn().mockResolvedValue('Hello, World!')
}))

describe('EncryptionDemo Integration', () => {
  beforeEach(() => {
    // Clear console logs for cleaner test output
    console.debug = () => {}
  })

  it('should render the encryption demo component', () => {
    render(<EncryptionDemo />)
    
    expect(screen.getByText('Real Hybrid Encryption Demo')).toBeInTheDocument()
    expect(screen.getByText(/Experience military-grade encryption in action/)).toBeInTheDocument()
  })

  it('should show encryption controls', () => {
    render(<EncryptionDemo />)
    
    expect(screen.getByPlaceholderText('Type your secret message here...')).toBeInTheDocument()
    expect(screen.getByText('Encrypt Message')).toBeInTheDocument()
    expect(screen.getByText('Decrypt Message')).toBeInTheDocument()
    expect(screen.getByText('Reset')).toBeInTheDocument()
  })

  it('should show real encryption logs section', () => {
    render(<EncryptionDemo />)
    
    expect(screen.getByText('Real Encryption Logs')).toBeInTheDocument()
  })

  it('should enable encrypt button when message is entered', async () => {
    render(<EncryptionDemo />)
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    const encryptButton = screen.getByText('Encrypt Message')
    
    expect(encryptButton).toBeDisabled()
    
    fireEvent.change(textarea, { target: { value: 'Hello, World!' } })
    
    await waitFor(() => {
      expect(encryptButton).not.toBeDisabled()
    })
  })

  it('should show encrypt button as disabled when no message', () => {
    render(<EncryptionDemo />)
    
    const encryptButton = screen.getByText('Encrypt Message')
    expect(encryptButton).toBeDisabled()
  })

  it('should show decrypt button as disabled when no ciphertext', () => {
    render(<EncryptionDemo />)
    
    const decryptButton = screen.getByText('Decrypt Message')
    expect(decryptButton).toBeDisabled()
  })

  it('should handle encryption flow', async () => {
    render(<EncryptionDemo />)
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    const encryptButton = screen.getByText('Encrypt Message')
    
    fireEvent.change(textarea, { target: { value: 'Hello, World!' } })
    fireEvent.click(encryptButton)
    
    await waitFor(() => {
      expect(screen.getByText('Starting encryption process...')).toBeInTheDocument()
    })
    
    await waitFor(() => {
      expect(screen.getByText('Encryption complete!')).toBeInTheDocument()
    })
  })

  it('should handle decryption flow', async () => {
    render(<EncryptionDemo />)
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    const encryptButton = screen.getByText('Encrypt Message')
    const decryptButton = screen.getByText('Decrypt Message')
    
    // First encrypt
    fireEvent.change(textarea, { target: { value: 'Hello, World!' } })
    fireEvent.click(encryptButton)
    
    await waitFor(() => {
      expect(screen.getByText('Encryption complete!')).toBeInTheDocument()
    })
    
    // Then decrypt
    fireEvent.click(decryptButton)
    
    await waitFor(() => {
      expect(screen.getByText('Starting decryption process...')).toBeInTheDocument()
    })
    
    await waitFor(() => {
      expect(screen.getByText('Decryption complete!')).toBeInTheDocument()
    })
  })

  it('should handle reset functionality', async () => {
    render(<EncryptionDemo />)
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    const encryptButton = screen.getByText('Encrypt Message')
    const resetButton = screen.getByText('Reset')
    
    // Enter message and encrypt
    fireEvent.change(textarea, { target: { value: 'Hello, World!' } })
    fireEvent.click(encryptButton)
    
    await waitFor(() => {
      expect(screen.getByText('Encryption complete!')).toBeInTheDocument()
    })
    
    // Reset
    fireEvent.click(resetButton)
    
    await waitFor(() => {
      expect(textarea).toHaveValue('')
    })
  })

  it('should show encryption pipeline visualization during encryption', async () => {
    render(<EncryptionDemo />)
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    const encryptButton = screen.getByText('Encrypt Message')
    
    fireEvent.change(textarea, { target: { value: 'Hello, World!' } })
    fireEvent.click(encryptButton)
    
    await waitFor(() => {
      expect(screen.getByText('Real-Time Encryption Pipeline')).toBeInTheDocument()
    })
  })

  it('should display marketing overlay during encryption', async () => {
    render(<EncryptionDemo />)
    
    const textarea = screen.getByPlaceholderText('Type your secret message here...')
    const encryptButton = screen.getByText('Encrypt Message')
    
    fireEvent.change(textarea, { target: { value: 'Hello, World!' } })
    fireEvent.click(encryptButton)
    
    await waitFor(() => {
      expect(screen.getByText(/Watch as your message is protected with AES-256-GCM and quantum-resistant PQC encryption/)).toBeInTheDocument()
    })
  })
})
