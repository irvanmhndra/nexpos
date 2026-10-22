import { useState, useRef, useEffect } from 'react';
import { cn } from '@/utils';

interface SelectOption {
  value: string;
  label: string;
}

interface CustomSelectProps {
  value: string;
  onChange: (value: string) => void;
  options: SelectOption[];
  placeholder?: string;
  className?: string;
  leftIcon?: React.ReactNode;
  disabled?: boolean;
}

export function Dropdown({
  value,
  onChange,
  options,
  placeholder = 'Pilih...',
  className,
  leftIcon,
  disabled = false,
}: CustomSelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const selectRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (selectRef.current && !selectRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [isOpen]);

  const selectedOption = options.find(opt => opt.value === value);

  const handleSelect = (optionValue: string) => {
    onChange(optionValue);
    setIsOpen(false);
  };

  return (
    <div ref={selectRef} className={cn('relative', className)}>
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={cn(
          'w-full h-12 rounded-lg border bg-white dark:bg-card-dark text-text-main dark:text-white text-sm font-medium',
          'focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary',
          'hover:border-primary/50 dark:hover:border-primary/50',
          'transition-all shadow-sm flex items-center justify-between',
          'border-border-light dark:border-border-dark',
          leftIcon ? 'pl-10' : 'pl-4',
          'pr-10',
          disabled && 'opacity-50 cursor-not-allowed bg-gray-50 dark:bg-gray-800'
        )}
      >
        {leftIcon && (
          <div className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-sub dark:text-gray-400">
            {leftIcon}
          </div>
        )}
        <span className={cn(!selectedOption && 'text-text-sub dark:text-gray-400')}>
          {selectedOption?.label || placeholder}
        </span>
        <span
          className={cn(
            'material-symbols-outlined text-[20px] text-text-sub dark:text-gray-400 transition-transform',
            isOpen && 'rotate-180'
          )}
        >
          expand_more
        </span>
      </button>

      {isOpen && !disabled && (
        <div className="absolute z-50 w-full mt-1 bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark rounded-lg shadow-lg max-h-60 overflow-auto animate-in fade-in slide-in-from-top-2 duration-200">
          {options.map(option => (
            <button
              key={option.value}
              onClick={() => handleSelect(option.value)}
              className={cn(
                'w-full px-4 py-3 text-sm text-left transition-colors',
                'hover:bg-background-light dark:hover:bg-background-dark',
                option.value === value
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-text-main dark:text-white'
              )}
            >
              {option.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
