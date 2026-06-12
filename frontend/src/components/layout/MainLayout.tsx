import { useState } from 'react';
import { Outlet, useNavigation } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Header } from './Header';

interface MainLayoutProps {
  title?: string;
}

export function MainLayout({ title }: MainLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const navigation = useNavigation();
  const isLoading = navigation.state === 'loading';

  return (
    <div className="flex h-screen w-full overflow-hidden">
      {/* Sidebar */}
      <Sidebar />

      {/* Mobile Sidebar Overlay */}
      {sidebarOpen && (
        <>
          <div
            className="fixed inset-0 bg-black/50 z-40 lg:hidden"
            onClick={() => setSidebarOpen(false)}
          />
          <div className="fixed inset-y-0 left-0 w-72 max-w-[85vw] z-50 lg:hidden shadow-xl">
            <Sidebar mobile onNavigate={() => setSidebarOpen(false)} />
          </div>
        </>
      )}

      {/* Main Content */}
      <main className="flex-1 flex flex-col h-full overflow-hidden relative">
        {isLoading && (
          <div className="absolute top-0 left-0 right-0 z-50 h-0.5 bg-blue-100 dark:bg-blue-950 overflow-hidden">
            <div className="h-full bg-blue-500 animate-[progress_1.2s_ease-in-out_infinite]" style={{ width: '40%' }} />
          </div>
        )}
        <Header title={title} onMenuClick={() => setSidebarOpen(true)} />

        {/* Page Content */}
        <div className="flex-1 overflow-y-auto p-4 md:p-6 lg:p-8 scroll-smooth bg-background-light dark:bg-background-dark">
          <div className="mx-auto max-w-7xl">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}
