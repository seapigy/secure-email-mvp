import React, { useState, useRef, useEffect } from 'react';
import { 
  Bold, 
  Italic, 
  Underline, 
  List, 
  Link, 
  Image, 
  AlignLeft, 
  AlignCenter, 
  AlignRight,
  Quote,
  Code,
  Heading1,
  Heading2,
  Type,
  Palette
} from 'lucide-react';

interface RichTextEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  maxLength?: number;
  onFeaturesUsed?: (features: string[]) => void;
}

interface ToolbarButtonProps {
  icon: React.ComponentType<{ className?: string }>;
  onClick: () => void;
  isActive?: boolean;
  title: string;
  disabled?: boolean;
}

const ToolbarButton: React.FC<ToolbarButtonProps> = ({ 
  icon: Icon, 
  onClick, 
  isActive, 
  title, 
  disabled 
}) => (
  <button
    type="button"
    onClick={onClick}
    disabled={disabled}
    title={title}
    className={`
      p-2 rounded-md transition-colors
      ${isActive 
        ? 'bg-blue-100 text-blue-700 border border-blue-300' 
        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-800'
      }
      ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
    `}
  >
    <Icon className="h-4 w-4" />
  </button>
);

const RichTextEditor: React.FC<RichTextEditorProps> = ({
  value,
  onChange,
  placeholder = "Type your message here...",
  disabled = false,
  maxLength = 10000,
  onFeaturesUsed
}) => {
  const editorRef = useRef<HTMLDivElement>(null);
  const [isFocused, setIsFocused] = useState(false);
  const [featuresUsed, setFeaturesUsed] = useState<string[]>([]);

  // Initialize editor content
  useEffect(() => {
    if (editorRef.current && value !== editorRef.current.innerHTML) {
      editorRef.current.innerHTML = value;
    }
  }, [value]);

  // Track features used
  useEffect(() => {
    if (onFeaturesUsed) {
      onFeaturesUsed(featuresUsed);
    }
  }, [featuresUsed, onFeaturesUsed]);

  const execCommand = (command: string, value?: string) => {
    if (editorRef.current) {
      editorRef.current.focus();
      document.execCommand(command, false, value);
      updateContent();
      trackFeatures();
    }
  };

  const updateContent = () => {
    if (editorRef.current) {
      const content = editorRef.current.innerHTML;
      if (content !== value) {
        onChange(content);
      }
    }
  };

  const trackFeatures = () => {
    if (editorRef.current) {
      const features: string[] = [];
      const content = editorRef.current.innerHTML;

      // Check for bold text
      if (content.includes('<strong>') || content.includes('<b>')) {
        features.push('bold');
      }

      // Check for italic text
      if (content.includes('<em>') || content.includes('<i>')) {
        features.push('italic');
      }

      // Check for underlined text
      if (content.includes('<u>')) {
        features.push('underline');
      }

      // Check for lists
      if (content.includes('<ul>') || content.includes('<ol>')) {
        features.push('lists');
      }

      // Check for links
      if (content.includes('<a href=')) {
        features.push('links');
      }

      // Check for images
      if (content.includes('<img')) {
        features.push('images');
      }

      // Check for headings
      if (content.includes('<h1>') || content.includes('<h2>') || content.includes('<h3>')) {
        features.push('headings');
      }

      // Check for quotes
      if (content.includes('<blockquote>')) {
        features.push('quotes');
      }

      // Check for code
      if (content.includes('<code>') || content.includes('<pre>')) {
        features.push('code');
      }

      setFeaturesUsed(features);
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const text = e.clipboardData.getData('text/plain');
    document.execCommand('insertText', false, text);
    updateContent();
    trackFeatures();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Handle Ctrl+B for bold
    if (e.ctrlKey && e.key === 'b') {
      e.preventDefault();
      execCommand('bold');
    }
    // Handle Ctrl+I for italic
    if (e.ctrlKey && e.key === 'i') {
      e.preventDefault();
      execCommand('italic');
    }
    // Handle Ctrl+U for underline
    if (e.ctrlKey && e.key === 'u') {
      e.preventDefault();
      execCommand('underline');
    }
  };

  const insertLink = () => {
    const url = prompt('Enter URL:');
    if (url) {
      execCommand('createLink', url);
    }
  };

  const insertImage = () => {
    const url = prompt('Enter image URL:');
    if (url) {
      execCommand('insertImage', url);
    }
  };

  const insertQuote = () => {
    execCommand('formatBlock', '<blockquote>');
  };

  const insertCode = () => {
    execCommand('formatBlock', '<pre>');
  };

  const insertHeading = (level: number) => {
    execCommand('formatBlock', `<h${level}>`);
  };

  const changeColor = () => {
    const color = prompt('Enter color (e.g., #ff0000 or red):');
    if (color) {
      execCommand('foreColor', color);
    }
  };

  const changeFontSize = () => {
    const size = prompt('Enter font size (1-7):');
    if (size && /^[1-7]$/.test(size)) {
      execCommand('fontSize', size);
    }
  };

  return (
    <div className="border border-gray-300 rounded-lg overflow-hidden">
      {/* Toolbar */}
      <div className="bg-gray-50 border-b border-gray-300 p-2 flex flex-wrap gap-1">
        {/* Text formatting */}
        <div className="flex gap-1 mr-2">
          <ToolbarButton
            icon={Bold}
            onClick={() => execCommand('bold')}
            title="Bold (Ctrl+B)"
            disabled={disabled}
          />
          <ToolbarButton
            icon={Italic}
            onClick={() => execCommand('italic')}
            title="Italic (Ctrl+I)"
            disabled={disabled}
          />
          <ToolbarButton
            icon={Underline}
            onClick={() => execCommand('underline')}
            title="Underline (Ctrl+U)"
            disabled={disabled}
          />
        </div>

        {/* Alignment */}
        <div className="flex gap-1 mr-2">
          <ToolbarButton
            icon={AlignLeft}
            onClick={() => execCommand('justifyLeft')}
            title="Align Left"
            disabled={disabled}
          />
          <ToolbarButton
            icon={AlignCenter}
            onClick={() => execCommand('justifyCenter')}
            title="Align Center"
            disabled={disabled}
          />
          <ToolbarButton
            icon={AlignRight}
            onClick={() => execCommand('justifyRight')}
            title="Align Right"
            disabled={disabled}
          />
        </div>

        {/* Lists */}
        <div className="flex gap-1 mr-2">
          <ToolbarButton
            icon={List}
            onClick={() => execCommand('insertUnorderedList')}
            title="Bullet List"
            disabled={disabled}
          />
        </div>

        {/* Links and media */}
        <div className="flex gap-1 mr-2">
          <ToolbarButton
            icon={Link}
            onClick={insertLink}
            title="Insert Link"
            disabled={disabled}
          />
          <ToolbarButton
            icon={Image}
            onClick={insertImage}
            title="Insert Image"
            disabled={disabled}
          />
        </div>

        {/* Headings */}
        <div className="flex gap-1 mr-2">
          <ToolbarButton
            icon={Heading1}
            onClick={() => insertHeading(1)}
            title="Heading 1"
            disabled={disabled}
          />
          <ToolbarButton
            icon={Heading2}
            onClick={() => insertHeading(2)}
            title="Heading 2"
            disabled={disabled}
          />
        </div>

        {/* Special formatting */}
        <div className="flex gap-1 mr-2">
          <ToolbarButton
            icon={Quote}
            onClick={insertQuote}
            title="Insert Quote"
            disabled={disabled}
          />
          <ToolbarButton
            icon={Code}
            onClick={insertCode}
            title="Insert Code Block"
            disabled={disabled}
          />
        </div>

        {/* Styling */}
        <div className="flex gap-1">
          <ToolbarButton
            icon={Type}
            onClick={changeFontSize}
            title="Change Font Size"
            disabled={disabled}
          />
          <ToolbarButton
            icon={Palette}
            onClick={changeColor}
            title="Change Color"
            disabled={disabled}
          />
        </div>
      </div>

      {/* Editor */}
      <div
        ref={editorRef}
        contentEditable={!disabled}
        onInput={updateContent}
        onBlur={() => setIsFocused(false)}
        onFocus={() => setIsFocused(true)}
        onPaste={handlePaste}
        onKeyDown={handleKeyDown}
        className={`
          min-h-[200px] max-h-[400px] p-4 overflow-y-auto
          focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset
          ${disabled ? 'bg-gray-100 cursor-not-allowed' : 'bg-white'}
          ${!value && !isFocused ? 'text-gray-400' : 'text-gray-900'}
        `}
        style={{ 
          minHeight: '200px',
          maxHeight: '400px'
        }}
        data-placeholder={placeholder}
      />

      {/* Character count */}
      {maxLength && (
        <div className="bg-gray-50 border-t border-gray-300 px-4 py-2 text-sm text-gray-500">
          {value.length} / {maxLength} characters
          {value.length > maxLength * 0.9 && (
            <span className="text-orange-600 ml-2">
              (Approaching limit)
            </span>
          )}
        </div>
      )}

      {/* Features used indicator */}
      {featuresUsed.length > 0 && (
        <div className="bg-blue-50 border-t border-blue-200 px-4 py-2 text-sm text-blue-700">
          <span className="font-medium">Features used:</span> {featuresUsed.join(', ')}
        </div>
      )}
    </div>
  );
};

export default RichTextEditor;
