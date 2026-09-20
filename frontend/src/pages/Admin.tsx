import { Component, createSignal, createResource, For, Show } from 'solid-js';
import { Plus, Trash2, Edit3, X, RefreshCw, Palette, CreditCard, Printer, Percent, Building2, Truck, Mail, Users, Image as ImageIcon, Copy, Check, UploadCloud, BarChart3, Download, Key, Send, ShieldCheck } from 'lucide-solid';
import { api } from '../lib/api';
import { useI18n } from '../store/i18n';
import { useCurrency } from '../store/currency';
import './Admin.scss';

export const Admin: Component = () => {
  const { t } = useI18n();
  const { formatPrice } = useCurrency();

  const [activeTab, setActiveTab] = createSignal<'products' | 'orders' | 'analytics' | 'customers' | 'media' | 'widgets' | 'coupons' | 'themes' | 'settings'>('products');
  const [settingsSubTab, setSettingsSubTab] = createSignal<'general' | 'shipping' | 'taxes' | 'payments' | 'pod' | 'email' | 'apikeys'>('general');
  
  // Analytics
  const [analyticsDays, setAnalyticsDays] = createSignal(30);

  // Product Modal states
  const [showProductModal, setShowProductModal] = createSignal(false);
  const [modalTab, setModalTab] = createSignal<'general' | 'variants' | 'images'>('general');
  const [editingProductId, setEditingProductId] = createSignal<string | null>(null);

  const [prodName, setProdName] = createSignal('');
  const [prodSlug, setProdSlug] = createSignal('');
  const [prodDesc, setProdDesc] = createSignal('');
  const [prodPrice, setProdPrice] = createSignal(39.95);
  const [prodFeatured, setProdFeatured] = createSignal(false);
  const [prodType, setProdType] = createSignal<'variable' | 'simple'>('variable');
  
  const [currentVariants, setCurrentVariants] = createSignal<any[]>([]);
  const [currentImages, setCurrentImages] = createSignal<any[]>([]);

  const [newVarSize, setNewVarSize] = createSignal('M');
  const [newVarColor, setNewVarColor] = createSignal('Black');
  const [newVarStock, setNewVarStock] = createSignal(25);
  const [newVarPrice, setNewVarPrice] = createSignal<number | null>(null);
  const [newVarImage, setNewVarImage] = createSignal('');
  const [newVarPODProvider, setNewVarPODProvider] = createSignal('none');
  const [newVarPODID, setNewVarPODID] = createSignal('');
  const [newVarPrintFile, setNewVarPrintFile] = createSignal('');

  const [newImageUrl, setNewImageUrl] = createSignal('');
  const [newImageIsPrimary, setNewImageIsPrimary] = createSignal(false);
  const [newImageVariantId, setNewImageVariantId] = createSignal<string>('');

  // Coupon Modal
  const [showCouponModal, setShowCouponModal] = createSignal(false);
  const [couponCode, setCouponCode] = createSignal('');
  const [couponType, setCouponType] = createSignal<'percent' | 'fixed'>('percent');
  const [couponValue, setCouponValue] = createSignal(10);
  const [couponMinSpend, setCouponMinSpend] = createSignal(25);

  // Widget Modal
  const [showWidgetModal, setShowWidgetModal] = createSignal(false);
  const [editingWidgetId, setEditingWidgetId] = createSignal<string | null>(null);
  const [widgetType, setWidgetType] = createSignal('hero');
  const [widgetTitle, setWidgetTitle] = createSignal('');
  const [widgetSubtitle, setWidgetSubtitle] = createSignal('');
  const [widgetOrder, setWidgetOrder] = createSignal(1);
  const [widgetActive, setWidgetActive] = createSignal(true);
  const [widgetBgImage, setWidgetBgImage] = createSignal('');
  const [widgetBtnText, setWidgetBtnText] = createSignal('');
  const [widgetBtnLink, setWidgetBtnLink] = createSignal('');

  // API Key Generatie Modal
  const [showApiKeyModal, setShowApiKeyModal] = createSignal(false);
  const [newKeyDesc, setNewKeyDesc] = createSignal('Gelato POD Integratie');
  const [newKeyPerms, setNewKeyPerms] = createSignal('read_write');
  const [generatedKeyResult, setGeneratedKeyResult] = createSignal<{ ck: string; cs: string } | null>(null);

  // Nieuwe Btw & Verzendmethoden invoer
  const [newTaxCountry, setNewTaxCountry] = createSignal('FR');
  const [newTaxName, setNewTaxName] = createSignal('Franse TVA (20%)');
  const [newTaxRate, setNewTaxRate] = createSignal(20.0);

  const [newShipTitle, setNewShipTitle] = createSignal('');
  const [newShipCost, setNewShipCost] = createSignal(4.95);
  const [newShipThreshold, setNewShipThreshold] = createSignal<number | null>(50);

  // Test E-mail state
  const [testEmailTo, setTestEmailTo] = createSignal('');
  const [syncingPOD, setSyncingPOD] = createSignal(false);
  const [copiedUrl, setCopiedUrl] = createSignal<string | null>(null);

  // Data resources
  const [stats, { refetch: refetchStats }] = createResource(() => api.get<any>('/admin/stats'));
  const [products, { refetch: refetchProducts }] = createResource(() => api.get<{ products: any[] }>('/products?limit=100'));
  const [orders, { refetch: refetchOrders }] = createResource(() => api.get<{ orders: any[] }>('/admin/orders'));
  const [customers] = createResource(() => api.get<any[]>('/admin/customers'));
  const [media, { refetch: refetchMedia }] = createResource(() => api.get<any[]>('/admin/media'));
  const [widgets, { refetch: refetchWidgets }] = createResource(() => api.get<any[]>('/admin/widgets'));
  const [themes, { refetch: refetchThemes }] = createResource(() => api.get<any[]>('/admin/themes'));
  const [settings, { refetch: refetchSettings }] = createResource(() => api.get<Record<string, any>>('/admin/settings'));
  const [coupons, { refetch: refetchCoupons }] = createResource(() => api.get<any[]>('/admin/coupons'));
  const [taxes, { refetch: refetchTaxes }] = createResource(() => api.get<any[]>('/admin/taxes'));
  const [shippingMethods, { refetch: refetchShipping }] = createResource(() => api.get<any[]>('/admin/shipping-methods'));
  const [apiKeys, { refetch: refetchApiKeys }] = createResource(() => api.get<any[]>('/admin/api-keys'));

  const [analytics] = createResource(
    () => analyticsDays(),
    (d) => api.get<any>(`/admin/analytics/overview?days=${d}`)
  );

  // Instellingen Form states
  const [storeName, setStoreName] = createSignal('Aesthetic Apparel');
  const [storeTagline, setStoreTagline] = createSignal('Minimalistisch design');
  const [contactEmail, setContactEmail] = createSignal('contact@aesthetic.be');
  const [storeAddress, setStoreAddress] = createSignal('Brussel, België');
  const [emailSenderName, setEmailSenderName] = createSignal('Aesthetic Studios');
  const [emailSenderAddress, setEmailSenderAddress] = createSignal('noreply@aesthetic.be');
  const [smtpHost, setSmtpHost] = createSignal('');
  const [smtpPort, setSmtpPort] = createSignal('587');
  const [smtpUser, setSmtpUser] = createSignal('');
  const [smtpPass, setSmtpPass] = createSignal('');
  const [mockPaymentsEnabled, setMockPaymentsEnabled] = createSignal(true);
  const [mollieKey, setMollieKey] = createSignal('');
  const [stripePub, setStripePub] = createSignal('');
  const [stripeSec, setStripeSec] = createSignal('');
  const [gelatoKey, setGelatoKey] = createSignal('');

  const initSettings = () => {
    const s = settings();
    if (!s) return;
    try {
      if (s.store_name) setStoreName(JSON.parse(s.store_name));
      if (s.store_tagline) setStoreTagline(JSON.parse(s.store_tagline));
      if (s.contact_email) setContactEmail(JSON.parse(s.contact_email));
      if (s.store_address) setStoreAddress(JSON.parse(s.store_address));
      if (s.email_sender_name) setEmailSenderName(JSON.parse(s.email_sender_name));
      if (s.email_sender_address) setEmailSenderAddress(JSON.parse(s.email_sender_address));
      if (s.smtp_host) setSmtpHost(JSON.parse(s.smtp_host));
      if (s.smtp_port) setSmtpPort(JSON.parse(s.smtp_port));
      if (s.smtp_username) setSmtpUser(JSON.parse(s.smtp_username));
      if (s.smtp_password) setSmtpPass(JSON.parse(s.smtp_password));
      if (s.mock_payments_enabled) setMockPaymentsEnabled(JSON.parse(s.mock_payments_enabled));
      if (s.mollie_api_key) setMollieKey(JSON.parse(s.mollie_api_key));
      if (s.stripe_publishable_key) setStripePub(JSON.parse(s.stripe_publishable_key));
      if (s.stripe_secret_key) setStripeSec(JSON.parse(s.stripe_secret_key));
      if (s.gelato_api_key) setGelatoKey(JSON.parse(s.gelato_api_key));
    } catch {}
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedUrl(text);
    setTimeout(() => setCopiedUrl(null), 2000);
  };

  const copyMediaUrl = (url: string) => {
    navigator.clipboard.writeText(url);
    setCopiedUrl(url);
    setTimeout(() => setCopiedUrl(null), 2000);
  };

  const handleMediaUpload = async (e: Event) => {
    const target = e.target as HTMLInputElement;
    if (!target.files || target.files.length === 0) return;
    const formData = new FormData();
    formData.append('file', target.files[0]);
    try {
      await api.post('/admin/upload', formData);
      refetchMedia();
      alert('Afbeelding geüpload!');
    } catch (err: any) {
      alert('Upload fout: ' + err.message);
    }
  };

  const handleDeleteMedia = async (filename: string) => {
    if (!confirm('Bestand definitief verwijderen?')) return;
    try {
      await api.delete(`/admin/media/${filename}`);
      refetchMedia();
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  const openNewProduct = () => {
    setEditingProductId(null);
    setProdName('');
    setProdSlug('');
    setProdDesc('');
    setProdPrice(39.95);
    setProdFeatured(false);
    setProdType('variable');
    setCurrentVariants([]);
    setCurrentImages([]);
    setModalTab('general');
    setShowProductModal(true);
  };

  const openEditProduct = async (p: any) => {
    setEditingProductId(p.id);
    setProdName(p.name);
    setProdSlug(p.slug);
    setProdDesc(p.description || '');
    setProdPrice(p.base_price);
    setProdFeatured(p.featured);
    setProdType(p.product_type || 'variable');

    try {
      const full = await api.get<any>(`/products/${p.slug}`);
      setCurrentVariants(full.variants || []);
      setCurrentImages(full.images || []);
    } catch {
      setCurrentVariants([]);
      setCurrentImages([]);
    }

    setModalTab('general');
    setShowProductModal(true);
  };

  const handleSaveGeneral = async (e: Event) => {
    e.preventDefault();
    try {
      if (editingProductId()) {
        await api.put(`/admin/products/${editingProductId()}`, {
          name: prodName(),
          slug: prodSlug() || prodName().toLowerCase().replace(/\s+/g, '-'),
          description: prodDesc(),
          base_price: Number(prodPrice()),
          featured: prodFeatured(),
          is_active: true,
        });
        alert(t('save_changes') + ' OK!');
      } else {
        const p = await api.post<any>('/admin/products', {
          name: prodName(),
          slug: prodSlug() || prodName().toLowerCase().replace(/\s+/g, '-'),
          description: prodDesc(),
          base_price: Number(prodPrice()),
          featured: prodFeatured(),
          is_active: true,
        });
        setEditingProductId(p.id);
        alert(t('save_changes') + ' OK!');
      }
      refetchProducts();
      refetchStats();
    } catch (err: any) {
      alert('Error: ' + err.message);
    }
  };

  const handleDeleteProduct = async (id: string) => {
    if (!confirm('Product definitief verwijderen?')) return;
    try {
      await api.delete(`/admin/products/${id}`);
      refetchProducts();
      refetchStats();
    } catch (err: any) {
      alert('Error: ' + err.message);
    }
  };

  const handleAddVariant = async () => {
    if (!editingProductId()) return;
    try {
      const sku = `${prodSlug().toUpperCase()}-${newVarSize()}-${newVarColor()}-${Date.now().toString().slice(-4)}`;
      await api.post('/admin/products/variants', {
        product_id: editingProductId(),
        sku,
        title: `${newVarSize()} / ${newVarColor()}`,
        size: newVarSize(),
        color: newVarColor(),
        price_override: newVarPrice() ? Number(newVarPrice()) : null,
        stock_quantity: Number(newVarStock()),
        image_url: newVarImage() || null,
        pod_provider: newVarPODProvider() !== 'none' ? newVarPODProvider() : null,
        pod_variant_id: newVarPODID() || null,
        print_file_url: newVarPrintFile() || null,
        is_active: true,
      });

      const full = await api.get<any>(`/products/${prodSlug()}`);
      setCurrentVariants(full.variants || []);
      setNewVarImage('');
      alert(`Variant ${newVarSize()} toegevoegd!`);
      refetchProducts();
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  const handleDeleteVariant = async (variantId: string) => {
    if (!confirm('Variant verwijderen?')) return;
    try {
      await api.delete(`/admin/products/variants/${variantId}`);
      setCurrentVariants((prev) => prev.filter((v) => v.id !== variantId));
      refetchProducts();
    } catch (err: any) {
      alert('Fout bij verwijderen variant: ' + err.message);
    }
  };

  const handleAddImage = async () => {
    if (!editingProductId() || !newImageUrl()) return;
    try {
      await api.post('/admin/products/images', {
        product_id: editingProductId(),
        variant_id: newImageVariantId() || null,
        url: newImageUrl(),
        is_primary: newImageIsPrimary(),
      });
      const full = await api.get<any>(`/products/${prodSlug()}`);
      setCurrentImages(full.images || []);
      setNewImageUrl('');
      refetchProducts();
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  const handleDeleteImage = async (imageId: string) => {
    if (!confirm('Afbeelding verwijderen?')) return;
    try {
      await api.delete(`/admin/products/images/${imageId}`);
      setCurrentImages((prev) => prev.filter((img) => img.id !== imageId));
      refetchProducts();
    } catch (err: any) {
      alert('Fout bij verwijderen afbeelding: ' + err.message);
    }
  };

  const handleActivateTheme = async (themeId: string) => {
    try {
      await api.post('/admin/themes/activate', { theme_id: themeId });
      refetchThemes();
      alert(t('theme_activated'));
    } catch (err: any) {
      alert('Error: ' + err.message);
    }
  };

  const handleSaveSettings = async (e: Event) => {
    e.preventDefault();
    try {
      await api.put('/admin/settings', {
        store_name: storeName(),
        store_tagline: storeTagline(),
        contact_email: contactEmail(),
        store_address: storeAddress(),
        email_sender_name: emailSenderName(),
        email_sender_address: emailSenderAddress(),
        smtp_host: smtpHost(),
        smtp_port: smtpPort(),
        smtp_username: smtpUser(),
        smtp_password: smtpPass(),
        mock_payments_enabled: mockPaymentsEnabled(),
        mollie_api_key: mollieKey(),
        stripe_publishable_key: stripePub(),
        stripe_secret_key: stripeSec(),
        gelato_api_key: gelatoKey(),
      });
      alert('Winkelinstellingen succesvol opgeslagen!');
      refetchSettings();
    } catch (err: any) {
      alert('Fout bij opslaan: ' + err.message);
    }
  };

  // Gelato POD Catalogus Sync
  const handleSyncGelato = async () => {
    setSyncingPOD(true);
    try {
      const res = await api.post<any>('/admin/pod/sync');
      alert(`Synchronisatie geslaagd! ${res.synced_products} Gelato producten/varianten bijgewerkt in je winkel.`);
      refetchProducts();
    } catch (err: any) {
      alert('Fout bij synchroniseren: ' + err.message);
    } finally {
      setSyncingPOD(false);
    }
  };

  // Test E-mail
  const handleSendTestEmail = async () => {
    if (!testEmailTo()) {
      alert('Vul een ontvangend e-mailadres in');
      return;
    }
    try {
      await api.post('/admin/email/test', { email: testEmailTo() });
      alert('Test e-mail verzonden naar ' + testEmailTo());
    } catch (err: any) {
      alert('Verzendfout: ' + err.message);
    }
  };

  // Btw Tarief toevoegen
  const handleAddTaxRate = async (e: Event) => {
    e.preventDefault();
    try {
      await api.post('/admin/taxes', {
        country_code: newTaxCountry(),
        name: newTaxName(),
        rate: Number(newTaxRate()),
      });
      refetchTaxes();
      alert('Btw-tarief toegevoegd!');
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  const handleDeleteTaxRate = async (id: string) => {
    if (!confirm('Tarief verwijderen?')) return;
    try {
      await api.delete(`/admin/taxes/${id}`);
      refetchTaxes();
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  // Verzendmethode toevoegen
  const handleAddShipping = async (e: Event) => {
    e.preventDefault();
    if (!newShipTitle()) return;
    try {
      await api.post('/admin/shipping-methods', {
        title: newShipTitle(),
        cost: Number(newShipCost()),
        free_threshold: newShipThreshold() ? Number(newShipThreshold()) : null,
        is_active: true,
      });
      setNewShipTitle('');
      refetchShipping();
      alert('Verzendmethode toegevoegd!');
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  const handleDeleteShipping = async (id: string) => {
    if (!confirm('Verzendoptie verwijderen?')) return;
    try {
      await api.delete(`/admin/shipping-methods/${id}`);
      refetchShipping();
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  // API Key Aanmaken
  const handleCreateApiKey = async (e: Event) => {
    e.preventDefault();
    try {
      const res = await api.post<any>('/admin/api-keys', {
        description: newKeyDesc(),
        permissions: newKeyPerms(),
      });
      setGeneratedKeyResult({
        ck: res.consumer_key,
        cs: res.consumer_secret,
      });
      refetchApiKeys();
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  const handleDeleteApiKey = async (id: string) => {
    if (!confirm('API sleutel intrekken? Externe apps verliezen direct toegang.')) return;
    try {
      await api.delete(`/admin/api-keys/${id}`);
      refetchApiKeys();
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  const openNewWidget = () => {
    setEditingWidgetId(null);
    setWidgetType('hero');
    setWidgetTitle('');
    setWidgetSubtitle('');
    setWidgetOrder(widgets() ? widgets()!.length + 1 : 1);
    setWidgetActive(true);
    setWidgetBgImage('');
    setWidgetBtnText('Bekijk Collectie');
    setWidgetBtnLink('/shop');
    setShowWidgetModal(true);
  };

  const openEditWidget = (w: any) => {
    setEditingWidgetId(w.id);
    setWidgetType(w.type);
    setWidgetTitle(w.title || '');
    setWidgetSubtitle(w.subtitle || '');
    setWidgetOrder(w.sort_order || 1);
    setWidgetActive(w.is_active);

    try {
      const cfg = typeof w.config === 'string' ? JSON.parse(w.config) : w.config || {};
      setWidgetBgImage(cfg.bg_image || '');
      setWidgetBtnText(cfg.btn_text || '');
      setWidgetBtnLink(cfg.btn_link || '');
    } catch {
      setWidgetBgImage('');
      setWidgetBtnText('');
      setWidgetBtnLink('');
    }

    setShowWidgetModal(true);
  };

  const handleSaveWidget = async (e: Event) => {
    e.preventDefault();
    try {
      const configObj: Record<string, any> = {};
      if (widgetBgImage()) configObj.bg_image = widgetBgImage();
      if (widgetBtnText()) configObj.btn_text = widgetBtnText();
      if (widgetBtnLink()) configObj.btn_link = widgetBtnLink();

      await api.post('/admin/widgets', {
        id: editingWidgetId() || undefined,
        type: widgetType(),
        title: widgetTitle(),
        subtitle: widgetSubtitle(),
        config: configObj,
        sort_order: Number(widgetOrder()),
        is_active: widgetActive(),
      });

      setShowWidgetModal(false);
      refetchWidgets();
      alert('Widget opgeslagen!');
    } catch (err: any) {
      alert('Fout bij opslaan widget: ' + err.message);
    }
  };

  const handleDeleteWidget = async (id: string) => {
    if (!confirm('Widget verwijderen?')) return;
    try {
      await api.delete(`/admin/widgets/${id}`);
      refetchWidgets();
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  const handleToggleWidget = async (w: any) => {
    try {
      await api.post('/admin/widgets', {
        id: w.id,
        type: w.type,
        title: w.title,
        subtitle: w.subtitle,
        config: w.config,
        sort_order: w.sort_order,
        is_active: !w.is_active,
      });
      refetchWidgets();
    } catch (err: any) {
      alert('Error: ' + err.message);
    }
  };

  const handleCreateCoupon = async (e: Event) => {
    e.preventDefault();
    try {
      await api.post('/admin/coupons', {
        code: couponCode(),
        discount_type: couponType(),
        discount_value: Number(couponValue()),
        min_spend: Number(couponMinSpend()),
        is_active: true,
      });
      setShowCouponModal(false);
      refetchCoupons();
      alert('Kortingscode aangemaakt!');
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  const handleDeleteCoupon = async (id: string) => {
    if (!confirm('Kortingscode verwijderen?')) return;
    try {
      await api.delete(`/admin/coupons/${id}`);
      refetchCoupons();
    } catch (err: any) {
      alert('Fout: ' + err.message);
    }
  };

  return (
    <div class="admin-page container">
      <div class="admin-header">
        <div>
          <h1>{t('admin_title')}</h1>
          <p style={{ color: 'var(--text-muted)' }}>{t('admin_desc')}</p>
        </div>
        <button class="glass-button" onClick={openNewProduct}>
          <Plus size={18} /> {t('new_product')}
        </button>
      </div>

      {/* KPI Kaarten */}
      <div class="stats-grid">
        <div class="stat-card glass-panel">
          <span class="stat-label">{t('net_revenue')}</span>
          <span class="stat-val">{formatPrice(stats()?.total_revenue ?? 0)}</span>
        </div>
        <div class="stat-card glass-panel">
          <span class="stat-label">{t('total_orders')}</span>
          <span class="stat-val">{stats()?.total_orders ?? 0}</span>
        </div>
        <div class="stat-card glass-panel">
          <span class="stat-label">{t('customers')}</span>
          <span class="stat-val">{stats()?.total_customers ?? 0}</span>
        </div>
        <div class="stat-card glass-panel">
          <span class="stat-label">{t('low_stock')}</span>
          <span class="stat-val" style={{ color: 'var(--danger)' }}>{stats()?.low_stock_count ?? 0}</span>
        </div>
      </div>

      {/* Hoofdtabs */}
      <div class="admin-tabs">
        <button class="tab-item" classList={{ active: activeTab() === 'products' }} onClick={() => setActiveTab('products')}>
          {t('tab_products')} ({products()?.products?.length ?? 0})
        </button>
        <button class="tab-item" classList={{ active: activeTab() === 'orders' }} onClick={() => setActiveTab('orders')}>
          {t('tab_orders')} ({orders()?.orders?.length ?? 0})
        </button>
        <button class="tab-item" classList={{ active: activeTab() === 'analytics' }} onClick={() => setActiveTab('analytics')}>
          <BarChart3 size={15} style={{ display: 'inline', 'margin-right': '4px' }} /> {t('tab_analytics')}
        </button>
        <button class="tab-item" classList={{ active: activeTab() === 'customers' }} onClick={() => setActiveTab('customers')}>
          <Users size={15} style={{ display: 'inline', 'margin-right': '4px' }} /> {t('tab_customers')} ({customers()?.length ?? 0})
        </button>
        <button class="tab-item" classList={{ active: activeTab() === 'media' }} onClick={() => setActiveTab('media')}>
          <ImageIcon size={15} style={{ display: 'inline', 'margin-right': '4px' }} /> {t('tab_media')} ({media()?.length ?? 0})
        </button>
        <button class="tab-item" classList={{ active: activeTab() === 'coupons' }} onClick={() => setActiveTab('coupons')}>
          {t('tab_coupons')} ({coupons()?.length ?? 0})
        </button>
        <button class="tab-item" classList={{ active: activeTab() === 'widgets' }} onClick={() => setActiveTab('widgets')}>
          {t('tab_widgets')} ({widgets()?.length ?? 0})
        </button>
        <button class="tab-item" classList={{ active: activeTab() === 'themes' }} onClick={() => setActiveTab('themes')}>
          {t('tab_themes')} ({themes()?.length ?? 0})
        </button>
        <button class="tab-item" classList={{ active: activeTab() === 'settings' }} onClick={() => { initSettings(); setActiveTab('settings'); }}>
          {t('tab_settings')}
        </button>
      </div>

      {/* Tab 1: Producten */}
      <Show when={activeTab() === 'products'}>
        <div class="admin-table-container glass-panel">
          <table>
            <thead>
              <tr>
                <th>{t('col_image')}</th>
                <th>{t('col_name')}</th>
                <th>{t('col_slug')}</th>
                <th>{t('col_price')}</th>
                <th>{t('col_featured')}</th>
                <th>{t('col_actions')}</th>
              </tr>
            </thead>
            <tbody>
              <For each={products()?.products}>
                {(p) => (
                  <tr>
                    <td><img src={p.primary_image || '/static/tshirt-placeholder.jpg'} alt="" style={{ width: '44px', height: '44px', 'border-radius': '8px', 'object-fit': 'cover' }} /></td>
                    <td><strong>{p.name}</strong><div style={{ "font-size": "0.78rem", color: "var(--text-muted)" }}>{p.category_name || 'General'}</div></td>
                    <td><code>{p.slug}</code></td>
                    <td><strong>{formatPrice(p.base_price)}</strong></td>
                    <td>{p.featured ? `⭐ ${t('yes')}` : t('no')}</td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button class="glass-button secondary" style={{ padding: '0.35rem 0.65rem', 'font-size': '0.8rem' }} onClick={() => openEditProduct(p)} title={t('edit')}>
                          <Edit3 size={14} /> {t('edit')}
                        </button>
                        <button style={{ color: 'var(--danger)', padding: '0.35rem' }} onClick={() => handleDeleteProduct(p.id)} title={t('delete')}>
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                )}
              </For>
            </tbody>
          </table>
        </div>
      </Show>

      {/* Tab 2: Bestellingen */}
      <Show when={activeTab() === 'orders'}>
        <div class="admin-table-container glass-panel">
          <table>
            <thead><tr><th>Order #</th><th>Datum</th><th>Totaal</th><th>Betaling</th><th>Status</th><th>Actie</th></tr></thead>
            <tbody>
              <For each={orders()?.orders}>
                {(o) => (
                  <tr>
                    <td><strong>{o.order_number}</strong></td>
                    <td>{new Date(o.created_at).toLocaleDateString()}</td>
                    <td><strong>{formatPrice(o.total_amount)}</strong></td>
                    <td><span class="glass-badge">{o.payment_status}</span></td>
                    <td>
                      <select class="glass-input" style={{ width: 'auto', padding: '0.3rem 0.6rem', 'font-size': '0.85rem' }} value={o.status} onChange={(e) => api.put(`/admin/orders/${o.id}/status`, { status: e.currentTarget.value, message: 'Status gewijzigd' }).then(() => refetchOrders())}>
                        <option value="pending">In afwachting</option>
                        <option value="processing">In verwerking</option>
                        <option value="shipped">Verzonden</option>
                        <option value="delivered">Afgeleverd</option>
                        <option value="cancelled">Geannuleerd</option>
                      </select>
                    </td>
                    <td>
                      <a href={`/track/${o.order_number}`} target="_blank" class="glass-button secondary" style={{ padding: '0.25rem 0.6rem', 'font-size': '0.8rem' }}>{t('details')}</a>
                    </td>
                  </tr>
                )}
              </For>
            </tbody>
          </table>
        </div>
      </Show>

      {/* Tab 3: Rapportages & Analytics */}
      <Show when={activeTab() === 'analytics'}>
        <div class="analytics-container">
          <div class="analytics-toolbar">
            <div class="time-pills">
              <button classList={{ active: analyticsDays() === 7 }} onClick={() => setAnalyticsDays(7)}>{t('last_7_days')}</button>
              <button classList={{ active: analyticsDays() === 30 }} onClick={() => setAnalyticsDays(30)}>{t('last_30_days')}</button>
              <button classList={{ active: analyticsDays() === 90 }} onClick={() => setAnalyticsDays(90)}>{t('last_90_days')}</button>
            </div>
            <a href="/api/admin/analytics/export-csv" target="_blank" class="glass-button secondary" style={{ "font-size": "0.85rem" }}>
              <Download size={15} /> {t('export_csv')}
            </a>
          </div>

          <div class="stats-grid" style={{ "margin-bottom": "0" }}>
            <div class="stat-card glass-panel"><span class="stat-label">{t('net_revenue')}</span><span class="stat-val">{formatPrice(analytics()?.total_revenue ?? 0)}</span></div>
            <div class="stat-card glass-panel"><span class="stat-label">{t('avg_order_value')}</span><span class="stat-val">{formatPrice(analytics()?.average_order ?? 0)}</span></div>
            <div class="stat-card glass-panel"><span class="stat-label">{t('items_sold')}</span><span class="stat-val">{analytics()?.items_sold ?? 0}</span></div>
            <div class="stat-card glass-panel"><span class="stat-label">{t('total_orders')}</span><span class="stat-val">{analytics()?.total_orders ?? 0}</span></div>
          </div>

          <div class="chart-card glass-panel">
            <div class="chart-header">
              <h3>{t('revenue_chart_title')}</h3>
              <span style={{ "font-size": "0.8rem", color: "var(--text-muted)" }}>{analytics()?.daily_sales?.length ?? 0} datapunten</span>
            </div>
            <div class="svg-chart-wrap">
              <Show when={analytics()?.daily_sales && analytics()!.daily_sales.length > 0} fallback={<p style={{ "text-align": "center", color: "var(--text-muted)", "padding-top": "80px" }}>Nog geen verkoopdata.</p>}>
                <svg viewBox="0 0 100 100" preserveAspectRatio="none">
                  <For each={analytics()?.daily_sales}>
                    {(d: any, i) => {
                      const maxRev = Math.max(...(analytics()?.daily_sales?.map((x: any) => x.revenue) || [10]), 10);
                      const barWidth = 100 / (analytics()?.daily_sales?.length || 1);
                      const height = (d.revenue / maxRev) * 80;
                      return (
                        <rect x={i() * barWidth + (barWidth * 0.15)} y={90 - height} width={barWidth * 0.7} height={height} rx="1.5" class="bar">
                          <title>{`${d.date}: ${formatPrice(d.revenue)} (${d.orders} orders)`}</title>
                        </rect>
                      );
                    }}
                  </For>
                </svg>
              </Show>
            </div>
          </div>
        </div>
      </Show>

      {/* Tab 4: Klanten CRM */}
      <Show when={activeTab() === 'customers'}>
        <div class="admin-table-container glass-panel">
          <table>
            <thead><tr><th>Klantnaam</th><th>E-mailadres</th><th>{t('total_orders')}</th><th>{t('customer_lifetime_value')}</th><th>{t('registered_on')}</th></tr></thead>
            <tbody>
              <For each={customers()} fallback={<tr><td colspan="5" style={{ "text-align": "center", color: "var(--text-muted)" }}>{t('no_customers_found')}</td></tr>}>
                {(c) => (
                  <tr>
                    <td><strong>{c.full_name}</strong></td>
                    <td><code>{c.email}</code></td>
                    <td><span class="glass-badge">{c.total_orders} orders</span></td>
                    <td><strong>{formatPrice(c.total_spent)}</strong></td>
                    <td>{new Date(c.registered_at).toLocaleDateString()}</td>
                  </tr>
                )}
              </For>
            </tbody>
          </table>
        </div>
      </Show>

      {/* Tab 5: Media Bibliotheek */}
      <Show when={activeTab() === 'media'}>
        <div class="media-library-wrap">
          <label class="upload-zone glass-panel">
            <UploadCloud size={36} color="var(--accent)" />
            <span style={{ "font-weight": 700 }}>Klik hier om een nieuwe afbeelding te uploaden</span>
            <input type="file" accept="image/*" onChange={handleMediaUpload} />
          </label>
          <div class="media-grid">
            <For each={media()} fallback={<p style={{ color: "var(--text-muted)", "text-align": "center" }}>{t('no_media_found')}</p>}>
              {(m) => (
                <div class="media-item-card glass-panel">
                  <div class="thumb-wrap"><img src={m.url} alt={m.name} /></div>
                  <div class="media-meta">
                    <span class="file-name" title={m.name}>{m.name}</span>
                    <div class="media-btn-row">
                      <button class="glass-button secondary" style={{ padding: '0.25rem 0.5rem', 'font-size': '0.75rem' }} onClick={() => copyMediaUrl(m.url)}>
                        {copiedUrl() === m.url ? <Check size={12} color="var(--success)" /> : <Copy size={12} />}
                      </button>
                      <button style={{ color: 'var(--danger)', background: 'none', border: 'none', cursor: 'pointer' }} onClick={() => handleDeleteMedia(m.name)}><Trash2 size={15} /></button>
                    </div>
                  </div>
                </div>
              )}
            </For>
          </div>
        </div>
      </Show>

      {/* Tab 6: Kortingscodes */}
      <Show when={activeTab() === 'coupons'}>
        <div style={{ display: 'flex', 'flex-direction': 'column', gap: '1.25rem' }}>
          <div style={{ display: 'flex', 'justify-content': 'flex-end' }}>
            <button class="glass-button" onClick={() => setShowCouponModal(true)}><Plus size={16} /> Nieuwe Kortingscode</button>
          </div>
          <div class="admin-table-container glass-panel">
            <table>
              <thead><tr><th>Code</th><th>Type</th><th>Waarde</th><th>Min. Besteding</th><th>Gebruikt</th><th>Actie</th></tr></thead>
              <tbody>
                <For each={coupons()} fallback={<tr><td colspan="6" style={{ "text-align": "center", color: "var(--text-muted)" }}>Geen kortingscodes.</td></tr>}>
                  {(c) => (
                    <tr>
                      <td><strong><code>{c.code}</code></strong></td>
                      <td>{c.discount_type === 'percent' ? 'Percentage' : 'Vast bedrag'}</td>
                      <td><strong>{c.discount_type === 'percent' ? `${c.discount_value}%` : formatPrice(c.discount_value)}</strong></td>
                      <td>{c.min_spend ? formatPrice(c.min_spend) : 'Geen'}</td>
                      <td>{c.uses_count} keer</td>
                      <td><button style={{ color: 'var(--danger)', background: 'none', border: 'none', cursor: 'pointer' }} onClick={() => handleDeleteCoupon(c.id)}><Trash2 size={16} /></button></td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        </div>
      </Show>

      {/* Tab 7: Widgets */}
      <Show when={activeTab() === 'widgets'}>
        <div style={{ display: 'flex', 'flex-direction': 'column', gap: '1.25rem' }}>
          <div style={{ display: 'flex', 'justify-content': 'flex-end' }}>
            <button class="glass-button" onClick={openNewWidget}><Plus size={16} /> {t('add_widget')}</button>
          </div>
          <div class="admin-table-container glass-panel">
            <table>
              <thead><tr><th>{t('widget_type')}</th><th>{t('widget_title')}</th><th>{t('widget_order')}</th><th>{t('widget_status')}</th><th>{t('col_actions')}</th></tr></thead>
              <tbody>
                <For each={widgets()}>
                  {(w) => (
                    <tr>
                      <td><code>{w.type}</code></td>
                      <td><strong>{w.title || '-'}</strong></td>
                      <td>{w.sort_order}</td>
                      <td><span class="glass-badge" style={{ background: w.is_active ? 'var(--success)' : 'var(--danger)', color: '#fff' }}>{w.is_active ? t('active') : t('inactive')}</span></td>
                      <td>
                        <div style={{ display: 'flex', gap: '0.4rem' }}>
                          <button class="glass-button secondary" style={{ padding: '0.3rem 0.6rem', 'font-size': '0.78rem' }} onClick={() => openEditWidget(w)}><Edit3 size={13} /> {t('edit')}</button>
                          <button class="glass-button secondary" style={{ padding: '0.3rem 0.6rem', 'font-size': '0.78rem' }} onClick={() => handleToggleWidget(w)}><RefreshCw size={13} /> Status</button>
                          <button style={{ color: 'var(--danger)', background: 'none', border: 'none', cursor: 'pointer', padding: '0.3rem' }} onClick={() => handleDeleteWidget(w.id)}><Trash2 size={15} /></button>
                        </div>
                      </td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        </div>
      </Show>

      {/* Tab 8: Thema's */}
      <Show when={activeTab() === 'themes'}>
        <div class="themes-grid">
          <For each={themes()}>
            {(th) => (
              <div class="theme-card glass-panel" classList={{ 'active-theme': th.is_active }}>
                <div class="theme-preview"><Palette size={48} /></div>
                <div class="theme-info">
                  <div class="theme-header"><h3>{th.name}</h3><span class="version-tag">v{th.version}</span></div>
                  <p>{th.description}</p>
                  <div class="theme-actions">
                    <Show when={th.is_active} fallback={<button class="glass-button" style={{ 'font-size': '0.85rem', padding: '0.4rem 0.9rem' }} onClick={() => handleActivateTheme(th.id)}>{t('activate')}</button>}>
                      <span class="glass-badge" style={{ background: 'var(--success)', color: '#fff' }}>✓ {t('active_theme')}</span>
                    </Show>
                  </div>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>

      {/* Tab 9: Instellingen */}
      <Show when={activeTab() === 'settings'}>
        <div style={{ display: 'flex', 'flex-direction': 'column', gap: '1.5rem' }}>
          <div class="settings-subnav">
            <button classList={{ active: settingsSubTab() === 'general' }} onClick={() => setSettingsSubTab('general')}><Building2 size={14} style={{ display: 'inline', 'margin-right': '4px' }} /> Algemeen</button>
            <button classList={{ active: settingsSubTab() === 'shipping' }} onClick={() => setSettingsSubTab('shipping')}><Truck size={14} style={{ display: 'inline', 'margin-right': '4px' }} /> Verzendmethoden ({shippingMethods()?.length ?? 0})</button>
            <button classList={{ active: settingsSubTab() === 'taxes' }} onClick={() => setSettingsSubTab('taxes')}><Percent size={14} style={{ display: 'inline', 'margin-right': '4px' }} /> Btw-Tarieven ({taxes()?.length ?? 0})</button>
            <button classList={{ active: settingsSubTab() === 'payments' }} onClick={() => setSettingsSubTab('payments')}><CreditCard size={14} style={{ display: 'inline', 'margin-right': '4px' }} /> Betaalproviders</button>
            <button classList={{ active: settingsSubTab() === 'pod' }} onClick={() => setSettingsSubTab('pod')}><Printer size={14} style={{ display: 'inline', 'margin-right': '4px' }} /> Gelato & POD</button>
            <button classList={{ active: settingsSubTab() === 'email' }} onClick={() => setSettingsSubTab('email')}><Mail size={14} style={{ display: 'inline', 'margin-right': '4px' }} /> E-mail (SMTP)</button>
            <button classList={{ active: settingsSubTab() === 'apikeys' }} onClick={() => setSettingsSubTab('apikeys')}><Key size={14} style={{ display: 'inline', 'margin-right': '4px' }} /> WooCommerce API ({apiKeys()?.length ?? 0})</button>
          </div>

          <Show when={settingsSubTab() === 'general'}>
            <form onSubmit={handleSaveSettings} class="settings-container">
              <div class="settings-card glass-panel">
                <h3>Winkelgegevens</h3>
                <div class="settings-fields">
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Winkelnaam</label><input type="text" class="glass-input" value={storeName()} onInput={(e) => setStoreName(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Slogan</label><input type="text" class="glass-input" value={storeTagline()} onInput={(e) => setStoreTagline(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Contact E-mail</label><input type="email" class="glass-input" value={contactEmail()} onInput={(e) => setContactEmail(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Adres</label><input type="text" class="glass-input" value={storeAddress()} onInput={(e) => setStoreAddress(e.currentTarget.value)} /></div>
                </div>
              </div>
              <button type="submit" class="glass-button" style={{ 'align-self': 'flex-start' }}>Opslaan</button>
            </form>
          </Show>

          <Show when={settingsSubTab() === 'shipping'}>
            <div style={{ display: 'flex', 'flex-direction': 'column', gap: '1.5rem' }}>
              <div class="admin-table-container glass-panel">
                <table>
                  <thead><tr><th>Verzendoptie</th><th>Kosten</th><th>Gratis Vanaf</th><th>Status</th><th>Actie</th></tr></thead>
                  <tbody>
                    <For each={shippingMethods()}>
                      {(sm) => (
                        <tr>
                          <td><strong>{sm.title}</strong></td>
                          <td><strong>{formatPrice(sm.cost)}</strong></td>
                          <td>{sm.free_threshold ? formatPrice(sm.free_threshold) : 'Geen drempel'}</td>
                          <td><span class="glass-badge" style={{ background: sm.is_active ? 'var(--success)' : 'var(--danger)', color: '#fff' }}>{sm.is_active ? 'Actief' : 'Inactief'}</span></td>
                          <td><button style={{ color: 'var(--danger)', background: 'none', border: 'none', cursor: 'pointer' }} onClick={() => handleDeleteShipping(sm.id)}><Trash2 size={16} /></button></td>
                        </tr>
                      )}
                    </For>
                  </tbody>
                </table>
              </div>
              <form onSubmit={handleAddShipping} class="settings-card glass-panel" style={{ 'max-width': '650px' }}>
                <h3>Nieuwe Verzendoptie</h3>
                <div style={{ display: 'grid', 'grid-template-columns': '2fr 1fr 1fr', gap: '0.75rem' }}>
                  <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Titel</label><input type="text" required class="glass-input" value={newShipTitle()} onInput={(e) => setNewShipTitle(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Kosten (€)</label><input type="number" step="0.05" class="glass-input" value={newShipCost()} onInput={(e) => setNewShipCost(Number(e.currentTarget.value))} /></div>
                  <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Gratis Vanaf (€)</label><input type="number" step="1" class="glass-input" value={newShipThreshold() ?? ''} onInput={(e) => setNewShipThreshold(e.currentTarget.value ? Number(e.currentTarget.value) : null)} /></div>
                </div>
                <button type="submit" class="glass-button" style={{ 'align-self': 'flex-start', 'margin-top': '0.5rem' }}><Plus size={16} /> Opslaan</button>
              </form>
            </div>
          </Show>

          <Show when={settingsSubTab() === 'taxes'}>
            <div style={{ display: 'flex', 'flex-direction': 'column', gap: '1.5rem' }}>
              <div class="admin-table-container glass-panel">
                <table>
                  <thead><tr><th>Land</th><th>Omschrijving</th><th>Btw-tarief</th><th>Actie</th></tr></thead>
                  <tbody>
                    <For each={taxes()}>
                      {(tr) => (
                        <tr>
                          <td><strong><code>{tr.country_code}</code></strong></td>
                          <td>{tr.name}</td>
                          <td><strong>{tr.rate}%</strong></td>
                          <td><button style={{ color: 'var(--danger)', background: 'none', border: 'none', cursor: 'pointer' }} onClick={() => handleDeleteTaxRate(tr.id)}><Trash2 size={16} /></button></td>
                        </tr>
                      )}
                    </For>
                  </tbody>
                </table>
              </div>
              <form onSubmit={handleAddTaxRate} class="settings-card glass-panel" style={{ 'max-width': '650px' }}>
                <h3>Btw-tarief Toevoegen</h3>
                <div style={{ display: 'grid', 'grid-template-columns': '1fr 2fr 1fr', gap: '0.75rem' }}>
                  <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Land (ISO2)</label><input type="text" maxlength="2" required class="glass-input" value={newTaxCountry()} onInput={(e) => setNewTaxCountry(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Omschrijving</label><input type="text" required class="glass-input" value={newTaxName()} onInput={(e) => setNewTaxName(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Btw (%)</label><input type="number" step="0.1" required class="glass-input" value={newTaxRate()} onInput={(e) => setNewTaxRate(Number(e.currentTarget.value))} /></div>
                </div>
                <button type="submit" class="glass-button" style={{ 'align-self': 'flex-start', 'margin-top': '0.5rem' }}><Plus size={16} /> Opslaan</button>
              </form>
            </div>
          </Show>

          <Show when={settingsSubTab() === 'payments'}>
            <form onSubmit={handleSaveSettings} class="settings-container">
              <div class="settings-card glass-panel">
                <h3>Betaalproviders & Testmodus</h3>
                <div style={{ display: 'flex', 'align-items': 'center', gap: '0.75rem', padding: '1rem', background: 'var(--input-bg)', 'border-radius': '12px' }}>
                  <input type="checkbox" id="mock_cb" checked={mockPaymentsEnabled()} onChange={(e) => setMockPaymentsEnabled(e.currentTarget.checked)} style={{ width: '18px', height: '18px' }} />
                  <label for="mock_cb" style={{ "font-weight": 700 }}>Directe Testbetaler (Mock Gateway) Actief houden op Checkout</label>
                </div>
                <div class="settings-fields" style={{ 'margin-top': '1rem' }}>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Mollie API Key</label><input type="password" class="glass-input" placeholder="live_... / test_..." value={mollieKey()} onInput={(e) => setMollieKey(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Stripe Publishable Key</label><input type="text" class="glass-input" placeholder="pk_..." value={stripePub()} onInput={(e) => setStripePub(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Stripe Secret Key</label><input type="password" class="glass-input" placeholder="sk_..." value={stripeSec()} onInput={(e) => setStripeSec(e.currentTarget.value)} /></div>
                </div>
              </div>
              <button type="submit" class="glass-button" style={{ 'align-self': 'flex-start' }}>Opslaan</button>
            </form>
          </Show>

          <Show when={settingsSubTab() === 'pod'}>
            <div class="settings-container">
              <div class="settings-card glass-panel">
                <div style={{ display: 'flex', 'justify-content': 'space-between', 'align-items': 'center', 'flex-wrap': 'wrap', gap: '1rem' }}>
                  <div><h3>Gelato POD Koppeling</h3><p style={{ "font-size": "0.85rem", color: "var(--text-muted)" }}>Order-doorsturing en realtime transporttarieven.</p></div>
                  <button class="glass-button" disabled={syncingPOD()} onClick={handleSyncGelato}><RefreshCw size={16} classList={{ 'animate-spin': syncingPOD() }} /> {syncingPOD() ? 'Synchroniseren...' : 'Synchroniseer Catalogus met Gelato'}</button>
                </div>
                <div class="settings-fields" style={{ 'margin-top': '1rem' }}>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Gelato API Key</label><input type="password" class="glass-input" placeholder="Laat leeg voor veilige test-simulatie" value={gelatoKey()} onInput={(e) => setGelatoKey(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Webhook URL voor Gelato:</label><input type="text" readonly class="glass-input" value="https://jouwdomein.be/api/webhooks/gelato" style={{ opacity: 0.8 }} /></div>
                </div>
                <button class="glass-button secondary" style={{ 'align-self': 'flex-start', 'margin-top': '0.5rem' }} onClick={handleSaveSettings}>Sleutel Opslaan</button>
              </div>
            </div>
          </Show>

          <Show when={settingsSubTab() === 'email'}>
            <div class="settings-container">
              <form onSubmit={handleSaveSettings} class="settings-card glass-panel">
                <h3>SMTP E-mailserver Instellingen</h3>
                <div class="settings-fields">
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Afzendernaam</label><input type="text" class="glass-input" value={emailSenderName()} onInput={(e) => setEmailSenderName(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Afzender E-mail</label><input type="email" class="glass-input" value={emailSenderAddress()} onInput={(e) => setEmailSenderAddress(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>SMTP Host</label><input type="text" placeholder="smtp.gmail.com" class="glass-input" value={smtpHost()} onInput={(e) => setSmtpHost(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>SMTP Poort</label><input type="text" class="glass-input" value={smtpPort()} onInput={(e) => setSmtpPort(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>SMTP Gebruiker</label><input type="text" class="glass-input" value={smtpUser()} onInput={(e) => setSmtpUser(e.currentTarget.value)} /></div>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>SMTP Wachtwoord</label><input type="password" class="glass-input" value={smtpPass()} onInput={(e) => setSmtpPass(e.currentTarget.value)} /></div>
                </div>
                <button type="submit" class="glass-button" style={{ 'align-self': 'flex-start' }}>Opslaan</button>
              </form>
              <div class="settings-card glass-panel">
                <h3>Test E-mail Sturen</h3>
                <div style={{ display: 'flex', gap: '0.75rem', 'max-width': '480px' }}>
                  <input type="email" placeholder="jouw@email.com..." class="glass-input" value={testEmailTo()} onInput={(e) => setTestEmailTo(e.currentTarget.value)} />
                  <button class="glass-button" onClick={handleSendTestEmail}><Send size={15} /> Verzenden</button>
                </div>
              </div>
            </div>
          </Show>

          <Show when={settingsSubTab() === 'apikeys'}>
            <div style={{ display: 'flex', 'flex-direction': 'column', gap: '1.5rem' }}>
              <div style={{ display: 'flex', 'justify-content': 'space-between', 'align-items': 'center' }}>
                <p style={{ "font-size": "0.85rem", color: "var(--text-muted)", "max-width": "600px" }}>Gebruik deze sleutels in <strong>Gelato</strong>, <strong>Sendcloud</strong> of <strong>Printful</strong> via <em>"Verbind met WooCommerce"</em>.</p>
                <button class="glass-button" onClick={() => { setGeneratedKeyResult(null); setShowApiKeyModal(true); }}><Plus size={16} /> Nieuwe Sleutel</button>
              </div>
              <div class="admin-table-container glass-panel">
                <table>
                  <thead><tr><th>Omschrijving</th><th>Consumer Key</th><th>Rechten</th><th>Aangemaakt</th><th>Actie</th></tr></thead>
                  <tbody>
                    <For each={apiKeys()} fallback={<tr><td colspan="5" style={{ "text-align": "center", color: "var(--text-muted)" }}>Geen sleutels.</td></tr>}>
                      {(k) => (
                        <tr>
                          <td><strong>{k.description}</strong></td>
                          <td><code>{k.consumer_key.slice(0, 8)}...{k.consumer_key.slice(-4)}</code></td>
                          <td><span class="glass-badge">{k.permissions}</span></td>
                          <td>{new Date(k.created_at).toLocaleDateString()}</td>
                          <td><button style={{ color: 'var(--danger)', background: 'none', border: 'none', cursor: 'pointer' }} onClick={() => handleDeleteApiKey(k.id)}><Trash2 size={16} /></button></td>
                        </tr>
                      )}
                    </For>
                  </tbody>
                </table>
              </div>
            </div>
          </Show>
        </div>
      </Show>

      {/* API Key Modal */}
      <Show when={showApiKeyModal()}>
        <div class="wc-modal-backdrop" onClick={() => setShowApiKeyModal(false)}>
          <div class="wc-modal" style={{ "max-width": "550px" }} onClick={(e) => e.stopPropagation()}>
            <div class="modal-head"><h3>WooCommerce API Sleutel Genereren</h3><button class="close-btn" onClick={() => setShowApiKeyModal(false)}><X size={20} /></button></div>
            <Show when={generatedKeyResult()} fallback={
              <form onSubmit={handleCreateApiKey} style={{ display: 'flex', 'flex-direction': 'column', gap: '1rem' }}>
                <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Omschrijving *</label><input type="text" required class="glass-input" value={newKeyDesc()} onInput={(e) => setNewKeyDesc(e.currentTarget.value)} /></div>
                <div>
                  <label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Rechten</label>
                  <select class="glass-input" value={newKeyPerms()} onChange={(e) => setNewKeyPerms(e.currentTarget.value)}>
                    <option value="read_write">Lezen en Schrijven</option>
                    <option value="read">Alleen Lezen</option>
                    <option value="write">Alleen Schrijven</option>
                  </select>
                </div>
                <button type="submit" class="glass-button" style={{ 'margin-top': '0.5rem' }}>Genereer Sleutels</button>
              </form>
            }>
              <div style={{ display: 'flex', 'flex-direction': 'column', gap: '1rem' }}>
                <div style={{ padding: '0.85rem', background: 'rgba(234, 179, 8, 0.1)', 'border-radius': '10px', 'font-size': '0.82rem', color: '#b45309' }}>
                  ⚠️ Let op: Kopieer het Consumer Secret nu. Deze wordt om veiligheidsredenen niet meer getoond!
                </div>
                <div>
                  <label style={{ "font-size": "0.78rem", "font-weight": 700 }}>Consumer Key:</label>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <input type="text" readonly class="glass-input" value={generatedKeyResult()!.ck} />
                    <button class="glass-button secondary" onClick={() => copyToClipboard(generatedKeyResult()!.ck)}>{copiedUrl() === generatedKeyResult()!.ck ? <Check size={14} color="var(--success)" /> : <Copy size={14} />}</button>
                  </div>
                </div>
                <div>
                  <label style={{ "font-size": "0.78rem", "font-weight": 700 }}>Consumer Secret:</label>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <input type="text" readonly class="glass-input" value={generatedKeyResult()!.cs} />
                    <button class="glass-button secondary" onClick={() => copyToClipboard(generatedKeyResult()!.cs)}>{copiedUrl() === generatedKeyResult()!.cs ? <Check size={14} color="var(--success)" /> : <Copy size={14} />}</button>
                  </div>
                </div>
                <button class="glass-button" style={{ 'margin-top': '0.5rem' }} onClick={() => setShowApiKeyModal(false)}>Klaar</button>
              </div>
            </Show>
          </div>
        </div>
      </Show>

      {/* Product Editor Modal */}
      <Show when={showProductModal()}>
        <div class="wc-modal-backdrop" onClick={() => setShowProductModal(false)}>
          <div class="wc-modal" onClick={(e) => e.stopPropagation()}>
            <div class="modal-head"><h3>{editingProductId() ? t('edit_product') : t('new_product')}</h3><button class="close-btn" onClick={() => setShowProductModal(false)}><X size={22} /></button></div>
            <div class="wc-editor-tabs">
              <button classList={{ active: modalTab() === 'general' }} onClick={() => setModalTab('general')}>{t('tab_general')}</button>
              <button classList={{ active: modalTab() === 'variants' }} onClick={() => setModalTab('variants')}>{t('tab_variants')} ({currentVariants().length})</button>
              <button classList={{ active: modalTab() === 'images' }} onClick={() => setModalTab('images')}>{t('tab_images')} ({currentImages().length})</button>
            </div>
            <Show when={modalTab() === 'general'}>
              <form onSubmit={handleSaveGeneral} style={{ display: 'flex', 'flex-direction': 'column', gap: '1.25rem' }}>
                <div style={{ display: 'grid', 'grid-template-columns': '2fr 1fr', gap: '1rem' }}>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>{t('product_name')} *</label><input type="text" required class="glass-input" value={prodName()} onInput={(e) => setProdName(e.currentTarget.value)} /></div>
                  <div>
                    <label style={{ "font-size": "0.82rem", "font-weight": 700 }}>Producttype</label>
                    <select class="glass-input" value={prodType()} onChange={(e) => setProdType(e.currentTarget.value as any)}>
                      <option value="variable">Variabel Product (Maten/Kleuren)</option>
                      <option value="simple">Simpel Product (Vast)</option>
                    </select>
                  </div>
                </div>
                <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>{t('col_slug')}</label><input type="text" class="glass-input" value={prodSlug()} onInput={(e) => setProdSlug(e.currentTarget.value)} /></div>
                <div style={{ display: 'grid', 'grid-template-columns': '1fr 1fr', gap: '1rem' }}>
                  <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>{t('base_price')} (€) *</label><input type="number" step="0.01" required class="glass-input" value={prodPrice()} onInput={(e) => setProdPrice(Number(e.currentTarget.value))} /></div>
                  <div style={{ display: 'flex', 'align-items': 'center', gap: '0.5rem', 'margin-top': '1.5rem' }}>
                    <input type="checkbox" id="featured_cb" checked={prodFeatured()} onChange={(e) => setProdFeatured(e.currentTarget.checked)} style={{ width: '18px', height: '18px' }} />
                    <label for="featured_cb" style={{ "font-weight": 600 }}>{t('featured_on_home')}</label>
                  </div>
                </div>
                <div><label style={{ "font-size": "0.82rem", "font-weight": 700 }}>{t('product_desc')}</label><textarea rows="4" class="glass-input" value={prodDesc()} onInput={(e) => setProdDesc(e.currentTarget.value)} /></div>
                <div style={{ display: 'flex', 'justify-content': 'flex-end', gap: '0.75rem' }}>
                  <button type="button" class="glass-button secondary" onClick={() => setShowProductModal(false)}>{t('cancel')}</button>
                  <button type="submit" class="glass-button">{t('save_changes')}</button>
                </div>
              </form>
            </Show>

            <Show when={modalTab() === 'variants'}>
              <div style={{ display: 'flex', 'flex-direction': 'column', gap: '1.25rem' }}>
                <div class="variants-table-wrap">
                  <table>
                    <thead><tr><th>Foto</th><th>Maat</th><th>Kleur</th><th>Voorraad</th><th>POD Provider</th><th>Actie</th></tr></thead>
                    <tbody>
                      <For each={currentVariants()} fallback={<tr><td colspan="6" style={{ "text-align": "center", color: "var(--text-muted)" }}>Geen varianten.</td></tr>}>
                        {(v) => (
                          <tr>
                            <td>{v.image_url ? <img src={v.image_url} alt="" style={{ width: '36px', height: '36px', 'border-radius': '6px', 'object-fit': 'cover' }} /> : '-'}</td>
                            <td><strong>{v.size || '-'}</strong></td>
                            <td>{v.color || '-'}</td>
                            <td><span class="glass-badge">{v.stock_quantity} stuks</span></td>
                            <td><code>{v.pod_provider || 'Eigen voorraad'}</code></td>
                            <td><button style={{ color: 'var(--danger)', background: 'none', border: 'none', cursor: 'pointer' }} onClick={() => handleDeleteVariant(v.id)}><Trash2 size={16} /></button></td>
                          </tr>
                        )}
                      </For>
                    </tbody>
                  </table>
                </div>

                <div class="glass-panel" style={{ padding: '1.25rem', display: 'flex', 'flex-direction': 'column', gap: '0.85rem' }}>
                  <h4 style={{ "font-size": "0.95rem" }}>Nieuwe Variant Toevoegen</h4>
                  <div style={{ display: 'grid', 'grid-template-columns': 'repeat(auto-fit, minmax(120px, 1fr))', gap: '0.75rem' }}>
                    <div><label style={{ "font-size": "0.78rem", "font-weight": 600 }}>Maat</label><input type="text" class="glass-input" value={newVarSize()} onInput={(e) => setNewVarSize(e.currentTarget.value)} /></div>
                    <div><label style={{ "font-size": "0.78rem", "font-weight": 600 }}>Kleur</label><input type="text" class="glass-input" value={newVarColor()} onInput={(e) => setNewVarColor(e.currentTarget.value)} /></div>
                    <div><label style={{ "font-size": "0.78rem", "font-weight": 600 }}>Voorraad</label><input type="number" class="glass-input" value={newVarStock()} onInput={(e) => setNewVarStock(Number(e.currentTarget.value))} /></div>
                    <div><label style={{ "font-size": "0.78rem", "font-weight": 600 }}>Provider</label><select class="glass-input" value={newVarPODProvider()} onChange={(e) => setNewVarPODProvider(e.currentTarget.value)}><option value="none">Eigen voorraad</option><option value="gelato">Gelato POD</option></select></div>
                    <div><label style={{ "font-size": "0.78rem", "font-weight": 600 }}>Gelato ID</label><input type="text" class="glass-input" value={newVarPODID()} onInput={(e) => setNewVarPODID(e.currentTarget.value)} /></div>
                  </div>
                  <div style={{ display: 'grid', 'grid-template-columns': '1fr 1fr', gap: '0.75rem' }}>
                    <div><label style={{ "font-size": "0.78rem", "font-weight": 600 }}>Foto URL</label><input type="text" class="glass-input" value={newVarImage()} onInput={(e) => setNewVarImage(e.currentTarget.value)} /></div>
                    <div><label style={{ "font-size": "0.78rem", "font-weight": 600 }}>Drukbestand URL</label><input type="text" class="glass-input" value={newVarPrintFile()} onInput={(e) => setNewVarPrintFile(e.currentTarget.value)} /></div>
                  </div>
                  <button type="button" class="glass-button" style={{ 'align-self': 'flex-start' }} onClick={handleAddVariant}><Plus size={16} /> Variant Opslaan</button>
                </div>
              </div>
            </Show>

            <Show when={modalTab() === 'images'}>
              <div style={{ display: 'flex', 'flex-direction': 'column', gap: '1.25rem' }}>
                <div class="glass-panel" style={{ padding: '1.25rem', display: 'flex', 'flex-direction': 'column', gap: '0.85rem' }}>
                  <div><label style={{ "font-size": "0.78rem", "font-weight": 600 }}>Afbeelding URL *</label><input type="text" class="glass-input" value={newImageUrl()} onInput={(e) => setNewImageUrl(e.currentTarget.value)} /></div>
                  <button type="button" class="glass-button" style={{ 'align-self': 'flex-start' }} onClick={handleAddImage}><Plus size={16} /> Toevoegen</button>
                </div>
                <div class="gallery-grid">
                  <For each={currentImages()}>
                    {(img) => (
                      <div class="gallery-card">
                        <img src={img.url} alt="" />
                        <button class="del-img-btn" onClick={() => handleDeleteImage(img.id)}><Trash2 size={13} /></button>
                      </div>
                    )}
                  </For>
                </div>
              </div>
            </Show>
          </div>
        </div>
      </Show>

      {/* Coupon Modal */}
      <Show when={showCouponModal()}>
        <div class="wc-modal-backdrop" onClick={() => setShowCouponModal(false)}>
          <div class="wc-modal" style={{ "max-width": "500px" }} onClick={(e) => e.stopPropagation()}>
            <div class="modal-head"><h3>Nieuwe Kortingscode</h3><button class="close-btn" onClick={() => setShowCouponModal(false)}><X size={20} /></button></div>
            <form onSubmit={handleCreateCoupon} style={{ display: 'flex', 'flex-direction': 'column', gap: '1rem' }}>
              <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Code *</label><input type="text" required class="glass-input" value={couponCode()} onInput={(e) => setCouponCode(e.currentTarget.value)} /></div>
              <div style={{ display: 'grid', 'grid-template-columns': '1fr 1fr', gap: '0.75rem' }}>
                <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Type</label><select class="glass-input" value={couponType()} onChange={(e) => setCouponType(e.currentTarget.value as any)}><option value="percent">Percentage (%)</option><option value="fixed">Vast Bedrag (€)</option></select></div>
                <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Waarde *</label><input type="number" step="0.5" required class="glass-input" value={couponValue()} onInput={(e) => setCouponValue(Number(e.currentTarget.value))} /></div>
              </div>
              <button type="submit" class="glass-button" style={{ 'align-self': 'flex-end', 'margin-top': '0.5rem' }}>Aanmaken</button>
            </form>
          </div>
        </div>
      </Show>

      {/* Widget Editor Modal */}
      <Show when={showWidgetModal()}>
        <div class="wc-modal-backdrop" onClick={() => setShowWidgetModal(false)}>
          <div class="wc-modal" style={{ "max-width": "580px" }} onClick={(e) => e.stopPropagation()}>
            <div class="modal-head"><h3>{editingWidgetId() ? t('edit_widget') : t('add_widget')}</h3><button class="close-btn" onClick={() => setShowWidgetModal(false)}><X size={20} /></button></div>
            <form onSubmit={handleSaveWidget} style={{ display: 'flex', 'flex-direction': 'column', gap: '1rem' }}>
              <div style={{ display: 'grid', 'grid-template-columns': '2fr 1fr', gap: '0.75rem' }}>
                <div>
                  <label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Type Widget *</label>
                  <select class="glass-input" value={widgetType()} onChange={(e) => setWidgetType(e.currentTarget.value)}>
                    <option value="hero">Hero Banner</option>
                    <option value="trust_badges">Trust Badges</option>
                    <option value="featured_products">Populaire Items</option>
                    <option value="newsletter">Nieuwsbrief Widget</option>
                    <option value="banner_promo">Promotie Banner</option>
                  </select>
                </div>
                <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Volgorde</label><input type="number" min="1" class="glass-input" value={widgetOrder()} onInput={(e) => setWidgetOrder(Number(e.currentTarget.value))} /></div>
              </div>
              <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Titel</label><input type="text" class="glass-input" value={widgetTitle()} onInput={(e) => setWidgetTitle(e.currentTarget.value)} /></div>
              <div><label style={{ "font-size": "0.8rem", "font-weight": 700 }}>Subtitel</label><textarea rows="2" class="glass-input" value={widgetSubtitle()} onInput={(e) => setWidgetSubtitle(e.currentTarget.value)} /></div>
              <Show when={widgetType() === 'hero' || widgetType() === 'banner_promo'}>
                <div class="glass-panel" style={{ padding: '1rem', display: 'flex', 'flex-direction': 'column', gap: '0.75rem' }}>
                  <div><label style={{ "font-size": "0.75rem", "font-weight": 600 }}>Foto URL</label><input type="text" class="glass-input" value={widgetBgImage()} onInput={(e) => setWidgetBgImage(e.currentTarget.value)} /></div>
                  <div style={{ display: 'grid', 'grid-template-columns': '1fr 1fr', gap: '0.75rem' }}>
                    <div><label style={{ "font-size": "0.75rem", "font-weight": 600 }}>Knoptekst</label><input type="text" class="glass-input" value={widgetBtnText()} onInput={(e) => setWidgetBtnText(e.currentTarget.value)} /></div>
                    <div><label style={{ "font-size": "0.75rem", "font-weight": 600 }}>Link</label><input type="text" class="glass-input" value={widgetBtnLink()} onInput={(e) => setWidgetBtnLink(e.currentTarget.value)} /></div>
                  </div>
                </div>
              </Show>
              <div style={{ display: 'flex', 'justify-content': 'flex-end', gap: '0.75rem' }}>
                <button type="button" class="glass-button secondary" onClick={() => setShowWidgetModal(false)}>Annuleren</button>
                <button type="submit" class="glass-button">Opslaan</button>
              </div>
            </form>
          </div>
        </div>
      </Show>
    </div>
  );
};