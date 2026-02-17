import { useState } from 'react';
import {
  Button,
  Input,
  Select,
  Textarea,
  Checkbox,
  Badge,
  Avatar,
  Icon,
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from './components/ui';

function App() {
  const [checked, setChecked] = useState(false);

  return (
    <div className="min-h-screen bg-background-light dark:bg-background-dark p-8">
      <div className="max-w-6xl mx-auto space-y-8">
        {/* Header */}
        <div className="space-y-2">
          <h1 className="text-4xl font-bold text-text-main dark:text-white">
            Aplikabizz UI Components
          </h1>
          <p className="text-text-sub dark:text-gray-400">
            Testing all base UI components with Tailwind CSS
          </p>
        </div>

        {/* Buttons */}
        <Card variant="bordered">
          <CardHeader>
            <CardTitle>Buttons</CardTitle>
            <CardDescription>Different button variants and sizes</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-3">
              <Button variant="primary">Primary Button</Button>
              <Button variant="secondary">Secondary Button</Button>
              <Button variant="ghost">Ghost Button</Button>
              <Button variant="danger">Danger Button</Button>
              <Button variant="outline">Outline Button</Button>
            </div>

            <div className="flex flex-wrap gap-3 items-center">
              <Button size="sm">Small</Button>
              <Button size="md">Medium</Button>
              <Button size="lg">Large</Button>
            </div>

            <div className="flex flex-wrap gap-3">
              <Button leftIcon={<Icon name="add" size="sm" />}>With Left Icon</Button>
              <Button rightIcon={<Icon name="arrow_forward" size="sm" />}>
                With Right Icon
              </Button>
              <Button isLoading>Loading State</Button>
            </div>
          </CardContent>
        </Card>

        {/* Form Inputs */}
        <Card variant="bordered">
          <CardHeader>
            <CardTitle>Form Inputs</CardTitle>
            <CardDescription>Input fields, select, textarea, and checkbox</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input
                label="Email"
                type="email"
                placeholder="nama@email.com"
                leftIcon={<Icon name="mail" size="sm" />}
              />
              <Input
                label="Password"
                type="password"
                placeholder="Enter password"
                leftIcon={<Icon name="lock" size="sm" />}
              />
            </div>

            <Input
              label="Search"
              type="text"
              placeholder="Search products..."
              leftIcon={<Icon name="search" size="sm" />}
            />

            <Select
              label="Branch"
              leftIcon={<Icon name="store" size="sm" />}
              options={[
                { value: 'all', label: 'Semua Cabang' },
                { value: 'jkt', label: 'Jakarta Pusat' },
                { value: 'bdg', label: 'Bandung' },
              ]}
            />

            <Textarea
              label="Description"
              placeholder="Enter description..."
              helperText="Maximum 500 characters"
            />

            <Checkbox
              label="Remember me"
              checked={checked}
              onChange={e => setChecked(e.target.checked)}
            />
          </CardContent>
        </Card>

        {/* Badges */}
        <Card variant="bordered">
          <CardHeader>
            <CardTitle>Badges</CardTitle>
            <CardDescription>Status indicators and labels</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-2">
              <Badge variant="default">Default</Badge>
              <Badge variant="primary">Primary</Badge>
              <Badge variant="success" dot>
                Success
              </Badge>
              <Badge variant="warning" dot>
                Warning
              </Badge>
              <Badge variant="danger" dot>
                Danger
              </Badge>
              <Badge variant="info">Info</Badge>
            </div>

            <div className="flex flex-wrap gap-2 items-center">
              <Badge size="sm">Small</Badge>
              <Badge size="md">Medium</Badge>
              <Badge size="lg">Large</Badge>
            </div>
          </CardContent>
        </Card>

        {/* Avatars & Icons */}
        <Card variant="bordered">
          <CardHeader>
            <CardTitle>Avatars & Icons</CardTitle>
            <CardDescription>User avatars and material icons</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-4 items-end">
              <Avatar name="Budi Santoso" size="xs" />
              <Avatar name="Sarah Wijaya" size="sm" />
              <Avatar name="John Doe" size="md" />
              <Avatar name="Jane Smith" size="lg" />
              <Avatar name="Admin User" size="xl" />
            </div>

            <div className="flex flex-wrap gap-4 items-center">
              <Icon name="dashboard" size="sm" />
              <Icon name="shopping_cart" size="md" />
              <Icon name="notifications" size="lg" filled />
              <Icon name="settings" size="xl" />
            </div>
          </CardContent>
        </Card>

        {/* Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card variant="default" padding="md">
            <div className="flex items-center justify-between mb-2">
              <p className="text-text-sub text-sm font-medium">Total Sales</p>
              <Icon name="payments" filled className="text-primary" />
            </div>
            <p className="text-2xl font-bold text-text-main dark:text-white">
              Rp 150.000.000
            </p>
            <div className="flex items-center gap-1 text-green-600 text-sm mt-2">
              <Icon name="trending_up" size="sm" />
              <span>+12% dari bulan lalu</span>
            </div>
          </Card>

          <Card variant="bordered" padding="md">
            <div className="flex items-center justify-between mb-2">
              <p className="text-text-sub text-sm font-medium">New Orders</p>
              <Icon name="shopping_bag" filled className="text-primary" />
            </div>
            <p className="text-2xl font-bold text-text-main dark:text-white">124</p>
            <div className="flex items-center gap-1 text-green-600 text-sm mt-2">
              <Icon name="trending_up" size="sm" />
              <span>+5% dari bulan lalu</span>
            </div>
          </Card>

          <Card variant="elevated" padding="md">
            <div className="flex items-center justify-between mb-2">
              <p className="text-text-sub text-sm font-medium">Customers</p>
              <Icon name="group" filled className="text-primary" />
            </div>
            <p className="text-2xl font-bold text-text-main dark:text-white">1,240</p>
            <div className="flex items-center gap-1 text-green-600 text-sm mt-2">
              <Icon name="trending_up" size="sm" />
              <span>+8% dari bulan lalu</span>
            </div>
          </Card>
        </div>

        {/* Dark Mode Toggle */}
        <Card variant="bordered">
          <CardContent className="flex items-center justify-between">
            <div>
              <p className="font-semibold text-text-main dark:text-white">Dark Mode</p>
              <p className="text-sm text-text-sub dark:text-gray-400">
                Toggle dark mode (add to html class="dark")
              </p>
            </div>
            <Button
              variant="outline"
              onClick={() => {
                document.documentElement.classList.toggle('dark');
              }}
            >
              <Icon name="dark_mode" size="sm" />
              Toggle
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default App;
