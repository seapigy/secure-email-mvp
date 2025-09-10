import { MotionDiv, useInView } from '../utils/animations'
import { useRef, useState } from 'react'
import { ChevronDown, ChevronUp } from '../utils/icons'

export default function FAQ() {
  const ref = useRef(null)
  const { isInView } = useInView({ once: true, margin: "-100px" })
  const [openItems, setOpenItems] = useState<number[]>([])

  const faqs = [
    {
      question: "What is Post-Quantum Cryptography (PQC)?",
      answer: "Post-Quantum Cryptography (PQC) refers to cryptographic algorithms that are designed to be secure against both classical and quantum computers. While current encryption like AES-256 is secure against classical computers, quantum computers could potentially break it. PQC algorithms are specifically designed to resist quantum attacks, ensuring your data remains secure even in the post-quantum era."
    },
    {
      question: "Why does SecureMail use PQC encryption?",
      answer: "SecureMail implements PQC because quantum computers are rapidly advancing and could break current encryption standards within the next decade. By using quantum-resistant algorithms today, we ensure your emails remain secure for decades to come, protecting your privacy even as technology evolves."
    },
    {
      question: "How does PQC work with AES-256-GCM?",
      answer: "SecureMail uses a hybrid approach combining both AES-256-GCM and PQC algorithms. The PQC system generates quantum-resistant keys that are used to encrypt the AES keys, which then encrypt your actual message content. This provides the security of quantum-resistant encryption with the performance benefits of AES-256-GCM."
    },
    {
      question: "Is PQC encryption slower than regular encryption?",
      answer: "PQC algorithms can be slightly slower than traditional encryption, but SecureMail's hybrid approach minimizes this impact. We use PQC for key exchange and AES-256-GCM for message encryption, giving you the best of both worlds: quantum resistance and optimal performance."
    },
    {
      question: "When will quantum computers break current encryption?",
      answer: "While exact timelines are uncertain, experts predict that quantum computers capable of breaking current encryption could emerge between 2025-2030. SecureMail's PQC implementation ensures you're protected today, so you don't need to worry about future quantum threats."
    },
    {
      question: "What security features are currently available?",
      answer: "SecureMail currently provides military-grade authentication with Argon2id password hashing, TOTP two-factor authentication, and quantum-resistant encryption. Our zero-knowledge architecture ensures we cannot access your data. Advanced email security features like self-destruct timers and geo-restrictions are in development and will be available in future updates."
    },
    {
      question: "How does SecureMail's authentication work?",
      answer: "SecureMail uses industry-leading security standards. Passwords are hashed with Argon2id (resistant to both classical and quantum attacks), and we support TOTP-based two-factor authentication. All user data is encrypted with AES-256-GCM, and we're implementing Post-Quantum Cryptography for future-proof security."
    },
    {
      question: "What account types are available?",
      answer: "SecureMail offers multiple account tiers: Free accounts for personal use, Premium accounts with enhanced features, and Enterprise accounts for organizations requiring advanced security and compliance features. Each tier includes our core quantum-resistant encryption and zero-knowledge architecture."
    },
    {
      question: "How does email verification work?",
      answer: "When you sign up, SecureMail sends a verification email to confirm your address. This ensures only legitimate users can access the platform and helps maintain the security of our network. The verification process is quick and secure, using our encrypted communication channels."
    },
    {
      question: "Is my data really secure with zero-knowledge architecture?",
      answer: "Yes! Our zero-knowledge architecture means we literally cannot see your emails, even if we wanted to. All data is encrypted client-side before transmission, and we use quantum-resistant encryption algorithms. This isn't just marketing - it's a technical reality that ensures your privacy is protected at the highest level."
    }
  ]

  const toggleItem = (index: number) => {
    setOpenItems(prev => 
      prev.includes(index) 
        ? prev.filter(i => i !== index)
        : [...prev, index]
    )
  }

  return (
    <section ref={ref} className="section-padding bg-white dark:bg-dark-800">
      <div className="max-w-4xl mx-auto">
        {/* Section Header */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8 }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl md:text-6xl font-bold mb-6">
            <span className="gradient-text">Frequently Asked</span>
            <br />
            <span className="text-dark-900 dark:text-white">Questions</span>
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            Everything you need to know about SecureMail's quantum-resistant encryption, 
            advanced security features, and how easy it is to customize protection for each email.
          </p>
        </MotionDiv>

        {/* FAQ Items */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="space-y-4"
        >
          {faqs.map((faq, index) => (
            <MotionDiv
              key={index}
              initial={{ opacity: 0, y: 20 }}
              animate={isInView ? { opacity: 1, y: 0 } : {}}
              transition={{ duration: 0.5, delay: 0.4 + index * 0.1 }}
              className="bg-white dark:bg-dark-800 rounded-2xl border border-gray-200 dark:border-gray-700 overflow-hidden shadow-sm hover:shadow-md transition-shadow"
            >
              <button
                onClick={() => toggleItem(index)}
                className="w-full px-6 py-6 text-left flex items-center justify-between hover:bg-gray-50 dark:hover:bg-dark-700 transition-colors"
              >
                <h3 className="text-lg font-semibold text-dark-900 dark:text-white pr-4">
                  {faq.question}
                </h3>
                {openItems.includes(index) ? (
                  <ChevronUp className="w-5 h-5 text-indigo-500 flex-shrink-0" />
                ) : (
                  <ChevronDown className="w-5 h-5 text-indigo-500 flex-shrink-0" />
                )}
              </button>
              
              {openItems.includes(index) && (
                <MotionDiv
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: "auto" }}
                  exit={{ opacity: 0, height: 0 }}
                  transition={{ duration: 0.3 }}
                  className="px-6 pb-6"
                >
                  <p className="text-gray-600 dark:text-gray-300 leading-relaxed">
                    {faq.answer}
                  </p>
                </MotionDiv>
              )}
            </MotionDiv>
          ))}
        </MotionDiv>

        {/* Bottom CTA */}
        <MotionDiv
          initial={{ opacity: 0, y: 30 }}
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="text-center mt-16"
        >
          <div className="glass-effect rounded-3xl p-8 max-w-3xl mx-auto">
            <h3 className="text-2xl md:text-3xl font-bold mb-4 text-dark-900 dark:text-white">
              Still Have Questions?
            </h3>
            
            <p className="text-lg text-gray-600 dark:text-gray-300 mb-6">
              Our security experts are here to help you understand how SecureMail 
              protects your privacy with quantum-resistant encryption and customizable security controls.
            </p>
            
            <MotionDiv
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-primary text-lg px-8 py-3"
              as="button"
            >
              Contact Security Team
            </MotionDiv>
          </div>
        </MotionDiv>
      </div>
    </section>
  )
}
