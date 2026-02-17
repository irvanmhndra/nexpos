import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Button, Input, Icon, ImageUpload, Card } from '@/components/ui';
import { useProduct, useCreateProduct, useUpdateProduct } from '@/hooks/useProducts';
import { useAllProductCategories } from '@/hooks/useProductCategories';
import type { CreateProductRequest, UpdateProductRequest, ProductVariantInput } from '@/types/product';

interface VariantFormData {
  id?: number;
  name: string;
  sku: string;
  price: number;
  standard_cost: number;
  is_default: boolean;
}

interface ProductFormData {
  name: string;
  category_id: string;
  description: string;
  image_data: string;
  hasMultipleVariants: boolean;
  // Single variant fields (when toggle OFF)
  sku: string;
  price: number;
  standard_cost: number;
  // Multiple variants (when toggle ON)
  variants: VariantFormData[];
}

export function ProductFormPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEditing = !!id;
  const productId = id ? parseInt(id) : 0;

  const { data: product, isLoading: isLoadingProduct } = useProduct(productId);
  const { data: categories } = useAllProductCategories();
  const createProduct = useCreateProduct();
  const updateProduct = useUpdateProduct();

  const [formData, setFormData] = useState<ProductFormData>({
    name: '',
    category_id: '',
    description: '',
    image_data: '',
    hasMultipleVariants: false,
    // Single variant defaults
    sku: '',
    price: 0,
    standard_cost: 0,
    // Multiple variants
    variants: [],
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (product) {
      const hasMultipleVariants = product.variants && product.variants.length > 1;
      const firstVariant = product.variants?.[0];

      setFormData({
        name: product.name,
        category_id: product.category_id?.toString() || '',
        description: product.description || '',
        image_data: product.image_data || '',
        hasMultipleVariants: !!hasMultipleVariants,
        // Single variant from first variant
        sku: firstVariant?.sku || '',
        price: firstVariant?.price || 0,
        standard_cost: firstVariant?.standard_cost || 0,
        // All variants for multiple
        variants: hasMultipleVariants
          ? product.variants.map(v => ({
              id: v.id,
              name: v.name,
              sku: v.sku,
              price: v.price,
              standard_cost: v.standard_cost,
              is_default: v.is_default,
            }))
          : [],
      });
    }
  }, [product]);

  const generateVariantSku = (index: number) => {
    const baseSku = formData.sku || 'SKU';
    return `${baseSku}-${index + 1}`;
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Nama produk wajib diisi';
    }

    if (formData.hasMultipleVariants) {
      // Validate multiple variants
      if (formData.variants.length < 2) {
        newErrors.variants = 'Minimal 2 varian harus ditambahkan';
      }
      formData.variants.forEach((variant, index) => {
        if (!variant.name.trim()) {
          newErrors[`variant_${index}_name`] = 'Nama varian wajib diisi';
        }
        if (!variant.sku.trim()) {
          newErrors[`variant_${index}_sku`] = 'SKU wajib diisi';
        }
        if (variant.price <= 0) {
          newErrors[`variant_${index}_price`] = 'Harga harus lebih dari 0';
        }
      });
    } else {
      // Validate single variant
      if (!formData.sku.trim()) {
        newErrors.sku = 'SKU wajib diisi';
      }
      if (formData.price <= 0) {
        newErrors.price = 'Harga harus lebih dari 0';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    // Build variants array - always at least 1 variant
    const variants: ProductVariantInput[] = formData.hasMultipleVariants
      ? formData.variants.map((v, idx) => ({
          id: v.id,
          name: v.name,
          sku: v.sku,
          price: v.price,
          standard_cost: v.standard_cost,
          is_default: v.is_default || idx === 0,
        }))
      : [
          {
            name: 'Default',
            sku: formData.sku,
            price: formData.price,
            standard_cost: formData.standard_cost,
            is_default: true,
          },
        ];

    const submitData: CreateProductRequest | UpdateProductRequest = {
      name: formData.name,
      category_id: formData.category_id ? parseInt(formData.category_id) : null,
      description: formData.description || undefined,
      image_data: formData.image_data || undefined,
      is_active: true,
      variants,
    };

    try {
      if (isEditing && productId) {
        await updateProduct.mutateAsync({ id: productId, product: submitData as UpdateProductRequest });
      } else {
        await createProduct.mutateAsync(submitData as CreateProductRequest);
      }
      navigate('/products');
    } catch (error) {
      console.error('Error saving product:', error);
    }
  };

  const handleToggleVariants = () => {
    if (!formData.hasMultipleVariants) {
      // Switching to multiple variants - initialize with 2 variants
      setFormData({
        ...formData,
        hasMultipleVariants: true,
        variants: [
          {
            name: '',
            sku: generateVariantSku(0),
            price: formData.price || 0,
            standard_cost: formData.standard_cost || 0,
            is_default: true,
          },
          {
            name: '',
            sku: generateVariantSku(1),
            price: formData.price || 0,
            standard_cost: formData.standard_cost || 0,
            is_default: false,
          },
        ],
      });
    } else {
      // Switching to single variant - take values from first variant
      const firstVariant = formData.variants[0];
      setFormData({
        ...formData,
        hasMultipleVariants: false,
        sku: firstVariant?.sku || formData.sku,
        price: firstVariant?.price || formData.price,
        standard_cost: firstVariant?.standard_cost || formData.standard_cost,
        variants: [],
      });
    }
  };

  const addVariant = () => {
    const newIndex = formData.variants.length;
    const lastVariant = formData.variants[formData.variants.length - 1];
    setFormData({
      ...formData,
      variants: [
        ...formData.variants,
        {
          name: '',
          sku: generateVariantSku(newIndex),
          price: lastVariant?.price || formData.price || 0,
          standard_cost: lastVariant?.standard_cost || formData.standard_cost || 0,
          is_default: false,
        },
      ],
    });
  };

  const removeVariant = (index: number) => {
    const newVariants = formData.variants.filter((_, i) => i !== index);
    // Ensure at least one variant is default
    if (newVariants.length > 0 && !newVariants.some(v => v.is_default)) {
      newVariants[0].is_default = true;
    }
    setFormData({ ...formData, variants: newVariants });
  };

  const updateVariant = (index: number, field: keyof VariantFormData, value: string | number | boolean) => {
    const newVariants = [...formData.variants];
    newVariants[index] = { ...newVariants[index], [field]: value };

    // If setting this variant as default, unset others
    if (field === 'is_default' && value === true) {
      newVariants.forEach((v, i) => {
        if (i !== index) v.is_default = false;
      });
    }

    setFormData({ ...formData, variants: newVariants });
  };

  const isLoading = createProduct.isPending || updateProduct.isPending;

  if (isEditing && isLoadingProduct) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Icon name="progress_activity" size={32} className="animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/products')}
          className="p-2 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
        >
          <Icon name="arrow_back" size={24} className="text-text-main dark:text-white" />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-text-main dark:text-white">
            {isEditing ? 'Edit Produk' : 'Tambah Produk Baru'}
          </h1>
          <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
            {isEditing ? 'Perbarui informasi produk' : 'Isi informasi produk yang akan ditambahkan'}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - Main Form */}
          <div className="lg:col-span-2 space-y-6">
            {/* Basic Info Card */}
            <Card variant="bordered" padding="lg">
              <h2 className="text-lg font-bold text-text-main dark:text-white mb-5">
                Informasi Produk
              </h2>

              <div className="space-y-5">
                {/* Name */}
                <div>
                  <label className="block text-sm font-medium text-text-main dark:text-gray-200 mb-2">
                    Nama Produk <span className="text-red-500">*</span>
                  </label>
                  <Input
                    value={formData.name}
                    onChange={e => setFormData({ ...formData, name: e.target.value })}
                    placeholder="Contoh: Kaos Polos Premium"
                    error={errors.name}
                  />
                </div>

                {/* Category */}
                <div>
                  <label className="block text-sm font-medium text-text-main dark:text-gray-200 mb-2">
                    Kategori
                  </label>
                  <select
                    value={formData.category_id}
                    onChange={e => setFormData({ ...formData, category_id: e.target.value })}
                    className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-12 px-4 text-base"
                  >
                    <option value="">Pilih Kategori (Opsional)</option>
                    {(categories || []).map(cat => (
                      <option key={cat.id} value={cat.id}>
                        {cat.name}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Description */}
                <div>
                  <label className="block text-sm font-medium text-text-main dark:text-gray-200 mb-2">
                    Deskripsi
                  </label>
                  <textarea
                    value={formData.description}
                    onChange={e => setFormData({ ...formData, description: e.target.value })}
                    placeholder="Deskripsi produk..."
                    rows={3}
                    className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary p-4 resize-none"
                  />
                </div>
              </div>
            </Card>

            {/* Variant Type Selection */}
            <Card variant="bordered" padding="lg">
              <h2 className="text-lg font-bold text-text-main dark:text-white mb-4">
                Tipe Produk
              </h2>

              {/* Selectable Cards */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
                {/* Option: Tanpa Varian */}
                <button
                  type="button"
                  onClick={() => {
                    if (formData.hasMultipleVariants) {
                      handleToggleVariants();
                    }
                  }}
                  className={`p-4 rounded-xl border-2 text-left transition-all ${
                    !formData.hasMultipleVariants
                      ? 'border-primary bg-primary/5 dark:bg-primary/10'
                      : 'border-border-light dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 mt-0.5 ${
                      !formData.hasMultipleVariants
                        ? 'border-primary'
                        : 'border-gray-300 dark:border-gray-500'
                    }`}>
                      {!formData.hasMultipleVariants && (
                        <div className="w-2.5 h-2.5 rounded-full bg-primary" />
                      )}
                    </div>
                    <div>
                      <p className={`font-semibold ${
                        !formData.hasMultipleVariants
                          ? 'text-primary'
                          : 'text-text-main dark:text-white'
                      }`}>
                        Produk tanpa varian
                      </p>
                      <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                        Satu SKU dan satu harga
                      </p>
                    </div>
                  </div>
                </button>

                {/* Option: Dengan Varian */}
                <button
                  type="button"
                  onClick={() => {
                    if (!formData.hasMultipleVariants) {
                      handleToggleVariants();
                    }
                  }}
                  className={`p-4 rounded-xl border-2 text-left transition-all ${
                    formData.hasMultipleVariants
                      ? 'border-primary bg-primary/5 dark:bg-primary/10'
                      : 'border-border-light dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 mt-0.5 ${
                      formData.hasMultipleVariants
                        ? 'border-primary'
                        : 'border-gray-300 dark:border-gray-500'
                    }`}>
                      {formData.hasMultipleVariants && (
                        <div className="w-2.5 h-2.5 rounded-full bg-primary" />
                      )}
                    </div>
                    <div>
                      <p className={`font-semibold ${
                        formData.hasMultipleVariants
                          ? 'text-primary'
                          : 'text-text-main dark:text-white'
                      }`}>
                        Produk dengan varian
                      </p>
                      <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                        Variasi ukuran, warna, atau pilihan lainnya
                      </p>
                    </div>
                  </div>
                </button>
              </div>

              {!formData.hasMultipleVariants ? (
                /* Single Variant Form */
                <div className="space-y-4">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-text-main dark:text-gray-200 mb-2">
                        SKU <span className="text-red-500">*</span>
                      </label>
                      <Input
                        value={formData.sku}
                        onChange={e => setFormData({ ...formData, sku: e.target.value })}
                        placeholder="Contoh: KP-001"
                        error={errors.sku}
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-text-main dark:text-gray-200 mb-2">
                        Harga Jual (Rp) <span className="text-red-500">*</span>
                      </label>
                      <Input
                        type="number"
                        value={formData.price}
                        onChange={e => setFormData({ ...formData, price: Number(e.target.value) })}
                        placeholder="50000"
                        min="0"
                        error={errors.price}
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-text-main dark:text-gray-200 mb-2">
                      Harga Modal (Rp)
                    </label>
                    <Input
                      type="number"
                      value={formData.standard_cost}
                      onChange={e => setFormData({ ...formData, standard_cost: Number(e.target.value) })}
                      placeholder="30000"
                      min="0"
                    />
                    <p className="text-xs text-text-sub dark:text-gray-400 mt-1">
                      Opsional, untuk perhitungan profit
                    </p>
                  </div>
                </div>
              ) : (
                /* Multiple Variants Form */
                <div className="space-y-4">
                  {errors.variants && (
                    <p className="text-red-500 text-sm">{errors.variants}</p>
                  )}

                  {/* Variant List */}
                  <div className="space-y-3">
                    {formData.variants.map((variant, index) => (
                      <div
                        key={index}
                        className="p-4 bg-background-light dark:bg-background-dark rounded-lg"
                      >
                        <div className="flex items-center justify-between mb-3">
                          <div className="flex items-center gap-2">
                            <span className="text-sm font-medium text-text-main dark:text-white">
                              Varian {index + 1}
                            </span>
                            {variant.is_default && (
                              <span className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded">
                                Default
                              </span>
                            )}
                          </div>
                          <div className="flex items-center gap-2">
                            {!variant.is_default && (
                              <button
                                type="button"
                                onClick={() => updateVariant(index, 'is_default', true)}
                                className="text-xs text-primary hover:underline"
                              >
                                Set Default
                              </button>
                            )}
                            {formData.variants.length > 2 && (
                              <button
                                type="button"
                                onClick={() => removeVariant(index)}
                                className="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                              >
                                <Icon name="delete" size={18} className="text-red-600 dark:text-red-400" />
                              </button>
                            )}
                          </div>
                        </div>

                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
                          <div>
                            <label className="block text-xs text-text-sub dark:text-gray-400 mb-1">
                              Nama <span className="text-red-500">*</span>
                            </label>
                            <Input
                              value={variant.name}
                              onChange={e => updateVariant(index, 'name', e.target.value)}
                              placeholder="cth: Merah M"
                              error={errors[`variant_${index}_name`]}
                            />
                          </div>
                          <div>
                            <label className="block text-xs text-text-sub dark:text-gray-400 mb-1">
                              SKU <span className="text-red-500">*</span>
                            </label>
                            <Input
                              value={variant.sku}
                              onChange={e => updateVariant(index, 'sku', e.target.value)}
                              placeholder={generateVariantSku(index)}
                              error={errors[`variant_${index}_sku`]}
                            />
                          </div>
                          <div>
                            <label className="block text-xs text-text-sub dark:text-gray-400 mb-1">
                              Harga <span className="text-red-500">*</span>
                            </label>
                            <Input
                              type="number"
                              value={variant.price}
                              onChange={e => updateVariant(index, 'price', Number(e.target.value))}
                              placeholder="50000"
                              min="0"
                              error={errors[`variant_${index}_price`]}
                            />
                          </div>
                          <div>
                            <label className="block text-xs text-text-sub dark:text-gray-400 mb-1">
                              Modal
                            </label>
                            <Input
                              type="number"
                              value={variant.standard_cost}
                              onChange={e => updateVariant(index, 'standard_cost', Number(e.target.value))}
                              placeholder="30000"
                              min="0"
                            />
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>

                  <Button
                    type="button"
                    variant="outline"
                    onClick={addVariant}
                    className="w-full"
                  >
                    <Icon name="add" size={20} />
                    Tambah Varian
                  </Button>
                </div>
              )}
            </Card>
          </div>

          {/* Right Column - Image & Actions */}
          <div className="space-y-6 lg:sticky lg:top-6 lg:self-start">
            {/* Image Card */}
            <Card variant="bordered" padding="lg">
              <h2 className="text-lg font-bold text-text-main dark:text-white mb-4">
                Gambar Produk
              </h2>
              <ImageUpload
                value={formData.image_data}
                onChange={(imageData) => setFormData({ ...formData, image_data: imageData })}
              />
            </Card>

            {/* Actions Card */}
            <Card variant="bordered" padding="lg">
              <div className="space-y-3">
                <Button
                  type="submit"
                  className="w-full"
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <>
                      <Icon name="progress_activity" size={20} className="animate-spin" />
                      Menyimpan...
                    </>
                  ) : (
                    <>
                      <Icon name="save" size={20} />
                      {isEditing ? 'Simpan Perubahan' : 'Simpan Produk'}
                    </>
                  )}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  className="w-full"
                  onClick={() => navigate('/products')}
                  disabled={isLoading}
                >
                  Batal
                </Button>
              </div>
            </Card>

            {/* Tips Card */}
            <Card variant="bordered" padding="md" className="bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800">
              <div className="flex gap-3">
                <Icon name="lightbulb" size={20} className="text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-blue-900 dark:text-blue-100">Tips</p>
                  <ul className="text-xs text-blue-700 dark:text-blue-300 mt-1 space-y-1">
                    <li>- Gunakan nama produk yang jelas dan deskriptif</li>
                    <li>- SKU harus unik untuk setiap varian</li>
                    <li>- Aktifkan toggle jika produk punya variasi (ukuran/warna)</li>
                  </ul>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </form>
    </div>
  );
}
