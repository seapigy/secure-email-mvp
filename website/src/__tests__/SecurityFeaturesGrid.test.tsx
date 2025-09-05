// DO NOT EDIT EXISTING CODE - This is a new test file for Phase 2

import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import SecurityFeaturesGrid from '../components/SecurityFeaturesGrid'

describe('SecurityFeaturesGrid Component', () => {
  it('renders security features grid section', () => {
    render(<SecurityFeaturesGrid />)
    
    expect(screen.getByText('Security Beyond')).toBeInTheDocument()
    expect(screen.getByText('Imagination')).toBeInTheDocument()
    expect(screen.getByText('"Security beyond imagination."')).toBeInTheDocument()
  })

  it('displays all security feature cards', () => {
    render(<SecurityFeaturesGrid />)
    
    // Check for key security features that actually exist in the component
    expect(screen.getByText('Zero-Knowledge Architecture')).toBeInTheDocument()
    expect(screen.getByText('Quantum-Resistant Encryption')).toBeInTheDocument()
    expect(screen.getByText('Decoy Messages')).toBeInTheDocument()
    expect(screen.getByText('Biometric Access')).toBeInTheDocument()
    expect(screen.getByText('Closed-Source Fortress')).toBeInTheDocument()
    expect(screen.getByText('Advanced Geolocation')).toBeInTheDocument()
    expect(screen.getByText('Smart Destruction')).toBeInTheDocument()
    expect(screen.getByText('Threat Intelligence')).toBeInTheDocument()
    expect(screen.getByText('Quantum Erasure')).toBeInTheDocument()
    expect(screen.getByText('Stealth Mode')).toBeInTheDocument()
    expect(screen.getByText('Holographic Security')).toBeInTheDocument()
    expect(screen.getByText('Temporal Encryption')).toBeInTheDocument()
  })

  it('renders the closed-source fortress feature card', () => {
    render(<SecurityFeaturesGrid />)
    
    expect(screen.getByText('Closed-Source Fortress')).toBeInTheDocument()
    expect(screen.getByText(/No leaks. No exposure. No attack surface./)).toBeInTheDocument()
    expect(screen.getByText(/Our code is sealed from prying eyes/)).toBeInTheDocument()
  })

  it('shows marketing highlight text', () => {
    render(<SecurityFeaturesGrid />)
    
    expect(screen.getByText(/Revolutionary security features that push the boundaries/)).toBeInTheDocument()
    expect(screen.getByText(/These advanced capabilities exist nowhere else in the email world/)).toBeInTheDocument()
  })

  it('displays bottom CTA section', () => {
    render(<SecurityFeaturesGrid />)
    
    // The bottom CTA should be present with the actual text from the component
    expect(screen.getByText('The Future of Email Security')).toBeInTheDocument()
    expect(screen.getByText('Explore All Features')).toBeInTheDocument()
    expect(screen.getByText(/These revolutionary features represent the cutting edge/)).toBeInTheDocument()
  })
})