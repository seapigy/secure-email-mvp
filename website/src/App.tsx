import { useState, useEffect, Suspense, lazy } from 'react'
import { AnimatePresence } from './utils/animations'
import Header from './components/Header'
import Hero from './components/Hero'
import TrustBreaker from './components/TrustBreaker'
import SecurityFeaturesGrid from './components/SecurityFeaturesGrid'
import Features from './components/Features'
import Comparison from './components/Comparison'
import Trust from './components/Trust'
import CTA from './components/CTA'
import Footer from './components/Footer'

// Lazy load heavy components
const EncryptionDemo = lazy(() => import('./components/EncryptionDemo'))

function App() {
  const [isDark, setIsDark] = useState(false)

  useEffect(() => {
    // Check system preference - with fallback for test environments
    if (typeof window !== 'undefined' && window.matchMedia) {
      if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        setIsDark(true)
      }
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
        <Suspense fallback={
          <div className="section-padding bg-background dark:bg-primary">
            <div className="max-w-7xl mx-auto text-center">
              <div className="animate-pulse">
                <div className="h-8 bg-gray-300 dark:bg-gray-600 rounded w-1/3 mx-auto mb-4"></div>
                <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-1/2 mx-auto mb-8"></div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                  <div className="h-64 bg-gray-300 dark:bg-gray-600 rounded"></div>
                  <div className="h-64 bg-gray-300 dark:bg-gray-600 rounded"></div>
                </div>
              </div>
            </div>
          </div>
        }>
          <EncryptionDemo />
        </Suspense>
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
