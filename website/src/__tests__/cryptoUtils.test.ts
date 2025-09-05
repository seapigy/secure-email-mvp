import { describe, it, expect, beforeEach, vi } from 'vitest'
import { 
  deriveAESKey, 
  aesEncrypt, 
  aesDecrypt, 
  generatePQCKeyPair, 
  pqcEncapsulate, 
  pqcDecapsulate,
  hybridEncrypt,
  hybridDecrypt
} from '../utils/cryptoUtils'

// Mock hash-wasm for test environment
vi.mock('hash-wasm', () => ({
  argon2id: vi.fn().mockImplementation(async ({ password, salt, hashLength }) => {
    // Simulate Argon2id with a simple hash for testing
    const encoder = new TextEncoder()
    const data = encoder.encode(password + salt.toString())
    const hash = await crypto.subtle.digest('SHA-256', data)
    return new Uint8Array(hash).slice(0, hashLength)
  })
}))

describe('cryptoUtils', () => {
  beforeEach(() => {
    // Clear console logs for cleaner test output
    console.debug = () => {}
  })

  describe('deriveAESKey', () => {
    it('should derive a 256-bit AES key', async () => {
      const key = await deriveAESKey('test password')
      const raw = new Uint8Array(await crypto.subtle.exportKey('raw', key))
      expect(raw.byteLength).toBe(32)
    })

    it('should derive consistent keys with same input', async () => {
      const salt = new Uint8Array(16)
      const key1 = await deriveAESKey('test', salt)
      const key2 = await deriveAESKey('test', salt)
      
      const raw1 = new Uint8Array(await crypto.subtle.exportKey('raw', key1))
      const raw2 = new Uint8Array(await crypto.subtle.exportKey('raw', key2))
      
      expect(raw1).toEqual(raw2)
    })

    it('should derive different keys with different salts', async () => {
      const salt1 = new Uint8Array(16)
      const salt2 = new Uint8Array(16)
      salt2[0] = 1 // Make salts different
      
      const key1 = await deriveAESKey('test', salt1)
      const key2 = await deriveAESKey('test', salt2)
      
      const raw1 = new Uint8Array(await crypto.subtle.exportKey('raw', key1))
      const raw2 = new Uint8Array(await crypto.subtle.exportKey('raw', key2))
      
      expect(raw1).not.toEqual(raw2)
    })
  })

  describe('AES-GCM encryption/decryption', () => {
    it('should encrypt and decrypt successfully', async () => {
      const key = await deriveAESKey('test message')
      const message = 'Hello, World!'
      
      const { iv, ciphertext } = await aesEncrypt(message, key)
      const decrypted = await aesDecrypt(ciphertext, iv, key)
      
      expect(decrypted).toBe(message)
    })

    it('should produce different ciphertexts for same message', async () => {
      const key = await deriveAESKey('test message')
      const message = 'Hello, World!'
      
      const result1 = await aesEncrypt(message, key)
      const result2 = await aesEncrypt(message, key)
      
      expect(result1.ciphertext).not.toEqual(result2.ciphertext)
      expect(result1.iv).not.toEqual(result2.iv)
    })

    it('should fail decryption with wrong key', async () => {
      const key1 = await deriveAESKey('test message')
      const key2 = await deriveAESKey('different message')
      const message = 'Hello, World!'
      
      const { iv, ciphertext } = await aesEncrypt(message, key1)
      
      await expect(aesDecrypt(ciphertext, iv, key2)).rejects.toThrow()
    })
  })

  describe('PQC key generation', () => {
    it('should generate key pairs with correct sizes', async () => {
      const keyPair = await generatePQCKeyPair()
      
      expect(keyPair.publicKey.byteLength).toBe(800) // Kyber-512 public key
      expect(keyPair.privateKey.byteLength).toBe(1632) // Kyber-512 private key
    })

    it('should generate different key pairs each time', async () => {
      const keyPair1 = await generatePQCKeyPair()
      const keyPair2 = await generatePQCKeyPair()
      
      expect(keyPair1.publicKey).not.toEqual(keyPair2.publicKey)
      expect(keyPair1.privateKey).not.toEqual(keyPair2.privateKey)
    })
  })

  describe('PQC encapsulation/decapsulation', () => {
    it('should encapsulate and decapsulate successfully', async () => {
      const keyPair = await generatePQCKeyPair()
      
      const encaps = await pqcEncapsulate(keyPair.publicKey, keyPair.privateKey)
      const decSecret = await pqcDecapsulate(keyPair.privateKey, encaps.ciphertext)
      
      expect(decSecret).toEqual(encaps.sharedSecret)
    })

    it('should produce different ciphertexts for same public key', async () => {
      const keyPair = await generatePQCKeyPair()
      
      const encaps1 = await pqcEncapsulate(keyPair.publicKey, keyPair.privateKey)
      const encaps2 = await pqcEncapsulate(keyPair.publicKey, keyPair.privateKey)
      
      expect(encaps1.ciphertext).not.toEqual(encaps2.ciphertext)
    })

    it('should fail decapsulation with wrong private key', async () => {
      const keyPair1 = await generatePQCKeyPair()
      const keyPair2 = await generatePQCKeyPair()
      
      const encaps = await pqcEncapsulate(keyPair1.publicKey, keyPair1.privateKey)
      
      await expect(pqcDecapsulate(keyPair2.privateKey, encaps.ciphertext)).rejects.toThrow()
    })
  })

  describe('hybrid encryption/decryption', () => {
    it('should encrypt and decrypt successfully', async () => {
      const message = 'Hello, SecureMail!'
      
      const result = await hybridEncrypt(message)
      expect(result.privateKey).toBeDefined()
      
      const privateKeyBuffer = new Uint8Array(atob(result.privateKey!).split('').map(c => c.charCodeAt(0)))
      const decrypted = await hybridDecrypt(result, privateKeyBuffer)
      
      expect(decrypted).toBe(message)
    })

    it('should produce different results for same message', async () => {
      const message = 'Hello, SecureMail!'
      
      const result1 = await hybridEncrypt(message)
      const result2 = await hybridEncrypt(message)
      
      expect(result1.encryptedMessage).not.toEqual(result2.encryptedMessage)
      expect(result1.encryptedAESKey).not.toEqual(result2.encryptedAESKey)
      expect(result1.iv).not.toEqual(result2.iv)
    })

    it('should include all required fields in result', async () => {
      const message = 'Test message'
      const result = await hybridEncrypt(message)
      
      expect(result.encryptedMessage).toBeDefined()
      expect(result.encryptedAESKey).toBeDefined()
      expect(result.iv).toBeDefined()
      expect(result.salt).toBeDefined()
      expect(result.publicKey).toBeDefined()
      expect(result.privateKey).toBeDefined()
      expect(result.algorithm).toBe('AES-256-GCM + Argon2id + Kyber-512')
      expect(result.timestamp).toBeTypeOf('number')
    })

    it('should fail decryption with wrong private key', async () => {
      const message = 'Hello, SecureMail!'
      const result = await hybridEncrypt(message)
      
      const wrongPrivateKey = new Uint8Array(1632)
      
      await expect(hybridDecrypt(result, wrongPrivateKey)).rejects.toThrow()
    })
  })
})
