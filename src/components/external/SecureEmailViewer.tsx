import React, { useState, useEffect } from 'react';
import { Shield, Mail, Download, Reply, AlertTriangle } from 'lucide-react';

// =============================================================================
// EXTERNAL SECURE EMAIL VIEWER
// =============================================================================

interface SecureEmailViewerProps {
  linkId: string;
  sessionToken?: string;
}

interface EmailData {
  subject: string;
  body: string;
  senderName: string;
  senderEmail: string;
  sentAt: string;
  attachments: Attachment[];
  securityInfo: SecurityInfo;
}

interface Attachment {
  id: string;
  name: string;
  size: number;
  contentType: string;
  secureUrl: string;
}

interface SecurityInfo {
  isSecure: boolean;
  encryptionType: string;
  expiresAt?: string;
  readOnce: boolean;
  autoDestruct: boolean;
}

export const SecureEmailViewer: React.FC<SecureEmailViewerProps> = ({ linkId, sessionToken }) => {
  const [emailData, setEmailData] = useState<EmailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showReply, setShowReply] = useState(false);
  const [replyText, setReplyText] = useState('');

  useEffect(() => {
    loadEmailData();
  }, [linkId, sessionToken]);

  const loadEmailData = async () => {
    try {
      setLoading(true);
      // TODO: Implement API call to load secure email data
      console.log('Loading email data for link:', linkId);
      
      // Mock data for now
      const mockData: EmailData = {
        subject: 'Secure Email Subject',
        body: 'This is a secure email body content.',
        senderName: 'John Doe',
        senderEmail: 'john@company.com',
        sentAt: new Date().toISOString(),
        attachments: [],
        securityInfo: {
          isSecure: true,
          encryptionType: 'AES-256-GCM',
          readOnce: false,
          autoDestruct: false
        }
      };
      
      setEmailData(mockData);
    } catch (err) {
      setError('Failed to load email data');
      console.error('Error loading email:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleReply = async () => {
    try {
      // TODO: Implement secure reply functionality
      console.log('Sending reply:', replyText);
      setShowReply(false);
      setReplyText('');
    } catch (err) {
      console.error('Error sending reply:', err);
    }
  };

  const downloadAttachment = async (attachment: Attachment) => {
    try {
      // TODO: Implement secure attachment download
      console.log('Downloading attachment:', attachment.name);
    } catch (err) {
      console.error('Error downloading attachment:', err);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <AlertTriangle className="mx-auto h-12 w-12 text-red-600" />
          <h2 className="mt-4 text-xl font-semibold text-gray-900">Error Loading Email</h2>
          <p className="mt-2 text-gray-600">{error}</p>
        </div>
      </div>
    );
  }

  if (!emailData) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Mail className="mx-auto h-12 w-12 text-gray-400" />
          <h2 className="mt-4 text-xl font-semibold text-gray-900">No Email Data</h2>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Security Banner */}
      <div className="bg-green-600 text-white px-4 py-3">
        <div className="flex items-center justify-center">
          <Shield className="h-5 w-5 mr-2" />
          <span className="text-sm font-medium">
            This email is secured with {emailData.securityInfo.encryptionType} encryption
          </span>
        </div>
      </div>

      {/* Email Content */}
      <div className="max-w-4xl mx-auto py-8 px-4">
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          {/* Email Header */}
          <div className="border-b border-gray-200 px-6 py-4">
            <h1 className="text-2xl font-bold text-gray-900">{emailData.subject}</h1>
            <div className="mt-2 flex items-center text-sm text-gray-600">
              <span>From: {emailData.senderName} &lt;{emailData.senderEmail}&gt;</span>
              <span className="mx-2">•</span>
              <span>{new Date(emailData.sentAt).toLocaleString()}</span>
            </div>
          </div>

          {/* Security Info */}
          {(emailData.securityInfo.readOnce || emailData.securityInfo.autoDestruct || emailData.securityInfo.expiresAt) && (
            <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4">
              <div className="flex">
                <AlertTriangle className="h-5 w-5 text-yellow-400" />
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-yellow-800">Security Notice</h3>
                  <div className="mt-2 text-sm text-yellow-700">
                    {emailData.securityInfo.readOnce && <p>• This email can only be viewed once</p>}
                    {emailData.securityInfo.autoDestruct && <p>• This email will self-destruct after viewing</p>}
                    {emailData.securityInfo.expiresAt && (
                      <p>• This email expires on {new Date(emailData.securityInfo.expiresAt).toLocaleString()}</p>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Email Body */}
          <div className="px-6 py-6">
            <div className="prose max-w-none">
              <div dangerouslySetInnerHTML={{ __html: emailData.body.replace(/\n/g, '<br />') }} />
            </div>
          </div>

          {/* Attachments */}
          {emailData.attachments.length > 0 && (
            <div className="border-t border-gray-200 px-6 py-4">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Attachments</h3>
              <div className="space-y-2">
                {emailData.attachments.map((attachment) => (
                  <div
                    key={attachment.id}
                    className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                  >
                    <div className="flex items-center">
                      <Download className="h-5 w-5 text-gray-400 mr-3" />
                      <div>
                        <p className="text-sm font-medium text-gray-900">{attachment.name}</p>
                        <p className="text-xs text-gray-500">
                          {(attachment.size / 1024 / 1024).toFixed(2)} MB
                        </p>
                      </div>
                    </div>
                    <button
                      onClick={() => downloadAttachment(attachment)}
                      className="bg-blue-600 text-white px-4 py-2 rounded-md text-sm hover:bg-blue-700"
                    >
                      Download
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Reply Section */}
          <div className="border-t border-gray-200 px-6 py-4">
            {!showReply ? (
              <button
                onClick={() => setShowReply(true)}
                className="flex items-center bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
              >
                <Reply className="h-4 w-4 mr-2" />
                Reply Securely
              </button>
            ) : (
              <div className="space-y-4">
                <h3 className="text-lg font-medium text-gray-900">Secure Reply</h3>
                <textarea
                  value={replyText}
                  onChange={(e) => setReplyText(e.target.value)}
                  placeholder="Type your secure reply here..."
                  className="w-full h-32 p-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
                <div className="flex space-x-3">
                  <button
                    onClick={handleReply}
                    className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
                  >
                    Send Reply
                  </button>
                  <button
                    onClick={() => setShowReply(false)}
                    className="bg-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-400"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SecureEmailViewer;
