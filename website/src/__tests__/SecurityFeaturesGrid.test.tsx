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
    
    // Check for key security features
    expect(screen.getByText('Geolocation Lock')).toBeInTheDocument()
    expect(screen.getByText('Timed Destruction')).toBeInTheDocument()
    expect(screen.getByText('One-Time Read')).toBeInTheDocument()
    expect(screen.getByText('Password Protection')).toBeInTheDocument()
    expect(screen.getByText('Time Lock')).toBeInTheDocument()
    expect(screen.getByText('Remote Revoke')).toBeInTheDocument()
    expect(screen.getByText('Strip Metadata')).toBeInTheDocument()
    expect(screen.getByText('Tamper Alerts')).toBeInTheDocument()
    expect(screen.getByText('Self-Destruct After Failed Attempts')).toBeInTheDocument()
    expect(screen.getByText('Quantum-Resistant Encryption')).toBeInTheDocument()
    expect(screen.getByText('Zero-Knowledge Architecture')).toBeInTheDocument()
    expect(screen.getByText('Biometric Access')).toBeInTheDocument()
  })

  it('renders the new closed-source fortress feature card', () => {
    render(<SecurityFeaturesGrid />)
    
    expect(screen.getByText('Closed-Source Fortress')).toBeInTheDocument()
    expect(screen.getByText(/No leaks. No exposure. No attack surface./)).toBeInTheDocument()
    expect(screen.getByText(/Our code is sealed from prying eyes, making SecureMail immune to the risks of open-source exploitation./)).toBeInTheDocument()
  })

  it('shows marketing highlight text', () => {
    render(<SecurityFeaturesGrid />)
    
    expect(screen.getByText('No other email system on Earth offers this level of security. Every feature is designed to give you absolute control over your privacy and data.')).toBeInTheDocument()
    expect(screen.getByText('No other email system on Earth offers this.')).toBeInTheDocument()
  })

  it('displays bottom CTA section', () => {
    render(<SecurityFeaturesGrid />)
    
    // The bottom CTA should be present
    expect(screen.getByText('This is Just the Beginning')).toBeInTheDocument()
    expect(screen.getByText('Explore All Features')).toBeInTheDocument()
  })
})
