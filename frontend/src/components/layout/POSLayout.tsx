import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { cn } from '@/utils';

/**
 * POSLayout — like MainLayout but the content area is full-height with no
 * padding and no max-width constraint, so POSPage can use the full viewport.
 */
export function POSLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);

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
          <div
            className={cn(
              'fixed inset-y-0 left-0 w-72 bg-card-light dark:bg-card-dark z-50 transform transition-transform lg:hidden',
              sidebarOpen ? 'translate-x-0' : '-translate-x-full'
            )}
          >
            <Sidebar />
          </div>
        </>
      )}

      {/* Main Content — no padding, full height */}
      <main className="flex-1 flex flex-col h-full overflow-hidden relative">
        <Header onMenuClick={() => setSidebarOpen(true)} />

        {/* Page Content — full height, no padding */}
        <div className="flex-1 overflow-hidden bg-background-light dark:bg-background-dark">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
