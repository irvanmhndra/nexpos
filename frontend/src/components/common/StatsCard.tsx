import { Card, Icon } from '@/components/ui';
import { cn } from '@/utils';

interface StatsCardProps {
  title: string;
  value: string | number;
  icon: string;
  trend?: {
    value: string;
    isPositive: boolean;
    label: string;
  };
  className?: string;
}

export function StatsCard({ title, value, icon, trend, className }: StatsCardProps) {
  return (
    <Card variant="bordered" padding="md" className={cn('transition-all hover:shadow-md', className)}>
      <div className="flex items-center justify-between mb-3">
        <p className="text-text-sub dark:text-gray-400 text-sm font-medium">{title}</p>
        <div className="bg-primary/10 p-2 rounded-lg">
          <Icon name={icon} filled className="text-primary" size="md" />
        </div>
      </div>

      <p className="text-2xl font-bold text-text-main dark:text-white mb-1">{value}</p>

      {trend && (
        <div className="flex items-center gap-1 text-sm">
          <Icon
            name={trend.isPositive ? 'trending_up' : 'trending_down'}
            size="sm"
            className={trend.isPositive ? 'text-green-600' : 'text-red-600'}
          />
          <span
            className={cn(
              'font-medium',
              trend.isPositive ? 'text-green-600' : 'text-red-600'
            )}
          >
            {trend.value}
          </span>
          <span className="text-text-sub dark:text-gray-500 ml-1">{trend.label}</span>
        </div>
      )}
    </Card>
  );
}
