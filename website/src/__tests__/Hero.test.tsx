import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import Hero from '../components/Hero'

// Mock the animations module
vi.mock('../utils/animations', () => ({
  MotionDiv: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  MotionButton: ({ children, ...props }: any) => <button {...props}>{children}</button>,
  fadeIn: {},
  slideUp: {},
  slideInLeft: {},
  slideInRight: {}
}))

describe('Hero Component', () => {
  it('renders the main headline', () => {
    render(<Hero />)
    expect(screen.getByText("The World's")).toBeInTheDocument()
    expect(screen.getByText('Most Secure')).toBeInTheDocument()
    expect(screen.getByText('Email.')).toBeInTheDocument()
  })

  it('renders the main subheadline', () => {
    render(<Hero />)
    expect(screen.getByText('The Most Secure Email in the World')).toBeInTheDocument()
  })

  it('renders the PQC subheadline', () => {
    render(<Hero />)
    expect(screen.getByText("World's First Quantum-Resistant Email.")).toBeInTheDocument()
  })

  it('renders encryption technology mentions', () => {
    render(<Hero />)
    expect(screen.getAllByText('AES-256-GCM')).toHaveLength(2) // Main text + trust indicator
    expect(screen.getByText('Argon2id')).toBeInTheDocument()
    expect(screen.getByText('TLS 1.3')).toBeInTheDocument()
    expect(screen.getByText('PQC hybrid encryption')).toBeInTheDocument()
  })

  it('renders privacy statements', () => {
    render(<Hero />)
    expect(screen.getByText('Zero-knowledge.')).toBeInTheDocument()
    expect(screen.getByText('Zero visibility.')).toBeInTheDocument()
    expect(screen.getByText('Absolute privacy.')).toBeInTheDocument()
  })

  it('renders CTA buttons', () => {
    render(<Hero />)
    expect(screen.getByText('Get Early Access')).toBeInTheDocument()
    expect(screen.getByText('Learn More')).toBeInTheDocument()
  })

  it('renders trust indicators', () => {
    render(<Hero />)
    expect(screen.getAllByText('AES-256-GCM')).toHaveLength(2) // Main text + trust indicator
    expect(screen.getByText('Zero-Knowledge')).toBeInTheDocument()
    expect(screen.getByText('No Visibility')).toBeInTheDocument()
    expect(screen.getByText('PQC Ready')).toBeInTheDocument()
  })

  it('renders scroll indicator', () => {
    render(<Hero />)
    // The scroll indicator is present as a visual element
    const scrollIndicator = document.querySelector('.animate-float')
    expect(scrollIndicator).toBeInTheDocument()
  })

  it('has tooltips on AES mentions', () => {
    render(<Hero />)
    
    // Check for tooltip attributes on AES mentions
    const aesMentions = screen.getAllByText('AES-256-GCM')
    expect(aesMentions.length).toBeGreaterThan(0)
    
    // Check that at least one has a title attribute (tooltip)
    const hasTooltip = aesMentions.some(element => 
      element.getAttribute('title')?.includes('PQC')
    )
    expect(hasTooltip).toBe(true)
  })
})
