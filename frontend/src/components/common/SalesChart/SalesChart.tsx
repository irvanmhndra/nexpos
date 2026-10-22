import { useMemo } from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { formatCurrency } from '@/utils';

interface SalesTrendData {
  date: string;
  sales: number;
  orders: number;
}

interface SalesChartProps {
  data: SalesTrendData[];
}

export function SalesChart({ data }: SalesChartProps) {
  const chartData = useMemo(() => {
    return data.map(item => ({
      ...item,
      dateFormatted: new Date(item.date).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'short',
      }),
    }));
  }, [data]);

  return (
    <div className="w-full h-[350px]">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart
          data={chartData}
          margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
        >
          <defs>
            <linearGradient id="colorSales" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#137fec" stopOpacity={0.3} />
              <stop offset="95%" stopColor="#137fec" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorOrders" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
              <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" className="stroke-border-light dark:stroke-border-dark" />
          <XAxis
            dataKey="dateFormatted"
            className="text-xs text-text-sub dark:text-gray-400"
            tick={{ fill: 'currentColor' }}
          />
          <YAxis
            yAxisId="left"
            className="text-xs text-text-sub dark:text-gray-400"
            tick={{ fill: 'currentColor' }}
            tickFormatter={value => `${(value / 1000000).toFixed(0)}M`}
          />
          <YAxis
            yAxisId="right"
            orientation="right"
            className="text-xs text-text-sub dark:text-gray-400"
            tick={{ fill: 'currentColor' }}
          />
          <Tooltip
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                return (
                  <div className="bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark rounded-lg shadow-lg p-3">
                    <p className="text-xs text-text-sub dark:text-gray-400 mb-2">
                      {payload[0].payload.date}
                    </p>
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-full bg-primary" />
                        <span className="text-xs text-text-main dark:text-white">
                          Penjualan: {formatCurrency(payload[0].value as number)}
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-full bg-green-500" />
                        <span className="text-xs text-text-main dark:text-white">
                          Pesanan: {payload[1].value}
                        </span>
                      </div>
                    </div>
                  </div>
                );
              }
              return null;
            }}
          />
          <Legend
            wrapperStyle={{ paddingTop: '20px' }}
            iconType="circle"
            formatter={value => (
              <span className="text-sm text-text-main dark:text-white">
                {value === 'sales' ? 'Penjualan' : 'Pesanan'}
              </span>
            )}
          />
          <Area
            yAxisId="left"
            type="monotone"
            dataKey="sales"
            stroke="#137fec"
            strokeWidth={2}
            fillOpacity={1}
            fill="url(#colorSales)"
          />
          <Area
            yAxisId="right"
            type="monotone"
            dataKey="orders"
            stroke="#10b981"
            strokeWidth={2}
            fillOpacity={1}
            fill="url(#colorOrders)"
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
