import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Input, Button, Checkbox, Icon } from '@/components/ui';
import { useAuth } from '@/contexts/AuthContext';

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [remember, setRemember] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await login(email, password);
      navigate('/');
    } catch (err) {
      setError('Email atau password salah. Silakan periksa kembali kredensial Anda.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full max-w-lg">
      <div className="mb-8">
        <h1 className="text-3xl sm:text-4xl font-extrabold text-text-main dark:text-white mb-3">
          Selamat Datang Kembali
        </h1>
        <p className="text-text-sub dark:text-gray-400 text-base font-medium">
          Masukan detail akun anda untuk mengakses dashboard.
        </p>
      </div>

      {/* Tabs */}
      <div className="mb-8">
        <div className="flex p-1 bg-gray-100 dark:bg-gray-800 rounded-xl w-full max-w-sm">
          <button className="flex-1 py-2.5 px-4 rounded-lg bg-white dark:bg-card-dark text-text-main dark:text-white shadow-sm text-sm font-bold">
            Masuk
          </button>
          <Link
            to="/register"
            className="flex-1 py-2.5 px-4 rounded-lg bg-transparent text-text-sub dark:text-gray-400 hover:text-text-main dark:hover:text-white text-sm font-bold text-center"
          >
            Daftar
          </Link>
        </div>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          </div>
        )}

        <Input
          label="Email atau Username"
          type="email"
          placeholder="nama@perusahaan.com"
          value={email}
          onChange={e => setEmail(e.target.value)}
          fullWidth
        />

        <Input
          label="Kata Sandi"
          type="password"
          placeholder="Masukkan kata sandi"
          value={password}
          onChange={e => setPassword(e.target.value)}
          fullWidth
        />

        <div className="flex items-center justify-between">
          <Checkbox
            label="Ingat saya"
            checked={remember}
            onChange={e => setRemember(e.target.checked)}
          />
          <a
            href="#"
            className="text-sm font-bold text-primary hover:text-blue-600 transition-colors"
          >
            Lupa Kata Sandi?
          </a>
        </div>

        <Button
          type="submit"
          variant="primary"
          size="lg"
          className="w-full"
          rightIcon={!isLoading && <Icon name="arrow_forward" size="sm" />}
          disabled={isLoading}
        >
          {isLoading ? 'Memproses...' : 'Masuk ke Dashboard'}
        </Button>

        <div className="relative flex py-2 items-center">
          <div className="flex-grow border-t border-border-light dark:border-border-dark"></div>
          <span className="flex-shrink-0 mx-4 text-gray-400 text-sm font-medium">
            Atau masuk dengan
          </span>
          <div className="flex-grow border-t border-border-light dark:border-border-dark"></div>
        </div>

        <Button
          type="button"
          variant="outline"
          size="lg"
          className="w-full"
          leftIcon={<Icon name="login" size="sm" />}
        >
          Masuk dengan Google
        </Button>

        <div className="text-center pt-2">
          <p className="text-text-sub dark:text-gray-400 text-sm">
            Belum punya akun?{' '}
            <Link to="/register" className="text-primary font-bold hover:underline">
              Daftar Sekarang
            </Link>
          </p>
        </div>
      </form>
    </div>
  );
}
