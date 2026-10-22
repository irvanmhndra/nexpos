import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Input, Button, Icon } from '@/components/ui';
import { useAuth } from '@/contexts/AuthContext';
import { isValidEmail, isValidPassword } from '@/utils/validators';

interface FormData {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
}

interface FormErrors {
  name?: string;
  email?: string;
  password?: string;
  confirmPassword?: string;
}

export function RegisterPage() {
  const navigate = useNavigate();
  const { register } = useAuth();
  const [formData, setFormData] = useState<FormData>({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  });
  const [acceptTerms, setAcceptTerms] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [errors, setErrors] = useState<FormErrors>({});

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Nama lengkap wajib diisi';
    } else if (formData.name.trim().length < 3) {
      newErrors.name = 'Nama minimal 3 karakter';
    }

    if (!formData.email.trim()) {
      newErrors.email = 'Email wajib diisi';
    } else if (!isValidEmail(formData.email)) {
      newErrors.email = 'Format email tidak valid';
    }

    if (!formData.password) {
      newErrors.password = 'Kata sandi wajib diisi';
    } else if (!isValidPassword(formData.password)) {
      newErrors.password = 'Kata sandi minimal 8 karakter, mengandung huruf dan angka';
    }

    if (!formData.confirmPassword) {
      newErrors.confirmPassword = 'Konfirmasi kata sandi wajib diisi';
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Kata sandi tidak cocok';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange = (field: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, [field]: e.target.value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!validateForm()) {
      return;
    }

    if (!acceptTerms) {
      setError('Anda harus menyetujui syarat dan ketentuan');
      return;
    }

    setIsLoading(true);

    try {
      await register(formData.name, formData.email, formData.password);
      navigate('/');
    } catch {
      setError('Pendaftaran gagal. Email mungkin sudah terdaftar.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full max-w-lg">
      <div className="mb-8">
        <h1 className="text-3xl sm:text-4xl font-extrabold text-text-main dark:text-white mb-3">
          Buat Akun Baru
        </h1>
        <p className="text-text-sub dark:text-gray-400 text-base font-medium">
          Daftar untuk mulai mengelola bisnis Anda dengan Nexpos.
        </p>
      </div>

      {/* Tabs */}
      <div className="mb-8">
        <div className="flex p-1 bg-gray-100 dark:bg-gray-800 rounded-xl w-full max-w-sm">
          <Link
            to="/login"
            className="flex-1 py-2.5 px-4 rounded-lg bg-transparent text-text-sub dark:text-gray-400 hover:text-text-main dark:hover:text-white text-sm font-bold text-center"
          >
            Masuk
          </Link>
          <button className="flex-1 py-2.5 px-4 rounded-lg bg-white dark:bg-card-dark text-text-main dark:text-white shadow-sm text-sm font-bold">
            Daftar
          </button>
        </div>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-5">
        {error && (
          <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          </div>
        )}

        <Input
          label="Nama Lengkap"
          type="text"
          placeholder="Masukkan nama lengkap"
          value={formData.name}
          onChange={handleChange('name')}
          error={errors.name}
          fullWidth
        />

        <Input
          label="Email"
          type="email"
          placeholder="nama@perusahaan.com"
          value={formData.email}
          onChange={handleChange('email')}
          error={errors.email}
          fullWidth
        />

        <Input
          label="Kata Sandi"
          type="password"
          placeholder="Minimal 8 karakter"
          value={formData.password}
          onChange={handleChange('password')}
          error={errors.password}
          helperText={!errors.password ? 'Gunakan kombinasi huruf dan angka' : undefined}
          fullWidth
        />

        <Input
          label="Konfirmasi Kata Sandi"
          type="password"
          placeholder="Masukkan ulang kata sandi"
          value={formData.confirmPassword}
          onChange={handleChange('confirmPassword')}
          error={errors.confirmPassword}
          fullWidth
        />

        <div className="pt-1">
          <label className="flex items-center gap-3 cursor-pointer group">
            <div className="relative flex items-center">
              <input
                type="checkbox"
                checked={acceptTerms}
                onChange={e => setAcceptTerms(e.target.checked)}
                className="peer size-5 cursor-pointer appearance-none rounded border bg-white dark:bg-card-dark checked:bg-primary checked:border-primary focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 transition-all border-border-light dark:border-border-dark"
              />
              <span className="material-symbols-outlined absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-white opacity-0 peer-checked:opacity-100 text-[16px] pointer-events-none">
                check
              </span>
            </div>
            <span className="text-sm text-text-sub dark:text-gray-400">
              Saya setuju dengan{' '}
              <a href="#" className="text-primary hover:underline">
                Syarat & Ketentuan
              </a>{' '}
              dan{' '}
              <a href="#" className="text-primary hover:underline">
                Kebijakan Privasi
              </a>
            </span>
          </label>
        </div>

        <Button
          type="submit"
          variant="primary"
          size="lg"
          className="w-full"
          rightIcon={!isLoading && <Icon name="arrow_forward" size="sm" />}
          disabled={isLoading}
        >
          {isLoading ? 'Memproses...' : 'Daftar Sekarang'}
        </Button>

        <div className="text-center pt-2">
          <p className="text-text-sub dark:text-gray-400 text-sm">
            Sudah punya akun?{' '}
            <Link to="/login" className="text-primary font-bold hover:underline">
              Masuk Sekarang
            </Link>
          </p>
        </div>
      </form>
    </div>
  );
}
