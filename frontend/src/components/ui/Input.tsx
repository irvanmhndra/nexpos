import { forwardRef, useState, useId } from 'react';
import type { InputHTMLAttributes } from 'react';
import { cn } from '@/utils';
import { Icon } from './Icon';

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  fullWidth?: boolean;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  (
    {
      className,
      label,
      error,
      helperText,
      leftIcon,
      rightIcon,
      fullWidth = false,
      id,
      type,
      ...props
    },
    ref
  ) => {
    const [showPassword, setShowPassword] = useState(false);
    const generatedId = useId();
    const inputId = id || generatedId;

    const isPassword = type === 'password';
    const inputType = isPassword && showPassword ? 'text' : type;

    const togglePassword = () => setShowPassword(!showPassword);

    return (
      <div className={cn('flex flex-col gap-2', fullWidth && 'w-full')}>
        {label && (
          <label
            htmlFor={inputId}
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

          <input
            ref={ref}
            id={inputId}
            type={inputType}
            className={cn(
              'w-full h-12 rounded-lg border bg-white dark:bg-card-dark text-text-main dark:text-white placeholder:text-text-sub dark:placeholder:text-gray-500',
              'focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary',
              'transition-all',
              error
                ? 'border-red-500 focus:ring-red-500 focus:border-red-500'
                : 'border-border-light dark:border-border-dark',
              leftIcon ? 'pl-10' : 'pl-4',
              rightIcon || isPassword ? 'pr-10' : 'pr-4',
              props.disabled && 'opacity-50 cursor-not-allowed bg-gray-50 dark:bg-gray-800',
              className
            )}
            {...props}
          />

          {(rightIcon || isPassword) && (
            <div className="absolute inset-y-0 right-0 flex items-center pr-3 text-text-sub dark:text-gray-400">
              {rightIcon}
              {isPassword && !rightIcon && (
                <button
                  type="button"
                  onClick={togglePassword}
                  className="focus:outline-none hover:text-text-main dark:hover:text-white transition-colors"
                  tabIndex={-1}
                >
                  <Icon name={showPassword ? 'visibility' : 'visibility_off'} size={20} />
                </button>
              )}
            </div>
          )}
        </div>

        {error && <p className="text-red-500 text-xs font-medium">{error}</p>}
        {helperText && !error && (
          <p className="text-text-sub dark:text-gray-400 text-xs">{helperText}</p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';
