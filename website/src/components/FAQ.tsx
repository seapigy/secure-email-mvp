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
      question: "How easy is it to add security features to each email?",
      answer: "Extremely easy! When composing an email, you'll see a simple security panel with toggle switches for each feature. Just click to enable password protection, set self-destruct timers, add geo-restrictions, or configure view limits. Each feature can be applied individually or combined for maximum security. The interface is designed to be intuitive - no technical knowledge required."
    },
    {
      question: "Can I customize security settings for different types of emails?",
      answer: "Absolutely! You can create security presets for different scenarios. For example, set up a 'High Security' preset with password protection, geo-restrictions, and self-destruct timers for sensitive business emails. Or create a 'Quick Secure' preset with just basic encryption for everyday use. You can also apply different settings to each individual email as needed."
    },
    {
      question: "Do I need to configure security features every time I send an email?",
      answer: "No! You can set default security preferences in your account settings, and they'll be applied automatically to all outgoing emails. You can then override these defaults on a per-email basis when you need different security levels. This gives you both convenience and flexibility."
    },
    {
      question: "How do recipients know if an email has special security features?",
      answer: "Recipients receive clear, user-friendly notifications about any security features enabled on the email. For example, if password protection is enabled, they'll see a simple prompt to enter the password. If geo-restrictions are active, they'll be informed if they're in an allowed location. The interface guides them through each security step without confusion."
    },
    {
      question: "Can I change security settings after sending an email?",
      answer: "Yes, for certain features! You can extend or reduce self-destruct timers, add or remove view limits, and modify geo-restrictions even after sending. However, once an email is opened or a self-destruct timer expires, those actions cannot be reversed. This gives you ongoing control while maintaining security integrity."
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
