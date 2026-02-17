# Aplikabizz Dashboard

Dashboard manajemen bisnis modern dengan React, TypeScript, dan Tailwind CSS.

## Tech Stack

- **React 18** - UI Library
- **TypeScript** - Type Safety
- **Vite** - Build Tool & Dev Server
- **Tailwind CSS** - Styling
- **React Router** - Routing
- **Zustand** - Client State Management
- **React Query** - Server State Management
- **React Hook Form** - Form Management
- **Zod** - Schema Validation
- **MSW** - API Mocking
- **Radix UI** - Accessible Component Primitives
- **Recharts** - Data Visualization

## Getting Started

### Prerequisites

- Node.js >= 18.0.0
- npm atau yarn

### Installation

1. Clone repository atau navigate ke direktori project:
   ```bash
   cd aplikabizz-web
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Copy environment file:
   ```bash
   cp .env.example .env.local
   ```

4. Start development server:
   ```bash
   npm run dev
   ```

   Server akan berjalan di `http://localhost:5173`

### Available Scripts

```bash
# Development
npm run dev          # Start dev server

# Build
npm run build        # Production build
npm run preview      # Preview production build

# Code Quality
npm run lint         # Run ESLint
npm run format       # Format code with Prettier
npm run type-check   # TypeScript type checking
```

## Project Structure

```
src/
├── components/          # Reusable components
│   ├── ui/             # Base UI components (Button, Input, etc)
│   ├── layout/         # Layout components (Sidebar, Header)
│   ├── common/         # Common shared components
│   └── features/       # Feature-specific components
├── pages/              # Page components (routes)
├── hooks/              # Custom React hooks
├── store/              # Zustand stores
├── services/           # API service layer
├── types/              # TypeScript type definitions
├── utils/              # Utility functions
├── config/             # App configuration
└── mocks/              # MSW mock data & handlers
```

## Features

### Implemented
- ✅ Project setup & configuration
- ✅ Tailwind CSS dengan custom theme
- ✅ TypeScript path aliases (@/ imports)
- ✅ MSW untuk API mocking
- ✅ Base utilities (cn, formatters, validators, constants)
- ✅ Folder structure yang scalable

### In Progress
- 🔄 Authentication pages
- 🔄 Dashboard page
- 🔄 Product management
- 🔄 Order management
- 🔄 Customer management
- 🔄 Settings pages

## Design System

### Colors
- Primary: `#137fec` (Blue)
- Background Light: `#f6f7f8`
- Background Dark: `#101922`
- Text Main: `#111418`
- Text Sub: `#617589`

### Typography
- Font Family: Manrope
- Icons: Material Symbols Outlined

### Dark Mode
Dark mode support via class-based theming:
```html
<html class="dark">
```

## API Mocking

Project menggunakan MSW (Mock Service Worker) untuk mock API responses. Mock data otomatis ter-generate menggunakan Faker.js.

Untuk disable MSW saat production:
```env
VITE_ENABLE_MSW=false
```

### Available Mock Endpoints

- `POST /api/auth/login` - Login
- `GET /api/dashboard/stats` - Dashboard statistics
- `GET /api/products` - Get products (with pagination & search)
- `GET /api/orders` - Get orders (with pagination & filters)
- `GET /api/customers` - Get customers (with pagination & search)
- `GET /api/branches` - Get branches

## Development Guidelines

### Code Style
- Use TypeScript strict mode
- Named exports preferred
- PascalCase untuk components
- camelCase untuk functions & variables

### Import Aliases
```typescript
import { Button } from '@/components/ui/Button';
import { useAuth } from '@/hooks/useAuth';
import { formatCurrency } from '@/utils/formatters';
```

### Component Pattern
```typescript
interface ComponentProps {
  // props definition
}

export function Component({ ...props }: ComponentProps) {
  // component logic
  return (
    // JSX
  );
}
```

## Contributing

1. Create feature branch
2. Make changes
3. Run linting & type check
4. Submit PR

## License

Private - All Rights Reserved

## Contact

Tim Aplikabizz - admin@aplikabizz.com
