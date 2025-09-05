import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import FAQ from '../components/FAQ'

// Mock the animations module
vi.mock('../utils/animations', () => ({
  MotionDiv: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  useInView: () => ({ isInView: true })
}))

describe('FAQ Component', () => {
  it('renders the main heading', () => {
    render(<FAQ />)
    expect(screen.getByText('Frequently Asked')).toBeInTheDocument()
    expect(screen.getByText('Questions')).toBeInTheDocument()
  })

  it('renders the section description', () => {
    render(<FAQ />)
    expect(screen.getByText(/Everything you need to know about SecureMail's quantum-resistant encryption/)).toBeInTheDocument()
  })

  it('renders all FAQ questions', () => {
    render(<FAQ />)
    
    expect(screen.getByText('What is Post-Quantum Cryptography (PQC)?')).toBeInTheDocument()
    expect(screen.getByText('Why does SecureMail use PQC encryption?')).toBeInTheDocument()
    expect(screen.getByText('How does PQC work with AES-256-GCM?')).toBeInTheDocument()
    expect(screen.getByText('Is PQC encryption slower than regular encryption?')).toBeInTheDocument()
    expect(screen.getByText('When will quantum computers break current encryption?')).toBeInTheDocument()
  })

  it('shows chevron down icons initially', () => {
    render(<FAQ />)
    const chevronIcons = screen.getAllByRole('button')
    expect(chevronIcons).toHaveLength(5) // 5 FAQ items
  })

  it('expands FAQ item when clicked', () => {
    render(<FAQ />)
    
    const firstQuestion = screen.getByText('What is Post-Quantum Cryptography (PQC)?')
    const firstButton = firstQuestion.closest('button')
    
    // Click to expand
    fireEvent.click(firstButton!)
    
    // Check that the answer is now visible
    expect(screen.getByText(/Post-Quantum Cryptography \(PQC\) refers to cryptographic algorithms/)).toBeInTheDocument()
  })

  it('collapses FAQ item when clicked again', () => {
    render(<FAQ />)
    
    const firstQuestion = screen.getByText('What is Post-Quantum Cryptography (PQC)?')
    const firstButton = firstQuestion.closest('button')
    
    // Click to expand
    fireEvent.click(firstButton!)
    
    // Click to collapse
    fireEvent.click(firstButton!)
    
    // Answer should not be visible
    expect(screen.queryByText(/Post-Quantum Cryptography \(PQC\) refers to cryptographic algorithms/)).not.toBeInTheDocument()
  })

  it('renders FAQ answers with PQC content', () => {
    render(<FAQ />)
    
    // Expand all FAQ items to check answers
    const buttons = screen.getAllByRole('button')
    buttons.forEach(button => fireEvent.click(button))
    
    // Check for PQC-related content in answers
    expect(screen.getByText(/quantum computers are rapidly advancing/)).toBeInTheDocument()
    expect(screen.getByText(/hybrid approach combining both AES-256-GCM and PQC/)).toBeInTheDocument()
    expect(screen.getByText(/PQC algorithms can be slightly slower/)).toBeInTheDocument()
    expect(screen.getByText(/quantum computers capable of breaking current encryption/)).toBeInTheDocument()
  })

  it('renders the bottom CTA section', () => {
    render(<FAQ />)
    
    expect(screen.getByText('Still Have Questions?')).toBeInTheDocument()
    expect(screen.getByText(/Our security experts are here to help/)).toBeInTheDocument()
    expect(screen.getByText('Contact Security Team')).toBeInTheDocument()
  })

  it('allows multiple FAQ items to be open simultaneously', () => {
    render(<FAQ />)
    
    const buttons = screen.getAllByRole('button')
    
    // Click first two buttons
    fireEvent.click(buttons[0])
    fireEvent.click(buttons[1])
    
    // Both answers should be visible
    expect(screen.getByText(/Post-Quantum Cryptography \(PQC\) refers to cryptographic algorithms/)).toBeInTheDocument()
    expect(screen.getByText(/quantum computers are rapidly advancing/)).toBeInTheDocument()
  })
})
