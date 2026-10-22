import type { HTMLAttributes } from 'react';
import { cn, getInitials } from '@/utils';

export interface AvatarProps extends HTMLAttributes<HTMLDivElement> {
  src?: string;
  alt?: string;
  name?: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  variant?: 'circular' | 'rounded' | 'square';
}

export function Avatar({
  className,
  src,
  alt,
  name,
  size = 'md',
  variant = 'circular',
  ...props
}: AvatarProps) {
  const sizes = {
    xs: 'size-6 text-[10px]',
    sm: 'size-8 text-xs',
    md: 'size-10 text-sm',
    lg: 'size-12 text-base',
    xl: 'size-16 text-lg',
  };

  const variants = {
    circular: 'rounded-full',
    rounded: 'rounded-lg',
    square: 'rounded-none',
  };

  const initials = name ? getInitials(name) : '?';

  if (src) {
    return (
      <div
        className={cn(
          'bg-center bg-no-repeat bg-cover flex-shrink-0 border-2 border-white dark:border-card-dark shadow-sm',
          sizes[size],
          variants[variant],
          className
        )}
        style={{ backgroundImage: `url(${src})` }}
        title={alt || name}
        {...props}
      />
    );
  }

  // Fallback to initials
  return (
    <div
      className={cn(
        'flex items-center justify-center font-bold bg-primary/10 text-primary flex-shrink-0',
        sizes[size],
        variants[variant],
        className
      )}
      title={name}
      {...props}
    >
      {initials}
    </div>
  );
}
