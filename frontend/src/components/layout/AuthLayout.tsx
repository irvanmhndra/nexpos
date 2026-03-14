import { Outlet, Navigate } from 'react-router-dom';
import { Icon } from '@/components/ui';
import { useAuth } from '@/contexts/AuthContext';

export function AuthLayout() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background-light dark:bg-background-dark">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-text-sub dark:text-gray-400">Loading...</p>
        </div>
      </div>
    );
  }

  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return (
    <div className="h-screen w-full flex overflow-hidden bg-background-light dark:bg-background-dark">
      {/* Left Column: Form Section */}
      <div className="w-full lg:w-1/2 flex flex-col bg-white dark:bg-card-dark relative z-10 overflow-y-auto">
        <div className="flex-1 flex flex-col justify-center px-6 py-12 sm:px-12 lg:px-20 xl:px-32">
          {/* Logo */}
          <div className="mb-10 flex items-center gap-3">
            <div className="flex items-center justify-center size-10 rounded-xl bg-primary/10 text-primary">
              <Icon name="analytics" filled size={28} />
            </div>
            <h2 className="text-text-main dark:text-white text-2xl font-bold tracking-tight">
              Nexpos
            </h2>
          </div>

          {/* Form Content (from Outlet) */}
          <Outlet />
        </div>

        {/* Footer */}
        <div className="px-6 pb-6 text-center lg:text-left lg:px-20 xl:px-32">
          <p className="text-xs text-gray-400">
            © 2024 Nexpos Inc. All rights reserved.
          </p>
        </div>
      </div>

      {/* Right Column: Visual Section */}
      <div className="hidden lg:flex w-1/2 relative bg-primary/5 dark:bg-background-dark flex-col justify-between overflow-hidden">
        {/* Decorative Background Elements */}
        <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-blue-400/20 rounded-full blur-[100px] -translate-y-1/2 translate-x-1/2"></div>
        <div className="absolute bottom-0 left-0 w-[500px] h-[500px] bg-purple-400/20 rounded-full blur-[100px] translate-y-1/2 -translate-x-1/2"></div>

        <div className="relative z-10 flex flex-col h-full items-center justify-center px-12">
          {/* Illustration */}
          <div className="relative w-full max-w-lg aspect-square mb-12">
            <div className="w-full h-full rounded-2xl bg-gradient-to-br from-primary/20 to-purple-500/20 flex items-center justify-center border border-white/10">
              <Icon name="bar_chart" filled size={120} className="text-primary" />
            </div>

            {/* Floating Card 1 */}
            <div className="absolute -bottom-8 -left-8 bg-white dark:bg-card-dark p-4 rounded-xl shadow-xl border border-gray-100 dark:border-border-dark flex items-center gap-4 animate-float">
              <div className="size-10 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center text-green-600 dark:text-green-400">
                <Icon name="trending_up" filled size="md" />
              </div>
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400 font-medium">
                  Total Penjualan
                </p>
                <p className="text-sm font-bold text-text-main dark:text-white">
                  +124.5% Growth
                </p>
              </div>
            </div>

            {/* Floating Card 2 */}
            <div className="absolute -top-6 -right-6 bg-white dark:bg-card-dark p-3 rounded-xl shadow-xl border border-gray-100 dark:border-border-dark">
              <p className="text-[10px] font-semibold text-primary">1,240 Pelanggan</p>
            </div>
          </div>

          {/* Quote */}
          <div className="text-center max-w-md">
            <blockquote className="text-xl font-bold text-text-main dark:text-white mb-4 leading-relaxed">
              "Kelola bisnis Anda lebih efisien bersama kami. Platform all-in-one untuk
              pertumbuhan yang nyata."
            </blockquote>
            <div className="flex items-center justify-center gap-1 text-yellow-500">
              {[...Array(5)].map((_, i) => (
                <Icon key={i} name="star" filled size="sm" />
              ))}
            </div>
            <p className="text-sm text-gray-500 mt-2">Dipercaya oleh 10,000+ Bisnis</p>
          </div>
        </div>
      </div>
    </div>
  );
}
