import { useState, useRef, useEffect } from 'react';
import { cn } from '@/utils';
import { Icon } from './Icon';

export interface DateRangePickerProps {
  startDate: string;
  endDate: string;
  onChange: (startDate: string, endDate: string) => void;
  placeholder?: string;
  className?: string;
}

export function DateRangePicker({
  startDate,
  endDate,
  onChange,
  placeholder = 'Pilih tanggal',
  className,
}: DateRangePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [tempStartDate, setTempStartDate] = useState(startDate);
  const [tempEndDate, setTempEndDate] = useState(endDate);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setTempStartDate(startDate);
     
    setTempEndDate(endDate);
  }, [startDate, endDate]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const formatDisplayDate = (date: string) => {
    if (!date) return '';
    const d = new Date(date);
    return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
  };

  const getDisplayValue = () => {
    if (startDate && endDate) {
      return `${formatDisplayDate(startDate)} - ${formatDisplayDate(endDate)}`;
    }
    if (startDate) {
      return `Dari ${formatDisplayDate(startDate)}`;
    }
    return '';
  };

  const handleApply = () => {
    onChange(tempStartDate, tempEndDate);
    setIsOpen(false);
  };

  const handleClear = () => {
    setTempStartDate('');
    setTempEndDate('');
    onChange('', '');
    setIsOpen(false);
  };

  const handleQuickSelect = (days: number) => {
    const end = new Date();
    const start = new Date();
    start.setDate(end.getDate() - days);

    const formatDate = (d: Date) => d.toISOString().split('T')[0];
    setTempStartDate(formatDate(start));
    setTempEndDate(formatDate(end));
  };

  const displayValue = getDisplayValue();

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'w-full h-12 rounded-lg border bg-white dark:bg-card-dark text-left px-4 flex items-center gap-3',
          'focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary',
          'hover:border-primary/50 dark:hover:border-primary/50 transition-all',
          'border-border-light dark:border-border-dark'
        )}
      >
        <Icon name="calendar_today" size={20} className="text-text-sub dark:text-gray-400 flex-shrink-0" />
        <span
          className={cn(
            'truncate text-sm flex-1',
            displayValue ? 'text-text-main dark:text-white' : 'text-text-sub dark:text-gray-400'
          )}
        >
          {displayValue || placeholder}
        </span>
        {displayValue && (
          // span (not button) — this sits inside the trigger button; nested
          // buttons are invalid HTML and break hydration.
          <span
            role="button"
            tabIndex={0}
            aria-label="Hapus rentang tanggal"
            onClick={(e) => {
              e.stopPropagation();
              handleClear();
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                e.stopPropagation();
                handleClear();
              }
            }}
            className="text-text-sub hover:text-text-main dark:hover:text-white cursor-pointer"
          >
            <Icon name="close" size={18} />
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute z-50 mt-1 bg-white dark:bg-card-dark border border-border-light dark:border-border-dark rounded-lg shadow-lg p-4 min-w-[320px] right-0">
          {/* Quick Select */}
          <div className="flex flex-wrap gap-2 mb-4">
            <button
              type="button"
              onClick={() => handleQuickSelect(0)}
              className="px-3 py-1.5 text-xs font-medium rounded-lg bg-background-light dark:bg-background-dark hover:bg-primary/10 hover:text-primary transition-colors"
            >
              Hari Ini
            </button>
            <button
              type="button"
              onClick={() => handleQuickSelect(7)}
              className="px-3 py-1.5 text-xs font-medium rounded-lg bg-background-light dark:bg-background-dark hover:bg-primary/10 hover:text-primary transition-colors"
            >
              7 Hari
            </button>
            <button
              type="button"
              onClick={() => handleQuickSelect(30)}
              className="px-3 py-1.5 text-xs font-medium rounded-lg bg-background-light dark:bg-background-dark hover:bg-primary/10 hover:text-primary transition-colors"
            >
              30 Hari
            </button>
            <button
              type="button"
              onClick={() => handleQuickSelect(90)}
              className="px-3 py-1.5 text-xs font-medium rounded-lg bg-background-light dark:bg-background-dark hover:bg-primary/10 hover:text-primary transition-colors"
            >
              3 Bulan
            </button>
          </div>

          {/* Date Inputs */}
          <div className="grid grid-cols-2 gap-3 mb-4">
            <div>
              <label className="block text-xs font-medium text-text-sub dark:text-gray-400 mb-1">
                Dari
              </label>
              <input
                type="date"
                value={tempStartDate}
                onChange={(e) => setTempStartDate(e.target.value)}
                className="w-full h-10 px-3 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-white text-sm focus:outline-none focus:border-primary"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-text-sub dark:text-gray-400 mb-1">
                Sampai
              </label>
              <input
                type="date"
                value={tempEndDate}
                onChange={(e) => setTempEndDate(e.target.value)}
                min={tempStartDate}
                className="w-full h-10 px-3 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-white text-sm focus:outline-none focus:border-primary"
              />
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-2 justify-end">
            <button
              type="button"
              onClick={handleClear}
              className="px-4 py-2 text-sm font-medium text-text-sub hover:text-text-main dark:hover:text-white transition-colors"
            >
              Reset
            </button>
            <button
              type="button"
              onClick={handleApply}
              className="px-4 py-2 text-sm font-medium bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors"
            >
              Terapkan
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
