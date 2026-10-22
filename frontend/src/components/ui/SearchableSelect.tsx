import { useState, useRef, useEffect } from 'react';
import { cn } from '@/utils';
import { Icon } from './Icon';

export interface SearchableSelectOption {
  value: string;
  label: string;
  description?: string;
}

export interface SearchableSelectProps {
  options: SearchableSelectOption[];
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  searchPlaceholder?: string;
  disabled?: boolean;
  error?: string;
  className?: string;
}

export function SearchableSelect({
  options,
  value,
  onChange,
  placeholder = 'Pilih...',
  searchPlaceholder = 'Cari...',
  disabled = false,
  error,
  className,
}: SearchableSelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const selectedOption = options.find(opt => opt.value === value);

  const filteredOptions = options.filter(
    opt =>
      opt.label.toLowerCase().includes(search.toLowerCase()) ||
      opt.description?.toLowerCase().includes(search.toLowerCase())
  );

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
        setSearch('');
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  const handleSelect = (optionValue: string) => {
    onChange(optionValue);
    setIsOpen(false);
    setSearch('');
  };

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <button
        type="button"
        disabled={disabled}
        onClick={() => !disabled && setIsOpen(!isOpen)}
        className={cn(
          'w-full h-12 rounded-lg border bg-white dark:bg-card-dark text-left px-4 flex items-center justify-between',
          'focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary',
          'hover:border-primary/50 dark:hover:border-primary/50 transition-all',
          error
            ? 'border-red-500 focus:ring-red-500/20 focus:border-red-500'
            : 'border-border-light dark:border-border-dark',
          disabled && 'opacity-50 cursor-not-allowed bg-gray-50 dark:bg-gray-800'
        )}
      >
        <span
          className={cn(
            'truncate text-sm',
            selectedOption ? 'text-text-main dark:text-white' : 'text-text-sub dark:text-gray-400'
          )}
        >
          {selectedOption ? selectedOption.label : placeholder}
        </span>
        <Icon
          name={isOpen ? 'expand_less' : 'expand_more'}
          size={20}
          className="text-text-sub dark:text-gray-400 flex-shrink-0"
        />
      </button>

      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-white dark:bg-card-dark border border-border-light dark:border-border-dark rounded-lg shadow-lg overflow-hidden">
          <div className="p-2 border-b border-border-light dark:border-border-dark">
            <div className="relative">
              <Icon
                name="search"
                size={18}
                className="absolute left-3 top-1/2 -translate-y-1/2 text-text-sub dark:text-gray-400"
              />
              <input
                ref={inputRef}
                type="text"
                value={search}
                onChange={e => setSearch(e.target.value)}
                placeholder={searchPlaceholder}
                className="w-full h-10 pl-9 pr-4 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-main dark:text-white text-sm focus:outline-none focus:border-primary"
              />
            </div>
          </div>

          <div className="max-h-60 overflow-y-auto">
            {filteredOptions.length === 0 ? (
              <div className="px-4 py-3 text-sm text-text-sub dark:text-gray-400 text-center">
                Tidak ada hasil
              </div>
            ) : (
              filteredOptions.map(option => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => handleSelect(option.value)}
                  className={cn(
                    'w-full px-4 py-3 text-left hover:bg-background-light dark:hover:bg-background-dark transition-colors',
                    option.value === value && 'bg-primary/5'
                  )}
                >
                  <div className="text-sm font-medium text-text-main dark:text-white">
                    {option.label}
                  </div>
                  {option.description && (
                    <div className="text-xs text-text-sub dark:text-gray-400 mt-0.5">
                      {option.description}
                    </div>
                  )}
                </button>
              ))
            )}
          </div>
        </div>
      )}

      {error && <p className="text-red-500 text-xs font-medium mt-1">{error}</p>}
    </div>
  );
}
