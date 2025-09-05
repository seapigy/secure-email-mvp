import { MotionDiv, useInView } from '../utils/animations'
import { useRef } from 'react'
import { Shield, Building2, Users, Award, Globe, Lock } from '../utils/icons'
import ComplianceFeatures from './Trust/ComplianceFeatures'
import EnterpriseFeatures from './Trust/EnterpriseFeatures'
import TrustSection from './Trust/TrustSection'

export default function Trust() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })

  return (
    <section ref={ref} id="trust" className="section-padding bg-gradient-to-b from-gray-100 to-gray-200 dark:from-dark-800 dark:to-dark-900">
      <div className="max-w-7xl mx-auto">
        {/* Main Trust Section */}
        <TrustSection />
        
        {/* Compliance Features */}
        <ComplianceFeatures />
        
        {/* Enterprise Features */}
        <EnterpriseFeatures />
      </div>
    </section>
  )
}