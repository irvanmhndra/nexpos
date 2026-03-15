import { forwardRef, useId } from 'react';
import type { TextareaHTMLAttributes } from 'react';
import { cn } from '@/utils';

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
  helperText?: string;
  fullWidth?: boolean;
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, label, error, helperText, fullWidth = false, id, ...props }, ref) => {
    const generatedId = useId();
    const textareaId = id || generatedId;

    return (
      <div className={cn('flex flex-col gap-2', fullWidth && 'w-full')}>
        {label && (
          <label
            htmlFor={textareaId}
            className="text-text-main dark:text-gray-200 text-sm font-semibold"
          >
            {label}
          </label>
        )}

        <textarea
          ref={ref}
          id={textareaId}
          className={cn(
            'w-full min-h-[100px] p-4 rounded-lg border bg-white dark:bg-card-dark text-text-main dark:text-white placeholder:text-text-sub dark:placeholder:text-gray-500',
            'focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary',
            'transition-all resize-y',
            error
              ? 'border-red-500 focus:ring-red-500 focus:border-red-500'
              : 'border-border-light dark:border-border-dark',
            props.disabled && 'opacity-50 cursor-not-allowed bg-gray-50 dark:bg-gray-800',
            className
          )}
          {...props}
        />

        {error && <p className="text-red-500 text-xs font-medium">{error}</p>}
        {helperText && !error && (
          <p className="text-text-sub dark:text-gray-400 text-xs">{helperText}</p>
        )}
      </div>
    );
  }
);

Textarea.displayName = 'Textarea';
