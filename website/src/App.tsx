import { useState, useEffect, Suspense, lazy } from 'react'
import { AnimatePresence } from './utils/animations'
import Header from './components/Header'
import Hero from './components/Hero'
import TrustBreaker from './components/TrustBreaker'
import CTA from './components/CTA'
import Footer from './components/Footer'

// Lazy load heavy components
const EncryptionDemo = lazy(() => import('./components/EncryptionDemo'))
const Features = lazy(() => import('./components/Features'))
const QuantumResistant = lazy(() => import('./components/QuantumResistant'))
const Trust = lazy(() => import('./components/Trust'))
const Comparison = lazy(() => import('./components/Comparison'))
const FAQ = lazy(() => import('./components/FAQ'))

function App() {
  const [isDark, setIsDark] = useState(true)

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
          <div className="section-padding bg-gray-100 dark:bg-dark-800">
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
        <Suspense fallback={
          <div className="section-padding bg-gray-100 dark:bg-dark-800">
            <div className="max-w-7xl mx-auto text-center">
              <div className="animate-pulse">
                <div className="h-8 bg-gray-300 dark:bg-gray-600 rounded w-1/2 mx-auto mb-8"></div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                  {[...Array(9)].map((_, i) => (
                    <div key={i} className="h-32 bg-gray-300 dark:bg-gray-600 rounded"></div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        }>
          <Features />
        </Suspense>
        <Suspense fallback={
          <div className="section-padding bg-gray-100 dark:bg-dark-800">
            <div className="max-w-7xl mx-auto text-center">
              <div className="animate-pulse">
                <div className="h-8 bg-gray-300 dark:bg-gray-600 rounded w-1/2 mx-auto mb-8"></div>
                <div className="h-64 bg-gray-300 dark:bg-gray-600 rounded"></div>
              </div>
            </div>
          </div>
        }>
          <QuantumResistant />
        </Suspense>
        <Suspense fallback={
          <div className="section-padding bg-gray-100 dark:bg-dark-800">
            <div className="max-w-7xl mx-auto text-center">
              <div className="animate-pulse">
                <div className="h-8 bg-gray-300 dark:bg-gray-600 rounded w-1/2 mx-auto mb-8"></div>
                <div className="h-64 bg-gray-300 dark:bg-gray-600 rounded"></div>
              </div>
            </div>
          </div>
        }>
          <Comparison />
        </Suspense>
        <Suspense fallback={
          <div className="section-padding bg-gray-100 dark:bg-dark-800">
            <div className="max-w-7xl mx-auto text-center">
              <div className="animate-pulse">
                <div className="h-8 bg-gray-300 dark:bg-gray-600 rounded w-1/2 mx-auto mb-8"></div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                  {[...Array(4)].map((_, i) => (
                    <div key={i} className="h-32 bg-gray-300 dark:bg-gray-600 rounded"></div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        }>
          <Trust />
        </Suspense>
        <Suspense fallback={
          <div className="section-padding bg-gray-100 dark:bg-dark-800">
            <div className="max-w-7xl mx-auto text-center">
              <div className="animate-pulse">
                <div className="h-8 bg-gray-300 dark:bg-gray-600 rounded w-1/2 mx-auto mb-8"></div>
                <div className="space-y-4">
                  {[...Array(5)].map((_, i) => (
                    <div key={i} className="h-16 bg-gray-300 dark:bg-gray-600 rounded"></div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        }>
          <FAQ />
        </Suspense>
        <CTA />
      </main>
      <Footer />
    </div>
  )
}

export default App
