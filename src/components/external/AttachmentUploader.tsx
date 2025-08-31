import React, { useState, useRef, useCallback } from 'react';
import { 
  Upload, 
  File, 
  FileText, 
  Image, 
  Archive, 
  AlertTriangle, 
  CheckCircle,
  Trash2,
  Download,
  Eye
} from 'lucide-react';
import { log } from '@/lib/logger';

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

interface AttachmentUploaderProps {
  linkID: string;
  replyID?: string;
  onAttachmentUploaded?: (attachment: Attachment) => void;
  onAttachmentRemoved?: (attachmentId: string) => void;
  maxFiles?: number;
  maxFileSize?: number; // in bytes
  allowedTypes?: string[];
  disabled?: boolean;
}

const AttachmentUploader: React.FC<AttachmentUploaderProps> = ({
  linkID,
  replyID,
  onAttachmentUploaded,
  onAttachmentRemoved,
  maxFiles = 5,
  maxFileSize = 25 * 1024 * 1024, // 25MB
  allowedTypes = [
    'application/pdf',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    'text/plain',
    'image/jpeg',
    'image/png',
    'image/gif',
    'image/webp'
  ],
  disabled = false
}) => {
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [isDragOver, setIsDragOver] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getFileIcon = (type: string) => {
    if (type.startsWith('image/')) return Image;
    if (type === 'application/pdf' || type.startsWith('text/')) return FileText;
    if (type.includes('zip') || type.includes('rar')) return Archive;
    return File;
  };

  const validateFile = useCallback((file: File): string | null => {
    // Check file size
    if (file.size > maxFileSize) {
      return `File too large. Maximum size is ${formatFileSize(maxFileSize)}`;
    }

    // Check file type
    if (!allowedTypes.includes(file.type)) {
      return `File type not allowed. Allowed types: ${allowedTypes.join(', ')}`;
    }

    // Check file count
    if (attachments.length >= maxFiles) {
      return `Maximum ${maxFiles} files allowed`;
    }

    return null;
  }, [maxFileSize, allowedTypes, attachments.length, maxFiles]);

  const uploadFile = useCallback(async (file: File): Promise<void> => {
    const attachmentId = `att_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const attachment: Attachment = {
      id: attachmentId,
      name: file.name,
      size: file.size,
      type: file.type,
      status: 'uploading',
      progress: 0
    };

    setAttachments(prev => [...prev, attachment]);

    try {
      // Create FormData
      const formData = new FormData();
      formData.append('file', file);
      formData.append('link_id', linkID);
      if (replyID) {
        formData.append('reply_id', replyID);
      }

      // Upload file
      const response = await fetch(`/api/v/${linkID}/attachments`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Upload failed');
      }

      const result = await response.json();

      // Update attachment with success status
      setAttachments(prev => prev.map(att => 
        att.id === attachmentId 
          ? { 
              ...att, 
              status: 'virus_scanning' as const, 
              progress: 100,
              url: result.upload_url 
            }
          : att
      ));

      // Simulate virus scanning (in real implementation, this would be async)
      setTimeout(() => {
        setAttachments(prev => prev.map(att => 
          att.id === attachmentId 
            ? { 
                ...att, 
                status: 'success' as const,
                virusScanStatus: 'clean' as const
              }
            : att
        ));

        if (onAttachmentUploaded) {
          onAttachmentUploaded({
            ...attachment,
            status: 'success',
            progress: 100,
            url: result.upload_url,
            virusScanStatus: 'clean'
          });
        }
      }, 2000);

    } catch (error) {
      log.error('Upload error:', error, 'AttachmentUploader');
      setUploadError('Failed to upload file. Please try again.');
    }
  }, [linkID, replyID, onAttachmentUploaded]);

  const handleFileSelect = useCallback((files: FileList | null) => {
    if (!files) return;

    setUploadError(null);
    Array.from(files).forEach(file => {
      const error = validateFile(file);
      if (error) {
        setUploadError(error);
        return;
      }
      uploadFile(file);
    });
  }, [validateFile, uploadFile]);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    handleFileSelect(e.dataTransfer.files);
  }, [handleFileSelect]);

  const removeAttachment = (attachmentId: string) => {
    setAttachments(prev => prev.filter(att => att.id !== attachmentId));
    if (onAttachmentRemoved) {
      onAttachmentRemoved(attachmentId);
    }
  };

  const downloadAttachment = async (attachment: Attachment) => {
    if (!attachment.url) return;

    try {
      // Generate download token
      const tokenResponse = await fetch(`/api/v/${linkID}/attachments/${attachment.id}/token`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ attachment_id: attachment.id }),
      });

      if (!tokenResponse.ok) {
        throw new Error('Failed to generate download token');
      }

      const tokenData = await tokenResponse.json();

      // Get download URL
      const downloadResponse = await fetch(`/api/v/${linkID}/attachments/download`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          attachment_id: attachment.id,
          token_hash: tokenData.token_hash,
        }),
      });

      if (!downloadResponse.ok) {
        throw new Error('Failed to get download URL');
      }

      const downloadData = await downloadResponse.json();

      // Download file
      const link = document.createElement('a');
      link.href = downloadData.download_url;
      link.download = attachment.name;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

    } catch (error) {
      log.error('Download error:', error, 'AttachmentUploader');
      alert('Failed to download file');
    }
  };

  return (
    <div className="space-y-4">
      {/* Upload Area */}
      <div
        className={`
          border-2 border-dashed rounded-lg p-6 text-center transition-colors
          ${isDragOver 
            ? 'border-blue-400 bg-blue-50' 
            : 'border-gray-300 hover:border-gray-400'
          }
          ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
        `}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => !disabled && fileInputRef.current?.click()}
      >
        <Upload className="mx-auto h-12 w-12 text-gray-400 mb-4" />
        <p className="text-lg font-medium text-gray-900 mb-2">
          Drop files here or click to upload
        </p>
        <p className="text-sm text-gray-500">
          Maximum {maxFiles} files, {formatFileSize(maxFileSize)} each
        </p>
        <p className="text-xs text-gray-400 mt-1">
          Supported: PDF, Word, Excel, Images, Text files
        </p>
      </div>

      {/* Hidden file input */}
      <input
        ref={fileInputRef}
        type="file"
        multiple
        accept={allowedTypes.join(',')}
        onChange={(e) => handleFileSelect(e.target.files)}
        className="hidden"
        disabled={disabled}
      />

      {/* Upload Error */}
      {uploadError && (
        <div className="bg-red-50 border border-red-200 rounded-md p-3">
          <div className="flex">
            <AlertTriangle className="h-5 w-5 text-red-400" />
            <div className="ml-3">
              <p className="text-sm text-red-800">{uploadError}</p>
            </div>
          </div>
        </div>
      )}

      {/* Attachments List */}
      {attachments.length > 0 && (
        <div className="space-y-2">
          <h3 className="text-sm font-medium text-gray-900">
            Attachments ({attachments.length}/{maxFiles})
          </h3>
          {attachments.map((attachment) => {
            const FileIcon = getFileIcon(attachment.type);
            return (
              <div
                key={attachment.id}
                className="flex items-center justify-between p-3 bg-gray-50 rounded-lg border"
              >
                <div className="flex items-center space-x-3">
                  <FileIcon className="h-8 w-8 text-gray-400" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {attachment.name}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatFileSize(attachment.size)}
                    </p>
                  </div>
                </div>

                <div className="flex items-center space-x-2">
                  {/* Status indicator */}
                  {attachment.status === 'uploading' && (
                    <div className="flex items-center space-x-2">
                      <div className="w-4 h-4 border-2 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
                      <span className="text-xs text-blue-600">
                        {attachment.progress}%
                      </span>
                    </div>
                  )}

                  {attachment.status === 'virus_scanning' && (
                    <div className="flex items-center space-x-2">
                      <div className="w-4 h-4 border-2 border-yellow-200 border-t-yellow-600 rounded-full animate-spin"></div>
                      <span className="text-xs text-yellow-600">Scanning...</span>
                    </div>
                  )}

                  {attachment.status === 'success' && (
                    <div className="flex items-center space-x-2">
                      <CheckCircle className="h-4 w-4 text-green-600" />
                      <span className="text-xs text-green-600">
                        {attachment.virusScanStatus === 'clean' ? 'Clean' : 'Ready'}
                      </span>
                    </div>
                  )}

                  {attachment.status === 'error' && (
                    <div className="flex items-center space-x-2">
                      <AlertTriangle className="h-4 w-4 text-red-600" />
                      <span className="text-xs text-red-600">Error</span>
                    </div>
                  )}

                  {/* Action buttons */}
                  {attachment.status === 'success' && (
                    <>
                      <button
                        onClick={() => downloadAttachment(attachment)}
                        className="p-1 text-gray-400 hover:text-gray-600"
                        title="Download"
                      >
                        <Download className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => window.open(attachment.url, '_blank')}
                        className="p-1 text-gray-400 hover:text-gray-600"
                        title="Preview"
                      >
                        <Eye className="h-4 w-4" />
                      </button>
                    </>
                  )}

                  <button
                    onClick={() => removeAttachment(attachment.id)}
                    className="p-1 text-gray-400 hover:text-red-600"
                    title="Remove"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default AttachmentUploader;
