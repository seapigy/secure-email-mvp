import React, { useState, useEffect } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';

interface DailyDigestDelivery {
  delivery_id: string;
  user_id: string;
  digest_date: string;
  email_sent: boolean;
  sms_sent: boolean;
  email_sent_at?: string;
  sms_sent_at?: string;
  event_count: number;
  email_count: number;
  success_count: number;
  failure_count: number;
  blocked_count: number;
  suppression_count: number;
  created_at: string;
}

interface DigestSummary {
  user_id: string;
  digest_date: string;
  total_events: number;
  total_emails: number;
  success_count: number;
  failure_count: number;
  blocked_count: number;
  suppression_count: number;
  email_summaries: Array<{
    email_id: string;
    email_subject: string;
    recipient: string;
    success_count: number;
    failure_count: number;
    blocked_count: number;
    last_access_at?: string;
    last_ip_address?: string;
    last_device_type?: string;
    last_country?: string;
    last_city?: string;
    suppression_count: number;
  }>;
}

const DigestHistory: React.FC = () => {
  const [digestHistory, setDigestHistory] = useState<DailyDigestDelivery[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedDate, setSelectedDate] = useState<string>('');
  const [generatedDigest, setGeneratedDigest] = useState<DigestSummary | null>(null);
  const [generating, setGenerating] = useState(false);

  const fetchDigestHistory = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch('/api/notifications/digest/history?limit=20', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch digest history');
      }

      const data = await response.json();
      setDigestHistory(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const generateDigest = async () => {
    if (!selectedDate) {
      setError('Please select a date');
      return;
    }

    setGenerating(true);
    setError(null);
    
    try {
      const response = await fetch('/api/notifications/digest/generate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
        body: JSON.stringify({ date: selectedDate }),
      });

      if (!response.ok) {
        throw new Error('Failed to generate digest');
      }

      const data = await response.json();
      setGeneratedDigest(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setGenerating(false);
    }
  };

  useEffect(() => {
    fetchDigestHistory();
  }, []);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="space-y-6">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
          Daily Digest History
        </h2>
        
        <div className="mb-6">
          <Button
            onClick={fetchDigestHistory}
            disabled={loading}
            className="mr-4"
          >
            {loading ? 'Loading...' : 'Refresh History'}
          </Button>
        </div>

        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 mb-4">
            <p className="text-red-800 dark:text-red-200">{error}</p>
          </div>
        )}

        {digestHistory.length === 0 && !loading ? (
          <p className="text-gray-500 dark:text-gray-400">No digest history found.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Date
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Events
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Emails
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Success
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Failed
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Blocked
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Suppressed
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Delivery Status
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {digestHistory.map((digest) => (
                  <tr key={digest.delivery_id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      {formatDate(digest.digest_date)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      {digest.event_count}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      {digest.email_count}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-green-600 dark:text-green-400">
                      {digest.success_count}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-red-600 dark:text-red-400">
                      {digest.failure_count}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-orange-600 dark:text-orange-400">
                      {digest.blocked_count}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                      {digest.suppression_count}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      <div className="flex flex-col space-y-1">
                        {digest.email_sent && (
                          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                            Email ✓
                          </span>
                        )}
                        {digest.sms_sent && (
                          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                            SMS ✓
                          </span>
                        )}
                        {!digest.email_sent && !digest.sms_sent && (
                          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200">
                            No delivery
                          </span>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Generate Digest Section */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h3 className="text-xl font-bold text-gray-900 dark:text-white mb-4">
          Generate Digest
        </h3>
        
        <div className="flex items-center space-x-4 mb-4">
          <Input
            type="date"
            value={selectedDate}
            onChange={(e) => setSelectedDate(e.target.value)}
            placeholder="Select date (YYYY-MM-DD)"
            className="w-48"
          />
          <Button
            onClick={generateDigest}
            disabled={generating || !selectedDate}
          >
            {generating ? 'Generating...' : 'Generate Digest'}
          </Button>
        </div>

        {generatedDigest && (
          <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
            <h4 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
              Digest Summary for {formatDate(generatedDigest.digest_date)}
            </h4>
            
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
              <div className="text-center">
                <div className="text-2xl font-bold text-gray-900 dark:text-white">
                  {generatedDigest.total_events}
                </div>
                <div className="text-sm text-gray-500 dark:text-gray-400">Total Events</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-gray-900 dark:text-white">
                  {generatedDigest.total_emails}
                </div>
                <div className="text-sm text-gray-500 dark:text-gray-400">Emails</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                  {generatedDigest.success_count}
                </div>
                <div className="text-sm text-gray-500 dark:text-gray-400">Success</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-red-600 dark:text-red-400">
                  {generatedDigest.failure_count}
                </div>
                <div className="text-sm text-gray-500 dark:text-gray-400">Failed</div>
              </div>
            </div>

            {generatedDigest.email_summaries.length > 0 && (
              <div>
                <h5 className="font-semibold text-gray-900 dark:text-white mb-2">
                  Email Details
                </h5>
                <div className="space-y-2">
                  {generatedDigest.email_summaries.map((email) => (
                    <div key={email.email_id} className="bg-white dark:bg-gray-800 rounded p-3">
                      <div className="font-medium text-gray-900 dark:text-white">
                        {email.email_subject}
                      </div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">
                        To: {email.recipient}
                      </div>
                      <div className="flex space-x-4 text-sm mt-1">
                        <span className="text-green-600 dark:text-green-400">
                          ✓ {email.success_count} success
                        </span>
                        <span className="text-red-600 dark:text-red-400">
                          ✗ {email.failure_count} failed
                        </span>
                        <span className="text-orange-600 dark:text-orange-400">
                          🚫 {email.blocked_count} blocked
                        </span>
                        {email.suppression_count > 0 && (
                          <span className="text-gray-600 dark:text-gray-400">
                            🔇 {email.suppression_count} suppressed
                          </span>
                        )}
                      </div>
                      {email.last_access_at && (
                        <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                          Last access: {formatDateTime(email.last_access_at)}
                          {email.last_country && ` from ${email.last_country}`}
                          {email.last_city && `, ${email.last_city}`}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default DigestHistory;
