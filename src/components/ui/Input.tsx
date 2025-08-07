import React from 'react';
import { cn } from '@/lib/utils';
import type { InputProps } from '@/types';

const Input: React.FC<InputProps> = ({
  className,
  type = 'text',
  placeholder,
  value,
  onChange,
  error,
  disabled = false,
  required = false,
  ...props
}) => {
  const baseClasses = 'block w-full rounded-lg border border-secondary-300 bg-white px-3 py-2 text-sm text-secondary-900 placeholder-secondary-500 focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white dark:placeholder-secondary-400 dark:focus:border-primary-500 dark:focus:ring-primary-500';

  const errorClasses = error ? 'border-error-500 focus:border-error-500 focus:ring-error-500' : '';

  const classes = cn(baseClasses, errorClasses, className);

  return (
    <div className="relative">
      <input
        type={type}
        className={classes}
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        disabled={disabled}
        required={required}
        {...props}
      />
      {error && (
        <p className="mt-1 text-sm text-error-600 dark:text-error-400">
          {error}
        </p>
      )}
    </div>
  );
};

export default Input; 