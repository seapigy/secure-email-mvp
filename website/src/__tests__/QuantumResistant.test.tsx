import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import QuantumResistant from '../components/QuantumResistant'

// Mock the animations module
vi.mock('../utils/animations', () => ({
  MotionDiv: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  useInView: () => ({ isInView: true })
}))

describe('QuantumResistant Component', () => {
  it('renders the main heading', () => {
    render(<QuantumResistant />)
    expect(screen.getByText('Why Quantum-Resistant')).toBeInTheDocument()
    expect(screen.getByText('Matters')).toBeInTheDocument()
  })

  it('renders the technical context section', () => {
    render(<QuantumResistant />)
    expect(screen.getByText('The Quantum Threat Explained')).toBeInTheDocument()
    expect(screen.getByText('Why Current Encryption Fails')).toBeInTheDocument()
    expect(screen.getByText('The PQC Solution')).toBeInTheDocument()
  })

  it('renders all timeline items with correct years', () => {
    render(<QuantumResistant />)
    
    // Check for all timeline years
    expect(screen.getByText('1994')).toBeInTheDocument()
    expect(screen.getByText('2000s')).toBeInTheDocument()
    expect(screen.getByText('2010s')).toBeInTheDocument()
    expect(screen.getByText('2020-2024')).toBeInTheDocument()
    expect(screen.getByText('2024')).toBeInTheDocument()
    expect(screen.getByText('2025-2030')).toBeInTheDocument()
    expect(screen.getByText('2030+')).toBeInTheDocument()
  })

  it('renders timeline titles', () => {
    render(<QuantumResistant />)
    
    expect(screen.getByText("Shor's Algorithm Discovery")).toBeInTheDocument()
    expect(screen.getByText('Early Quantum Research')).toBeInTheDocument()
    expect(screen.getByText('Quantum Supremacy Race')).toBeInTheDocument()
    expect(screen.getByText('PQC Standardization')).toBeInTheDocument()
    expect(screen.getByText('Today - SecureMail Leads')).toBeInTheDocument()
    expect(screen.getByText('Quantum Threat Emerges')).toBeInTheDocument()
    expect(screen.getByText('Post-Quantum Era')).toBeInTheDocument()
  })

  it('renders timeline descriptions', () => {
    render(<QuantumResistant />)
    
    expect(screen.getByText(/Peter Shor proves quantum computers can break RSA/)).toBeInTheDocument()
    expect(screen.getByText(/IBM, Google, and others begin serious quantum computing research/)).toBeInTheDocument()
    expect(screen.getByText(/Google achieves quantum supremacy in 2019/)).toBeInTheDocument()
    expect(screen.getByText(/NIST selects final PQC algorithms/)).toBeInTheDocument()
    expect(screen.getByText(/SecureMail already implements PQC hybrid encryption/)).toBeInTheDocument()
  })

  it('renders the bottom CTA section', () => {
    render(<QuantumResistant />)
    
    expect(screen.getByText('Future-Proof Security Today')).toBeInTheDocument()
    expect(screen.getByText(/Don't wait for quantum computers to break your encryption/)).toBeInTheDocument()
    expect(screen.getByText('Learn About PQC')).toBeInTheDocument()
  })

  it('renders technical explanations', () => {
    render(<QuantumResistant />)
    
    expect(screen.getByText(/Shor's algorithm allows quantum computers to factor large numbers/)).toBeInTheDocument()
    expect(screen.getByText(/Post-Quantum Cryptography uses mathematical problems/)).toBeInTheDocument()
  })
})
