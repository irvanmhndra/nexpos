import { format, formatDistanceToNow } from 'date-fns';
import { id as idLocale } from 'date-fns/locale';

/**
 * Format number to Indonesian Rupiah currency
 * @param amount - Number to format
 * @returns Formatted currency string (e.g., "Rp 1.250.000")
 */
export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

/**
 * Format number with thousand separators
 * @param num - Number to format
 * @returns Formatted number string (e.g., "1.250.000")
 */
export function formatNumber(num: number): string {
  return new Intl.NumberFormat('id-ID').format(num);
}

/**
 * Format date to Indonesian format
 * @param date - Date to format
 * @param formatStr - Format string (default: "dd MMM yyyy")
 * @returns Formatted date string (e.g., "12 Okt 2023")
 */
export function formatDate(date: Date | string, formatStr: string = 'dd MMM yyyy'): string {
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  return format(dateObj, formatStr, { locale: idLocale });
}

/**
 * Format date to relative time
 * @param date - Date to format
 * @returns Relative time string (e.g., "2 hari yang lalu")
 */
export function formatRelativeTime(date: Date | string): string {
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  return formatDistanceToNow(dateObj, { locale: idLocale, addSuffix: true });
}

/**
 * Format phone number to Indonesian format
 * @param phone - Phone number string
 * @returns Formatted phone string (e.g., "+62 812 3456 7890")
 */
export function formatPhoneNumber(phone: string): string {
  // Remove all non-numeric characters
  const cleaned = phone.replace(/\D/g, '');

  // Format based on Indonesian phone number pattern
  if (cleaned.startsWith('62')) {
    const number = cleaned.substring(2);
    return `+62 ${number.substring(0, 3)} ${number.substring(3, 7)} ${number.substring(7)}`;
  } else if (cleaned.startsWith('0')) {
    const number = cleaned.substring(1);
    return `+62 ${number.substring(0, 3)} ${number.substring(3, 7)} ${number.substring(7)}`;
  }

  return phone;
}

/**
 * Format percentage
 * @param value - Number to format as percentage
 * @param decimals - Number of decimal places (default: 0)
 * @returns Formatted percentage string (e.g., "12%")
 */
export function formatPercentage(value: number, decimals: number = 0): string {
  return `${value.toFixed(decimals)}%`;
}

/**
 * Truncate text with ellipsis
 * @param text - Text to truncate
 * @param maxLength - Maximum length before truncation
 * @returns Truncated text with ellipsis if needed
 */
export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

/**
 * Get initials from name
 * @param name - Full name
 * @returns Initials (e.g., "John Doe" -> "JD")
 */
export function getInitials(name: string): string {
  return name
    .split(' ')
    .map(word => word[0])
    .join('')
    .toUpperCase()
    .substring(0, 2);
}

/**
 * Format file size
 * @param bytes - File size in bytes
 * @returns Formatted file size (e.g., "2.5 MB")
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}
