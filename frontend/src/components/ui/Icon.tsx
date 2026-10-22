import type { HTMLAttributes } from 'react';
import { cn } from '@/utils';

export interface IconProps extends HTMLAttributes<HTMLSpanElement> {
  name: string;
  size?: number | 'sm' | 'md' | 'lg' | 'xl';
  filled?: boolean;
}

export function Icon({ className, name, size = 'md', filled = false, ...props }: IconProps) {
  const sizeMap = {
    sm: 16,
    md: 20,
    lg: 24,
    xl: 32,
  };

  const iconSize = typeof size === 'number' ? size : sizeMap[size];

  return (
    <span
      className={cn('material-symbols-outlined', filled && 'fill', className)}
      style={{ fontSize: `${iconSize}px` }}
      {...props}
    >
      {name}
    </span>
  );
}
