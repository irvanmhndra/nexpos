import { Button, Icon } from '@/components/ui';

interface DeleteConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title?: string;
  message?: string;
  itemName?: string;
  isLoading?: boolean;
}

export function DeleteConfirmModal({
  isOpen,
  onClose,
  onConfirm,
  title = 'Konfirmasi Hapus',
  message = 'Apakah Anda yakin ingin menghapus',
  itemName,
  isLoading = false,
}: DeleteConfirmModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <div className="bg-white dark:bg-card-dark rounded-xl shadow-xl w-full max-w-md">
        {/* Header */}
        <div className="flex items-center gap-3 p-6 border-b border-border-light dark:border-border-dark">
          <div className="flex items-center justify-center w-12 h-12 rounded-full bg-red-100 dark:bg-red-900/30">
            <Icon name="warning" size={24} className="text-red-600 dark:text-red-400" />
          </div>
          <div className="flex-1">
            <h2 className="text-lg font-bold text-text-main dark:text-white">{title}</h2>
          </div>
          <button
            onClick={onClose}
            disabled={isLoading}
            className="text-text-sub hover:text-text-main dark:hover:text-white transition-colors"
          >
            <Icon name="close" size={24} />
          </button>
        </div>

        {/* Content */}
        <div className="p-6">
          <p className="text-text-main dark:text-gray-200">
            {message} {itemName && <span className="font-bold">"{itemName}"</span>}?
          </p>
          <p className="text-sm text-text-sub dark:text-gray-400 mt-2">
            Tindakan ini tidak dapat dibatalkan.
          </p>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
          <Button variant="outline" onClick={onClose} disabled={isLoading}>
            Batal
          </Button>
          <Button
            onClick={onConfirm}
            disabled={isLoading}
            className="bg-red-600 hover:bg-red-700 text-white"
          >
            {isLoading ? 'Menghapus...' : 'Hapus'}
          </Button>
        </div>
      </div>
    </div>
  );
}
