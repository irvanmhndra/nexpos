import type { ReactNode } from 'react';
import { cn } from '@/utils';
import { Icon } from '@/components/ui';

interface Column<T> {
  header: string;
  accessor: keyof T | ((row: T) => ReactNode);
  className?: string;
}

interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  onRowClick?: (row: T) => void;
  loading?: boolean;
  emptyMessage?: string;
  hoverable?: boolean;
}

export function DataTable<T extends { id: string | number }>({
  data,
  columns,
  onRowClick,
  loading,
  emptyMessage = 'Tidak ada data',
  hoverable = true,
}: DataTableProps<T>) {
  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex flex-col items-center gap-3">
          <Icon name="progress_activity" size={32} className="animate-spin text-primary" />
          <p className="text-text-sub dark:text-gray-400 text-sm">Loading...</p>
        </div>
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-text-sub dark:text-gray-400">
        <div className="text-center">
          <Icon name="inbox" size={48} className="mb-2 opacity-50" />
          <p>{emptyMessage}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[600px] text-left text-sm">
        <thead className="bg-background-light dark:bg-background-dark border-y border-border-light dark:border-border-dark">
          <tr>
            {columns.map((column, idx) => (
              <th
                key={idx}
                className={cn(
                  'px-6 py-3 font-medium text-text-sub dark:text-gray-400 text-xs uppercase tracking-wider',
                  column.className
                )}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-border-light dark:divide-border-dark bg-card-light dark:bg-card-dark">
          {data.map(row => (
            <tr
              key={row.id}
              onClick={() => onRowClick?.(row)}
              className={cn(
                'transition-colors',
                hoverable &&
                  'hover:bg-background-light dark:hover:bg-background-dark/50 cursor-pointer',
                onRowClick && 'cursor-pointer'
              )}
            >
              {columns.map((column, idx) => (
                <td key={idx} className={cn('px-6 py-4', column.className)}>
                  {typeof column.accessor === 'function'
                    ? column.accessor(row)
                    : String(row[column.accessor])}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
