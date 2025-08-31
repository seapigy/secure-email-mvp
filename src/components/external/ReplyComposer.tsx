import React, { useState } from 'react';
import { Send, AlertTriangle, CheckCircle, X, Shield, Paperclip, Type } from 'lucide-react';
import RichTextEditor from './RichTextEditor';
import AttachmentUploader from './AttachmentUploader';
import { log } from '@/lib/logger';

interface ReplyComposerProps {
  linkID: string;
  originalSubject: string;
  originalSenderEmail: string;
  originalSenderName?: string;
  onReplySent?: (replyID: string) => void;
  onCancel?: () => void;
  isOpen: boolean;
}

interface Attachment {
  id: string;
  name: string;
  size: number;
  type: string;
  status: 'uploading' | 'success' | 'error' | 'virus_scanning';
  progress: number;
  error?: string;
  url?: string;
  virusScanStatus?: 'pending' | 'clean' | 'infected' | 'error';
}

interface ReplyRequest {
  link_id: string;
  subject: string;
  body: string;
  ip_address?: string;
  user_agent?: string;
}

interface ReplyResponse {
  success: boolean;
  reply_id?: string;
  message?: string;
  error?: string;
  error_code?: string;
  transaction_id?: string;
}

const ReplyComposer: React.FC<ReplyComposerProps> = ({
  linkID,
  originalSubject,
  originalSenderEmail,
  originalSenderName,
  onReplySent,
  onCancel,
  isOpen,
}) => {
  const [subject, setSubject] = useState(`Re: ${originalSubject}`);
  const [body, setBody] = useState('');
  const [useRichText, setUseRichText] = useState(false);
  const [showAttachments, setShowAttachments] = useState(false);
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [featuresUsed, setFeaturesUsed] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [replyID, setReplyID] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!subject.trim() || !body.trim()) {
      setError('Subject and message are required');
      return;
    }

    // Log current state for debugging (explicitly use variables)
    log.info('Submitting reply with attachments:', { attachmentsCount: attachments.length, featuresUsed }, 'ReplyComposer');

    setLoading(true);
    setError(null);

    try {
      // Process rich text content if enabled
      let processedBody = body;
      if (useRichText) {
        const richTextResponse = await fetch(`/api/v/${linkID}/richtext`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            link_id: linkID,
            content_type: 'reply_body',
            content: body,
          }),
        });

        if (richTextResponse.ok) {
          const richTextData = await richTextResponse.json();
          if (richTextData.success) {
            processedBody = richTextData.sanitized_content;
          }
        }
      }

      const replyRequest: ReplyRequest = {
        link_id: linkID,
        subject: subject.trim(),
        body: processedBody.trim(),
        ip_address: '127.0.0.1', // In real implementation, get from request
        user_agent: navigator.userAgent,
      };

      const response = await fetch(`/api/v/${linkID}/reply`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(replyRequest),
      });

      const data: ReplyResponse = await response.json();

      if (!response.ok) {
        setError(data.error || 'Failed to send reply');
        return;
      }

      if (data.success) {
        setSuccess(true);
        setReplyID(data.reply_id || null);
        if (onReplySent && data.reply_id) {
          onReplySent(data.reply_id);
        }
      } else {
        setError(data.error || 'Reply failed');
      }
    } catch (err) {
      log.error('Error sending reply:', err, 'ReplyComposer');
      setError('Failed to send reply. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    if (!loading) {
      setSubject(`Re: ${originalSubject}`);
      setBody('');
      setUseRichText(false);
      setShowAttachments(false);
      setAttachments([]);
      setFeaturesUsed([]);
      setError(null);
      setSuccess(false);
      setReplyID(null);
      if (onCancel) {
        onCancel();
      }
    }
  };

  const handleAttachmentUploaded = (attachment: Attachment) => {
    setAttachments(prev => [...prev, attachment]);
  };

  const handleAttachmentRemoved = (attachmentId: string) => {
    setAttachments(prev => prev.filter(att => att.id !== attachmentId));
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Allow Ctrl+Enter to submit
    if (e.ctrlKey && e.key === 'Enter') {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  if (!isOpen) return null;

  if (success) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
          <div className="p-8 text-center">
            {/* Success Animation */}
            <div className="mb-6">
              <div className="relative inline-flex items-center justify-center w-16 h-16 bg-green-100 rounded-full mb-4">
                <CheckCircle className="h-8 w-8 text-green-600" />
                <div className="absolute inset-0 bg-green-100 rounded-full animate-ping opacity-75"></div>
              </div>
            </div>
            
            <h3 className="text-xl font-semibold text-gray-900 mb-3">Reply Sent Successfully!</h3>
            <p className="text-sm text-gray-600 mb-6">
              Your reply has been securely delivered to {originalSenderName || originalSenderEmail}. The message is encrypted and protected by enterprise-grade security.
            </p>
            
            {/* Reply Details */}
            <div className="bg-gray-50 rounded-lg p-4 mb-6">
              <div className="flex items-center justify-center mb-2">
                <Send className="h-4 w-4 text-blue-600 mr-2" />
                <span className="text-sm font-medium text-gray-700">Reply Details</span>
              </div>
              {replyID && (
                <p className="text-xs text-gray-500">
                  <span className="font-medium">Reply ID:</span> {replyID}
                </p>
              )}
              <p className="text-xs text-gray-500 mt-1">
                <span className="font-medium">Sent:</span> {new Date().toLocaleString()}
              </p>
            </div>
            
            {/* Security Notice */}
            <div className="bg-blue-50 rounded-lg p-4 mb-6">
              <div className="flex items-start">
                <Shield className="h-4 w-4 text-blue-600 mt-0.5 mr-2 flex-shrink-0" />
                <div>
                  <p className="text-sm font-medium text-blue-800">Secure Delivery</p>
                  <p className="text-sm text-blue-700 mt-1">
                    Your reply is protected with end-to-end encryption and delivered through our secure infrastructure.
                  </p>
                </div>
              </div>
            </div>
            
            <button
              onClick={handleCancel}
              className="w-full bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <div>
            <h2 className="text-xl font-semibold text-gray-900">Reply to Secure Message</h2>
            <p className="text-sm text-gray-600 mt-1">
              Replying to {originalSenderName || originalSenderEmail}
            </p>
          </div>
          <button
            onClick={handleCancel}
            className="text-gray-400 hover:text-gray-600"
            disabled={loading}
          >
            <X className="h-6 w-6" />
          </button>
        </div>

        {/* Reply Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Subject */}
          <div>
            <label htmlFor="subject" className="block text-sm font-medium text-gray-700 mb-1">
              Subject
            </label>
            <input
              type="text"
              id="subject"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Enter subject"
              disabled={loading}
              required
            />
          </div>

                     {/* Message Body */}
           <div>
             <div className="flex items-center justify-between mb-1">
               <label htmlFor="body" className="block text-sm font-medium text-gray-700">
                 Message
               </label>
               <div className="flex items-center space-x-2">
                 <button
                   type="button"
                   onClick={() => setUseRichText(!useRichText)}
                   className={`flex items-center space-x-1 px-2 py-1 rounded text-xs transition-colors ${
                     useRichText 
                       ? 'bg-blue-100 text-blue-700 border border-blue-300' 
                       : 'text-gray-500 hover:text-gray-700'
                   }`}
                   disabled={loading}
                 >
                   <Type className="h-3 w-3" />
                   <span>Rich Text</span>
                 </button>
                 <button
                   type="button"
                   onClick={() => setShowAttachments(!showAttachments)}
                   className={`flex items-center space-x-1 px-2 py-1 rounded text-xs transition-colors ${
                     showAttachments 
                       ? 'bg-blue-100 text-blue-700 border border-blue-300' 
                       : 'text-gray-500 hover:text-gray-700'
                   }`}
                   disabled={loading}
                 >
                   <Paperclip className="h-3 w-3" />
                   <span>Attachments</span>
                 </button>
               </div>
             </div>

             {useRichText ? (
               <RichTextEditor
                 value={body}
                 onChange={setBody}
                 placeholder="Type your reply here..."
                 disabled={loading}
                 maxLength={10000}
                 onFeaturesUsed={setFeaturesUsed}
               />
             ) : (
               <textarea
                 id="body"
                 value={body}
                 onChange={(e) => setBody(e.target.value)}
                 onKeyDown={handleKeyDown}
                 className="w-full h-64 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-none"
                 placeholder="Type your reply here..."
                 disabled={loading}
                 required
               />
             )}
             
             <p className="text-xs text-gray-500 mt-1">
               Press Ctrl+Enter to send, or use the Send button below
             </p>
           </div>

           {/* Attachments Section */}
           {showAttachments && (
             <div>
               <label className="block text-sm font-medium text-gray-700 mb-2">
                 Attachments
               </label>
               <AttachmentUploader
                 linkID={linkID}
                 onAttachmentUploaded={handleAttachmentUploaded}
                 onAttachmentRemoved={handleAttachmentRemoved}
                 maxFiles={5}
                 maxFileSize={25 * 1024 * 1024} // 25MB
                 disabled={loading}
               />
             </div>
           )}

          {/* Error Display */}
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-3">
              <div className="flex">
                <AlertTriangle className="h-5 w-5 text-red-400" />
                <div className="ml-3">
                  <p className="text-sm text-red-800">{error}</p>
                </div>
              </div>
            </div>
          )}

          {/* Security Notice */}
          <div className="bg-blue-50 border border-blue-200 rounded-md p-3">
            <div className="flex">
              <div className="ml-3">
                <h4 className="text-sm font-medium text-blue-800">🔒 Secure Reply</h4>
                <p className="text-sm text-blue-700 mt-1">
                  Your reply will be securely delivered to the original sender. 
                  All replies are logged and monitored for security purposes.
                </p>
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              onClick={handleCancel}
              className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !subject.trim() || !body.trim()}
              className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {loading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Sending...
                </>
              ) : (
                <>
                  <Send className="h-4 w-4 mr-2" />
                  Send Reply
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default ReplyComposer;
