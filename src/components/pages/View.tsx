import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeftIcon,
  ArrowUturnLeftIcon,
  ArrowUturnRightIcon,
  TrashIcon,
  StarIcon,
} from '@heroicons/react/24/outline';

/**
 * View Component
 * 
 * Individual email viewing interface with full email details.
 * Features email actions like reply, forward, and delete.
 * 
 * Features:
 * - Full email content display
 * - Email metadata (from, to, date, etc.)
 * - Action buttons (reply, forward, delete)
 * - Responsive design
 */
const View: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  // Mock email data for demonstration
  const email = {
    id: id || '1',
    from: 'john@securesystem.email',
    to: 'user@securesystem.email',
    subject: 'Welcome to SecureMail',
    date: '2024-01-15T10:30:00Z',
    content: `
      <p>Dear User,</p>
      
      <p>Welcome to SecureMail! We're excited to have you on board.</p>
      
      <p>SecureMail is a privacy-first email solution that provides:</p>
      <ul>
        <li>End-to-end encryption for all emails</li>
        <li>Secure link-based delivery</li>
        <li>TOTP authentication for enhanced security</li>
        <li>Modern, responsive interface</li>
      </ul>
      
      <p>Your account has been successfully configured and you can now start sending and receiving secure emails.</p>
      
      <p>If you have any questions or need assistance, please don't hesitate to contact our support team.</p>
      
      <p>Best regards,<br>
      The SecureMail Team</p>
    `,
    attachments: [
      { name: 'welcome-guide.pdf', size: '2.3 MB' },
      { name: 'security-overview.docx', size: '1.1 MB' },
    ],
  };

  const handleReply = () => {
    // Future implementation: Pre-populate send form with reply data
    navigate('/send');
  };

  const handleForward = () => {
    // Future implementation: Pre-populate send form with forward data
    navigate('/send');
  };

  const handleDelete = () => {
    // Future implementation: Delete email and return to dashboard
    navigate('/inbox');
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <button
            onClick={() => navigate('/inbox')}
            className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
          >
            <ArrowLeftIcon className="w-5 h-5" />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-secondary-900 dark:text-white">
              {email.subject}
            </h1>
            <p className="text-secondary-600 dark:text-secondary-400">
              Email ID: {email.id}
            </p>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <button className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200">
            <StarIcon className="w-5 h-5" />
          </button>
          <button
            onClick={handleReply}
            className="flex items-center space-x-2 px-4 py-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
          >
            <ArrowUturnLeftIcon className="w-4 h-4" />
            <span>Reply</span>
          </button>
          <button
            onClick={handleForward}
            className="flex items-center space-x-2 px-4 py-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
          >
            <ArrowUturnRightIcon className="w-4 h-4" />
            <span>Forward</span>
          </button>
          <button
            onClick={handleDelete}
            className="flex items-center space-x-2 px-4 py-2 text-red-600 hover:bg-red-100 dark:text-red-400 dark:hover:bg-red-900/20 rounded-lg transition-colors duration-200"
          >
            <TrashIcon className="w-4 h-4" />
            <span>Delete</span>
          </button>
        </div>
      </div>

      {/* Email Content */}
      <div className="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700">
        <div className="p-6 space-y-6">
          {/* Email Metadata */}
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">
                  From
                </label>
                <p className="text-secondary-900 dark:text-white">
                  {email.from}
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">
                  To
                </label>
                <p className="text-secondary-900 dark:text-white">
                  {email.to}
                </p>
              </div>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">
                Date
              </label>
              <p className="text-secondary-900 dark:text-white">
                {new Date(email.date).toLocaleString()}
              </p>
            </div>
          </div>

          {/* Attachments */}
          {email.attachments && email.attachments.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                Attachments
              </label>
              <div className="space-y-2">
                {email.attachments.map((attachment, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between p-3 bg-secondary-50 dark:bg-secondary-700 rounded-lg"
                  >
                    <div className="flex items-center space-x-3">
                      <div className="w-8 h-8 bg-secondary-200 dark:bg-secondary-600 rounded flex items-center justify-center">
                        <span className="text-xs font-medium text-secondary-700 dark:text-secondary-300">
                          {attachment.name.split('.').pop()?.toUpperCase()}
                        </span>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          {attachment.name}
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          {attachment.size}
                        </p>
                      </div>
                    </div>
                    <button className="text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 text-sm font-medium">
                      Download
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Email Content */}
          <div>
            <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Message
            </label>
            <div 
              className="prose prose-sm max-w-none dark:prose-invert"
              dangerouslySetInnerHTML={{ __html: email.content }}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default View; 