import {
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { formatCurrency } from '@/utils';

interface BarChartData {
  name: string;
  value: number;
}

interface BarChartProps {
  data: BarChartData[];
  color?: string;
  valuePrefix?: string;
  height?: number;
}

export function BarChart({ data, color = '#137fec', valuePrefix = '', height = 350 }: BarChartProps) {
  return (
    <div className="w-full" style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <RechartsBarChart
          data={data}
          margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
        >
          <CartesianGrid strokeDasharray="3 3" className="stroke-border-light dark:stroke-border-dark" />
          <XAxis
            dataKey="name"
            className="text-xs text-text-sub dark:text-gray-400"
            tick={{ fill: 'currentColor' }}
          />
          <YAxis
            className="text-xs text-text-sub dark:text-gray-400"
            tick={{ fill: 'currentColor' }}
            tickFormatter={value => {
              if (valuePrefix === 'Rp') {
                return `${(value / 1000000).toFixed(0)}M`;
              }
              return value;
            }}
          />
          <Tooltip
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                return (
                  <div className="bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark rounded-lg shadow-lg p-3">
                    <p className="text-xs text-text-sub dark:text-gray-400 mb-1">
                      {payload[0].payload.name}
                    </p>
                    <p className="text-sm font-medium text-text-main dark:text-white">
                      {valuePrefix === 'Rp'
                        ? formatCurrency(payload[0].value as number)
                        : `${payload[0].value} unit`}
                    </p>
                  </div>
                );
              }
              return null;
            }}
          />
          <Bar dataKey="value" fill={color} radius={[8, 8, 0, 0]} />
        </RechartsBarChart>
      </ResponsiveContainer>
    </div>
  );
}
