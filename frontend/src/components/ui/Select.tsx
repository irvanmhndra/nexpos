import { forwardRef, useId } from 'react';
import type { SelectHTMLAttributes } from 'react';
import { cn } from '@/utils';

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  error?: string;
  helperText?: string;
  leftIcon?: React.ReactNode;
  fullWidth?: boolean;
  options?: Array<{ value: string; label: string }>;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  (
    {
      className,
      label,
      error,
      helperText,
      leftIcon,
      fullWidth = false,
      options = [],
      children,
      id,
      ...props
    },
    ref
  ) => {
    const generatedId = useId();
    const selectId = id || generatedId;

    return (
      <div className={cn('flex flex-col gap-2', fullWidth && 'w-full')}>
        {label && (
          <label
            htmlFor={selectId}
            className="text-text-main dark:text-gray-200 text-sm font-semibold"
          >
            {label}
          </label>
        )}

        <div className="relative">
          {leftIcon && (
            <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-text-sub dark:text-gray-400">
              {leftIcon}
            </div>
          )}

          <select
            ref={ref}
            id={selectId}
            className={cn(
              'w-full h-12 rounded-lg border bg-white dark:bg-card-dark text-text-main dark:text-white text-sm font-medium',
              'focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary',
              'hover:border-primary/50 dark:hover:border-primary/50',
              'transition-all appearance-none cursor-pointer shadow-sm',
              error
                ? 'border-red-500 focus:ring-red-500/20 focus:border-red-500'
                : 'border-border-light dark:border-border-dark',
              leftIcon ? 'pl-10' : 'pl-4',
              'pr-10',
              props.disabled && 'opacity-50 cursor-not-allowed bg-gray-50 dark:bg-gray-800',
              className
            )}
            {...props}
          >
            {children ||
              options.map(option => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
          </select>

          <div className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-sub dark:text-gray-400">
            <span className="material-symbols-outlined text-[20px]">expand_more</span>
          </div>
        </div>

        {error && <p className="text-red-500 text-xs font-medium">{error}</p>}
        {helperText && !error && (
          <p className="text-text-sub dark:text-gray-400 text-xs">{helperText}</p>
        )}
      </div>
    );
  }
);

Select.displayName = 'Select';
