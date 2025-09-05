import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import Header from './components/Header'
import Hero from './components/Hero'
import TrustBreaker from './components/TrustBreaker'
import EncryptionDemo from './components/EncryptionDemo'
import SecurityFeaturesGrid from './components/SecurityFeaturesGrid'
import Features from './components/Features'
import Comparison from './components/Comparison'
import Trust from './components/Trust'
import CTA from './components/CTA'
import Footer from './components/Footer'

function App() {
  const [isDark, setIsDark] = useState(false)

  useEffect(() => {
    // Check system preference
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      setIsDark(true)
    }
  }, [])

  useEffect(() => {
    if (isDark) {
      document.documentElement.classList.add('dark')
      document.body.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
      document.body.classList.remove('dark')
    }
  }, [isDark])

  return (
    <div className={`min-h-screen transition-colors duration-300 ${
      isDark ? 'dark bg-dark-900 text-white' : 'bg-white text-dark-900'
    }`}>
      <Header isDark={isDark} setIsDark={setIsDark} />
      <main>
        <Hero />
        <TrustBreaker />
        <EncryptionDemo />
        <SecurityFeaturesGrid />
        <Features />
        <Comparison />
        <Trust />
        <CTA />
      </main>
      <Footer />
    </div>
  )
}

export default App
