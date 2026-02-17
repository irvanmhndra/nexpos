import { useState } from 'react';
import { Button, Icon, Input, Badge, Avatar } from '@/components/ui';
import { DeleteConfirmModal } from '@/components/common/DeleteConfirmModal';
import { useUsers, useCreateUser, useUpdateUser, useUpdateUserStatus, useDeleteUser } from '@/hooks/useUsers';
import type { User, UserStatus, CreateUserRequest, UpdateUserRequest } from '@/types/user';

interface UserFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  user?: User | null;
}

function UserFormModal({ isOpen, onClose, user }: UserFormModalProps) {
  const [formData, setFormData] = useState({
    name: user?.name || '',
    email: user?.email || '',
    password: '',
    confirmPassword: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const createUser = useCreateUser();
  const updateUser = useUpdateUser();
  const isLoading = createUser.isPending || updateUser.isPending;

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Nama wajib diisi';
    }
    if (!formData.email.trim()) {
      newErrors.email = 'Email wajib diisi';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Format email tidak valid';
    }
    if (!user) {
      // Creating new user - password required
      if (!formData.password) {
        newErrors.password = 'Password wajib diisi';
      } else if (formData.password.length < 8) {
        newErrors.password = 'Password minimal 8 karakter';
      }
      if (formData.password !== formData.confirmPassword) {
        newErrors.confirmPassword = 'Password tidak sama';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validateForm()) return;

    try {
      if (user) {
        // Update existing user
        const updateData: UpdateUserRequest = {
          name: formData.name,
          email: formData.email,
        };
        await updateUser.mutateAsync({ id: user.id, data: updateData });
      } else {
        // Create new user
        const createData: CreateUserRequest = {
          name: formData.name,
          email: formData.email,
          password: formData.password,
        };
        await createUser.mutateAsync(createData);
      }
      onClose();
    } catch (error) {
      console.error('Error saving user:', error);
    }
  };

  // Reset form when modal opens/closes
  useState(() => {
    if (isOpen) {
      setFormData({
        name: user?.name || '',
        email: user?.email || '',
        password: '',
        confirmPassword: '',
      });
      setErrors({});
    }
  });

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <div className="bg-white dark:bg-card-dark rounded-xl shadow-xl w-full max-w-md">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border-light dark:border-border-dark">
          <h2 className="text-xl font-bold text-text-main dark:text-white">
            {user ? 'Edit Pengguna' : 'Tambah Pengguna'}
          </h2>
          <button
            onClick={onClose}
            className="text-text-sub hover:text-text-main dark:hover:text-white transition-colors"
          >
            <Icon name="close" size={24} />
          </button>
        </div>

        {/* Form */}
        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
              Nama Lengkap <span className="text-red-500">*</span>
            </label>
            <Input
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Masukkan nama lengkap"
              error={errors.name}
            />
          </div>

          <div>
            <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
              Email <span className="text-red-500">*</span>
            </label>
            <Input
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              placeholder="email@contoh.com"
              error={errors.email}
            />
          </div>

          {!user && (
            <>
              <div>
                <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
                  Password <span className="text-red-500">*</span>
                </label>
                <Input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder="Minimal 8 karakter"
                  error={errors.password}
                />
              </div>

              <div>
                <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
                  Konfirmasi Password <span className="text-red-500">*</span>
                </label>
                <Input
                  type="password"
                  value={formData.confirmPassword}
                  onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                  placeholder="Ulangi password"
                  error={errors.confirmPassword}
                />
              </div>
            </>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
          <Button variant="outline" onClick={onClose} disabled={isLoading}>
            Batal
          </Button>
          <Button onClick={handleSubmit} disabled={isLoading}>
            {isLoading ? 'Menyimpan...' : user ? 'Simpan Perubahan' : 'Tambah Pengguna'}
          </Button>
        </div>
      </div>
    </div>
  );
}

export function UserManagement() {
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<UserStatus | ''>('');
  const [isFormModalOpen, setIsFormModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  const { data, isLoading } = useUsers({
    search: search || undefined,
    status: statusFilter || undefined,
  });

  const updateUserStatus = useUpdateUserStatus();
  const deleteUser = useDeleteUser();

  const handleAddUser = () => {
    setSelectedUser(null);
    setIsFormModalOpen(true);
  };

  const handleEditUser = (user: User) => {
    setSelectedUser(user);
    setIsFormModalOpen(true);
  };

  const handleDeleteClick = (user: User) => {
    setSelectedUser(user);
    setIsDeleteModalOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!selectedUser) return;
    try {
      await deleteUser.mutateAsync(selectedUser.id);
      setIsDeleteModalOpen(false);
      setSelectedUser(null);
    } catch (error) {
      console.error('Error deleting user:', error);
    }
  };

  const handleToggleStatus = async (user: User) => {
    const newStatus: UserStatus = user.status === 'active' ? 'inactive' : 'active';
    try {
      await updateUserStatus.mutateAsync({
        id: user.id,
        data: { status: newStatus },
      });
    } catch (error) {
      console.error('Error updating user status:', error);
    }
  };

  const getStatusBadge = (status: UserStatus) => {
    return status === 'active' ? (
      <Badge variant="success" dot>Aktif</Badge>
    ) : (
      <Badge variant="gray" dot>Nonaktif</Badge>
    );
  };

  const getRoleBadge = (user: User) => {
    const roleName = user.role?.name || 'Staff';
    const isOwner = user.role?.code === 'owner';
    return (
      <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
        isOwner
          ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'
          : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300'
      }`}>
        {roleName}
      </span>
    );
  };

  return (
    <div className="flex flex-col gap-6 bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
      {/* Header */}
      <div className="flex flex-col gap-2 p-6 md:p-8 border-b border-border-light dark:border-border-dark">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h2 className="text-2xl md:text-[32px] font-bold text-text-main dark:text-white">
              Pengguna & Tim
            </h2>
            <p className="text-sm md:text-base text-text-sub dark:text-gray-400 mt-1">
              Kelola anggota tim dan akses pengguna
            </p>
          </div>
          <Button leftIcon={<Icon name="person_add" size={20} />} onClick={handleAddUser}>
            Tambah Pengguna
          </Button>
        </div>
      </div>

      {/* Filters */}
      <div className="px-6 md:px-8">
        <div className="flex flex-col sm:flex-row gap-3">
          <div className="relative flex-1">
            <Icon
              name="search"
              size={20}
              className="absolute left-4 top-1/2 -translate-y-1/2 text-text-sub pointer-events-none"
            />
            <input
              type="text"
              placeholder="Cari nama atau email..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full h-12 pl-12 pr-4 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
            />
          </div>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as UserStatus | '')}
            className="h-12 px-4 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
          >
            <option value="">Semua Status</option>
            <option value="active">Aktif</option>
            <option value="inactive">Nonaktif</option>
          </select>
        </div>
      </div>

      {/* User List */}
      <div className="px-6 md:px-8 pb-6 md:pb-8">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Icon name="progress_activity" size={32} className="animate-spin text-primary" />
          </div>
        ) : data?.data.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 gap-3">
            <div className="p-4 bg-background-light dark:bg-background-dark rounded-full">
              <Icon name="group" size={32} className="text-text-sub" />
            </div>
            <p className="text-text-main dark:text-white font-medium">Belum ada pengguna</p>
            <p className="text-sm text-text-sub dark:text-gray-400">
              Tambah pengguna pertama untuk tim Anda
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {data?.data.map((user) => (
              <div
                key={user.id}
                className="flex items-center gap-4 p-4 bg-background-light dark:bg-background-dark rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800/50 transition-colors"
              >
                <Avatar name={user.name} size="md" />

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="font-medium text-text-main dark:text-white truncate">
                      {user.name}
                    </p>
                    {getRoleBadge(user)}
                  </div>
                  <p className="text-sm text-text-sub dark:text-gray-400 truncate">
                    {user.email}
                  </p>
                </div>

                <div className="flex items-center gap-3">
                  {getStatusBadge(user.status)}

                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => handleToggleStatus(user)}
                      className="p-2 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors"
                      title={user.status === 'active' ? 'Nonaktifkan' : 'Aktifkan'}
                    >
                      <Icon
                        name={user.status === 'active' ? 'toggle_on' : 'toggle_off'}
                        size={24}
                        className={user.status === 'active' ? 'text-primary' : 'text-text-sub'}
                      />
                    </button>
                    <button
                      onClick={() => handleEditUser(user)}
                      className="p-2 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors"
                      title="Edit"
                    >
                      <Icon name="edit" size={20} className="text-text-sub" />
                    </button>
                    <button
                      onClick={() => handleDeleteClick(user)}
                      className="p-2 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                      title="Hapus"
                    >
                      <Icon name="delete" size={20} className="text-red-600 dark:text-red-400" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Summary */}
        {data && data.total > 0 && (
          <div className="mt-4 pt-4 border-t border-border-light dark:border-border-dark">
            <p className="text-sm text-text-sub dark:text-gray-400">
              Menampilkan {data.data.length} dari {data.total} pengguna
            </p>
          </div>
        )}
      </div>

      {/* Modals */}
      <UserFormModal
        isOpen={isFormModalOpen}
        onClose={() => {
          setIsFormModalOpen(false);
          setSelectedUser(null);
        }}
        user={selectedUser}
      />

      <DeleteConfirmModal
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setSelectedUser(null);
        }}
        onConfirm={handleDeleteConfirm}
        title="Hapus Pengguna"
        message="Apakah Anda yakin ingin menghapus pengguna"
        itemName={selectedUser?.name}
        isLoading={deleteUser.isPending}
      />
    </div>
  );
}
