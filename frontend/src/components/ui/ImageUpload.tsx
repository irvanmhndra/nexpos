import { useState, useRef, useCallback } from 'react';
import { cn } from '@/utils';
import { Icon } from './Icon';

export interface ImageUploadProps {
  value?: string;
  onChange: (imageData: string) => void;
  className?: string;
  error?: string;
}

export function ImageUpload({ value, onChange, className, error }: ImageUploadProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFile = useCallback(
    (file: File) => {
      if (!file.type.startsWith('image/')) {
        return;
      }

      // Max 5MB
      if (file.size > 5 * 1024 * 1024) {
        alert('Ukuran file maksimal 5MB');
        return;
      }

      setIsLoading(true);
      const reader = new FileReader();
      reader.onload = (e) => {
        const result = e.target?.result as string;
        onChange(result);
        setIsLoading(false);
      };
      reader.onerror = () => {
        setIsLoading(false);
      };
      reader.readAsDataURL(file);
    },
    [onChange]
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragging(false);

      const file = e.dataTransfer.files[0];
      if (file) {
        handleFile(file);
      }
    },
    [handleFile]
  );

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const handleClick = () => {
    inputRef.current?.click();
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleFile(file);
    }
  };

  const handleRemove = (e: React.MouseEvent) => {
    e.stopPropagation();
    onChange('');
    if (inputRef.current) {
      inputRef.current.value = '';
    }
  };

  return (
    <div className={className}>
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        onChange={handleInputChange}
        className="hidden"
      />

      {value ? (
        <div className="relative group">
          <img
            src={value}
            alt="Preview"
            className="w-full h-48 object-cover rounded-lg border border-border-light dark:border-border-dark"
          />
          <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity rounded-lg flex items-center justify-center gap-3">
            <button
              type="button"
              onClick={handleClick}
              className="p-2 bg-white rounded-lg hover:bg-gray-100 transition-colors"
              title="Ganti gambar"
            >
              <Icon name="edit" size={20} className="text-gray-700" />
            </button>
            <button
              type="button"
              onClick={handleRemove}
              className="p-2 bg-white rounded-lg hover:bg-gray-100 transition-colors"
              title="Hapus gambar"
            >
              <Icon name="delete" size={20} className="text-red-600" />
            </button>
          </div>
        </div>
      ) : (
        <div
          onClick={handleClick}
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          className={cn(
            'relative flex flex-col items-center justify-center w-full h-48 rounded-lg border-2 border-dashed cursor-pointer transition-all',
            isDragging
              ? 'border-primary bg-primary/5'
              : error
                ? 'border-red-500 bg-red-50 dark:bg-red-900/10'
                : 'border-border-light dark:border-border-dark hover:border-primary hover:bg-primary/5'
          )}
        >
          {isLoading ? (
            <Icon name="progress_activity" size={32} className="animate-spin text-primary" />
          ) : (
            <>
              <div className="p-3 rounded-full bg-primary/10 mb-3">
                <Icon name="cloud_upload" size={28} className="text-primary" />
              </div>
              <p className="text-sm font-medium text-text-main dark:text-white mb-1">
                {isDragging ? 'Lepaskan gambar di sini' : 'Klik atau drag gambar ke sini'}
              </p>
              <p className="text-xs text-text-sub dark:text-gray-400">
                PNG, JPG, WEBP (Maks. 5MB)
              </p>
            </>
          )}
        </div>
      )}

      {error && <p className="text-red-500 text-sm mt-1">{error}</p>}
    </div>
  );
}
