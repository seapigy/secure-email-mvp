import React, { useState } from 'react';
import { useUIStore } from '@/stores/uiStore';
import { cn, formatRelativeTime, truncateText, getInitials, getRandomColor } from '@/lib/utils';
import {
  StarIcon,
  PaperClipIcon,
  EyeIcon,
  EyeSlashIcon,
  TrashIcon,
  ArchiveBoxIcon,
  EnvelopeIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline';
import { StarIcon as StarIconSolid } from '@heroicons/react/24/solid';
import type { Email } from '@/types';

interface EmailListProps {
  className?: string;
}

const EmailList: React.FC<EmailListProps> = ({ className }) => {
  const { selectedEmails, toggleEmailSelection, clearEmailSelection } = useUIStore();
  const [hoveredEmail, setHoveredEmail] = useState<string | null>(null);

  // Mock emails data - in real app this would come from API
  const emails: Email[] = [
    {
      id: '1',
      subject: 'Security Update: New Encryption Features Available',
      sender: 'security@securemail.com',
      recipients: ['user@example.com'],
      content: 'We are excited to announce new end-to-end encryption features that will enhance your email security...',
      isRead: false,
      isStarred: true,
      isEncrypted: true,
      hasAttachments: false,
      createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2 hours ago
      receivedAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
      labels: ['important', 'security'],
    },
    {
      id: '2',
      subject: 'Meeting Reminder: Team Sync Tomorrow',
      sender: 'manager@company.com',
      recipients: ['user@example.com'],
      content: 'Hi team, just a reminder that we have our weekly sync meeting tomorrow at 10 AM...',
      isRead: true,
      isStarred: false,
      isEncrypted: false,
      hasAttachments: true,
      attachments: [
        { id: 'att1', name: 'agenda.pdf', size: 1024000, type: 'application/pdf', isEncrypted: false },
      ],
      createdAt: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(), // 4 hours ago
      receivedAt: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(),
      labels: ['work'],
    },
    {
      id: '3',
      subject: 'Your order has been shipped',
      sender: 'noreply@shop.com',
      recipients: ['user@example.com'],
      content: 'Great news! Your order #12345 has been shipped and is on its way to you...',
      isRead: true,
      isStarred: false,
      isEncrypted: false,
      hasAttachments: false,
      createdAt: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString(), // 6 hours ago
      receivedAt: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString(),
      labels: [],
    },
    {
      id: '4',
      subject: 'Important: Account Security Alert',
      sender: 'alerts@securemail.com',
      recipients: ['user@example.com'],
      content: 'We detected a login attempt from an unrecognized device. Please verify this was you...',
      isRead: false,
      isStarred: true,
      isEncrypted: true,
      hasAttachments: false,
      createdAt: new Date(Date.now() - 8 * 60 * 60 * 1000).toISOString(), // 8 hours ago
      receivedAt: new Date(Date.now() - 8 * 60 * 60 * 1000).toISOString(),
      labels: ['security', 'important'],
    },
    {
      id: '5',
      subject: 'Weekly Newsletter: Tech Updates',
      sender: 'newsletter@tech.com',
      recipients: ['user@example.com'],
      content: 'Stay up to date with the latest technology news and updates from around the world...',
      isRead: true,
      isStarred: false,
      isEncrypted: false,
      hasAttachments: false,
      createdAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), // 1 day ago
      receivedAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      labels: [],
    },
    {
      id: '6',
      subject: 'Project Proposal: Q4 Initiatives',
      sender: 'colleague@company.com',
      recipients: ['user@example.com'],
      content: 'I\'ve attached the Q4 project proposal for your review. Please let me know your thoughts...',
      isRead: true,
      isStarred: false,
      isEncrypted: false,
      hasAttachments: true,
      attachments: [
        { id: 'att2', name: 'proposal.docx', size: 2048000, type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', isEncrypted: false },
        { id: 'att3', name: 'budget.xlsx', size: 512000, type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', isEncrypted: false },
      ],
      createdAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(), // 2 days ago
      receivedAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
      labels: ['work', 'important'],
    },
  ];

  const handleEmailClick = (emailId: string) => {
    // In real app, this would navigate to email detail view
    console.log('Opening email:', emailId);
  };

  const handleStarToggle = (emailId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    // In real app, this would update the email star status via API
    console.log('Toggling star for email:', emailId);
  };

  const handleBulkAction = (action: 'delete' | 'archive' | 'mark-read' | 'mark-unread') => {
    if (selectedEmails.length === 0) return;
    
    // In real app, this would perform bulk actions via API
    console.log(`Performing ${action} on emails:`, selectedEmails);
    clearEmailSelection();
  };

  return (
    <div className={cn('flex flex-col h-full', className)}>
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-secondary-200 dark:border-secondary-700">
        <div className="flex items-center space-x-4">
          <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
            Inbox
          </h2>
          <span className="text-sm text-secondary-500 dark:text-secondary-400">
            {emails.filter(email => !email.isRead).length} unread
          </span>
        </div>
        
        {selectedEmails.length > 0 && (
          <div className="flex items-center space-x-2">
            <button
              onClick={() => handleBulkAction('mark-read')}
              className="p-2 text-secondary-600 hover:text-secondary-900 dark:text-secondary-400 dark:hover:text-white transition-colors"
            >
              <EyeIcon className="h-4 w-4" />
            </button>
            <button
              onClick={() => handleBulkAction('mark-unread')}
              className="p-2 text-secondary-600 hover:text-secondary-900 dark:text-secondary-400 dark:hover:text-white transition-colors"
            >
              <EyeSlashIcon className="h-4 w-4" />
            </button>
            <button
              onClick={() => handleBulkAction('archive')}
              className="p-2 text-secondary-600 hover:text-secondary-900 dark:text-secondary-400 dark:hover:text-white transition-colors"
            >
              <ArchiveBoxIcon className="h-4 w-4" />
            </button>
            <button
              onClick={() => handleBulkAction('delete')}
              className="p-2 text-error-600 hover:text-error-700 dark:text-error-400 dark:hover:text-error-300 transition-colors"
            >
              <TrashIcon className="h-4 w-4" />
            </button>
          </div>
        )}
      </div>

      {/* Email List */}
      <div className="flex-1 overflow-y-auto">
        {emails.map((email) => (
          <div
            key={email.id}
            className={cn(
              'flex items-center px-4 py-3 border-b border-secondary-100 dark:border-secondary-800 cursor-pointer transition-colors',
              !email.isRead && 'bg-primary-50 dark:bg-primary-900/20',
              selectedEmails.includes(email.id) && 'bg-primary-100 dark:bg-primary-900/40',
              hoveredEmail === email.id && 'bg-secondary-50 dark:bg-secondary-800/50'
            )}
            onClick={() => handleEmailClick(email.id)}
            onMouseEnter={() => setHoveredEmail(email.id)}
            onMouseLeave={() => setHoveredEmail(null)}
          >
            {/* Checkbox */}
            <input
              type="checkbox"
              checked={selectedEmails.includes(email.id)}
              onChange={(e) => {
                e.stopPropagation();
                toggleEmailSelection(email.id);
              }}
              className="mr-3 h-4 w-4 text-primary-600 focus:ring-primary-500 border-secondary-300 rounded"
            />

            {/* Star */}
            <button
              onClick={(e) => handleStarToggle(email.id, e)}
              className="mr-3 p-1 text-secondary-400 hover:text-warning-500 transition-colors"
            >
              {email.isStarred ? (
                <StarIconSolid className="h-4 w-4 text-warning-500" />
              ) : (
                <StarIcon className="h-4 w-4" />
              )}
            </button>

            {/* Encryption Icon */}
            {email.isEncrypted && (
              <ShieldCheckIcon className="mr-2 h-4 w-4 text-success-600" />
            )}

            {/* Sender Avatar */}
            <div className={cn('mr-3 h-8 w-8 rounded-full flex items-center justify-center text-white text-sm font-medium', getRandomColor())}>
              {getInitials(email.sender || '')}
            </div>

            {/* Email Content */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2 min-w-0">
                  <span className={cn(
                    'text-sm font-medium truncate',
                    !email.isRead ? 'text-secondary-900 dark:text-white' : 'text-secondary-700 dark:text-secondary-300'
                  )}>
                    {email.sender}
                  </span>
                  {email.hasAttachments && (
                    <PaperClipIcon className="h-4 w-4 text-secondary-400 flex-shrink-0" />
                  )}
                </div>
                <span className="text-xs text-secondary-500 dark:text-secondary-400 flex-shrink-0">
                  {formatRelativeTime(email.receivedAt || '')}
                </span>
              </div>
              <div className="flex items-center justify-between mt-1">
                <span className={cn(
                  'text-sm truncate',
                  !email.isRead ? 'font-semibold text-secondary-900 dark:text-white' : 'text-secondary-700 dark:text-secondary-300'
                )}>
                  {email.subject}
                </span>
              </div>
              <p className="text-xs text-secondary-500 dark:text-secondary-400 truncate mt-1">
                {truncateText(email.content, 80)}
              </p>
            </div>
          </div>
        ))}
      </div>

      {/* Empty State */}
      {emails.length === 0 && (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <EnvelopeIcon className="h-12 w-12 text-secondary-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-secondary-900 dark:text-white mb-2">
              No emails found
            </h3>
            <p className="text-secondary-500 dark:text-secondary-400">
              Your inbox is empty. New emails will appear here.
            </p>
          </div>
        </div>
      )}
    </div>
  );
};

export default EmailList; 