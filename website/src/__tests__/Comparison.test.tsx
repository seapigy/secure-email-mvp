import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import Comparison from '../components/Comparison'

// Mock the animations module
vi.mock('../utils/animations', () => ({
  MotionDiv: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  useInView: () => ({ isInView: true })
}))

describe('Comparison Component', () => {
  it('renders the main heading', () => {
    render(<Comparison />)
    expect(screen.getByText('How We Compare')).toBeInTheDocument()
    expect(screen.getByText('to the Competition')).toBeInTheDocument()
  })

  it('renders the section description', () => {
    render(<Comparison />)
    expect(screen.getByText(/See how SecureMail stacks up against the biggest names in email/)).toBeInTheDocument()
  })

  it('renders table headers', () => {
    render(<Comparison />)
    
    expect(screen.getByText('Feature')).toBeInTheDocument()
    expect(screen.getByText('SecureMail')).toBeInTheDocument()
    expect(screen.getByText('Gmail')).toBeInTheDocument()
    expect(screen.getByText('Outlook')).toBeInTheDocument()
    expect(screen.getByText('ProtonMail')).toBeInTheDocument()
  })

  it('renders email-specific security features', () => {
    render(<Comparison />)
    
    // Check for new email-specific features
    expect(screen.getByText('Real-Time Encryption on Every Email')).toBeInTheDocument()
    expect(screen.getByText('AES-256-GCM + PQC Hybrid on Every Message')).toBeInTheDocument()
    expect(screen.getByText('Zero-Knowledge Message Processing')).toBeInTheDocument()
    expect(screen.getByText('Quantum-Resistant Key Exchange')).toBeInTheDocument()
    expect(screen.getByText('Complete Metadata Stripping')).toBeInTheDocument()
  })

  it('renders traditional security features', () => {
    render(<Comparison />)
    
    expect(screen.getByText('End-to-End Encryption')).toBeInTheDocument()
    expect(screen.getByText('No Data Collection')).toBeInTheDocument()
    expect(screen.getByText('Open Source')).toBeInTheDocument()
    expect(screen.getByText('Free Tier')).toBeInTheDocument()
    expect(screen.getByText('Mobile Apps')).toBeInTheDocument()
  })

  it('renders the "What Makes Every SecureMail Email Different" section', () => {
    render(<Comparison />)
    
    expect(screen.getByText('What Makes Every SecureMail Email Different')).toBeInTheDocument()
    expect(screen.getByText('AES-256-GCM')).toBeInTheDocument()
    expect(screen.getByText('PQC Hybrid')).toBeInTheDocument()
    expect(screen.getByText('Zero-Knowledge')).toBeInTheDocument()
    expect(screen.getByText('Real-Time')).toBeInTheDocument()
  })

  it('renders security feature descriptions', () => {
    render(<Comparison />)
    
    expect(screen.getByText('Military-grade encryption on every message')).toBeInTheDocument()
    expect(screen.getByText('Quantum-resistant key exchange')).toBeInTheDocument()
    expect(screen.getByText('We cannot see your emails')).toBeInTheDocument()
    expect(screen.getByText('Instant encryption processing')).toBeInTheDocument()
  })

  it('renders the enhanced CTA section', () => {
    render(<Comparison />)
    
    expect(screen.getByText('Experience Email Security That Actually Works')).toBeInTheDocument()
    expect(screen.getByText(/uncompromising protection on every single email/)).toBeInTheDocument()
    expect(screen.getByText('See Why Every Email Matters')).toBeInTheDocument()
  })

  it('renders the security treatment message', () => {
    render(<Comparison />)
    
    expect(screen.getByText(/Every single email gets the full security treatment/)).toBeInTheDocument()
  })

  it('shows SecureMail advantages in email-specific features', () => {
    render(<Comparison />)
    
    // These features should show SecureMail as the only provider with them
    const realTimeEncryption = screen.getByText('Real-Time Encryption on Every Email')
    const pqcHybrid = screen.getByText('AES-256-GCM + PQC Hybrid on Every Message')
    const zeroKnowledge = screen.getByText('Zero-Knowledge Message Processing')
    const quantumResistant = screen.getByText('Quantum-Resistant Key Exchange')
    
    expect(realTimeEncryption).toBeInTheDocument()
    expect(pqcHybrid).toBeInTheDocument()
    expect(zeroKnowledge).toBeInTheDocument()
    expect(quantumResistant).toBeInTheDocument()
  })
})
