import React from 'react';
import EmailSendForm from '@/components/email/EmailSendForm';

/**
 * Send Component
 * 
 * Wrapper component that renders the EmailSendForm for the /send route.
 * This provides a clean separation between routing and the actual form logic.
 */
const Send: React.FC = () => {
  return <EmailSendForm />;
};

export default Send; 