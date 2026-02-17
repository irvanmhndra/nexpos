import { forwardRef } from 'react';
import type { InputHTMLAttributes } from 'react';
import { cn } from '@/utils';

export interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string;
  error?: string;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ className, label, error, id, ...props }, ref) => {
    const checkboxId = id || `checkbox-${Math.random().toString(36).substr(2, 9)}`;

    return (
      <div className="flex flex-col gap-1">
        <label htmlFor={checkboxId} className="flex items-center gap-3 cursor-pointer group">
          <div className="relative flex items-center">
            <input
              ref={ref}
              type="checkbox"
              id={checkboxId}
              className={cn(
                'peer size-5 cursor-pointer appearance-none rounded border bg-white dark:bg-card-dark',
                'checked:bg-primary checked:border-primary',
                'focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2',
                'transition-all',
                error
                  ? 'border-red-500'
                  : 'border-border-light dark:border-border-dark',
                props.disabled && 'opacity-50 cursor-not-allowed',
                className
              )}
              {...props}
            />
            <span className="material-symbols-outlined absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-white opacity-0 peer-checked:opacity-100 text-[16px] pointer-events-none">
              check
            </span>
          </div>

          {label && (
            <span className="text-sm font-medium text-text-main dark:text-gray-300 group-hover:text-primary transition-colors">
              {label}
            </span>
          )}
        </label>

        {error && <p className="text-red-500 text-xs font-medium ml-8">{error}</p>}
      </div>
    );
  }
);

Checkbox.displayName = 'Checkbox';
