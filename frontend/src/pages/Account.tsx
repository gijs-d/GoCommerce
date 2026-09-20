import { Component, createSignal, createResource, For, Show, createEffect } from 'solid-js';
import { useNavigate, useLocation } from '@solidjs/router';
import { Package, Heart, MapPin, LogOut } from 'lucide-solid';
import { useAuth } from '../store/auth';
import { useI18n } from '../store/i18n';
import { useCurrency } from '../store/currency';
import { api } from '../lib/api';
import './Account.scss';

export const Account: Component = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout, isLoading, isAuthenticated } = useAuth();
  const { t, lang } = useI18n();
  const { formatPrice } = useCurrency();

  // Bepaal de actieve tab op basis van de URL (/account/wishlist -> 'wishlist')
  const getTabFromPath = () => {
    if (location.pathname.includes('wishlist')) return 'wishlist';
    if (location.pathname.includes('addresses')) return 'addresses';
    return 'orders';
  };

  const [activeTab, setActiveTab] = createSignal<'orders' | 'wishlist' | 'addresses'>(getTabFromPath());

  createEffect(() => {
    if (!isLoading() && !isAuthenticated()) {
      navigate('/login');
    }
  });

  // Schakel tabblad mee zodra de URL verandert (bijv. bij klikken op hartje)
  createEffect(() => {
    setActiveTab(getTabFromPath());
  });

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const [orders] = createResource(() => api.get<any[]>('/user/orders'));
  const [wishlist] = createResource(() => api.get<any[]>('/user/wishlist'));
  const [addresses] = createResource(() => api.get<any[]>('/user/addresses'));

  return (
    <div class="account-page container">
      <div class="account-layout">
        {/* Zijbalk Tabs */}
        <aside class="nav-tabs glass-panel">
          <div style={{ padding: '0.5rem 1rem 1rem', 'border-bottom': '1px solid var(--glass-border)', 'margin-bottom': '0.5rem' }}>
            <h3 style={{ "font-size": "1.1rem" }}>{user()?.full_name}</h3>
            <span style={{ "font-size": "0.8rem", color: "var(--text-muted)" }}>{user()?.email}</span>
          </div>

          <button
            class="tab-btn"
            classList={{ active: activeTab() === 'orders' }}
            onClick={() => {
              setActiveTab('orders');
              navigate('/account');
            }}
          >
            <Package size={18} /> {t('orders')}
          </button>
          <button
            class="tab-btn"
            classList={{ active: activeTab() === 'wishlist' }}
            onClick={() => {
              setActiveTab('wishlist');
              navigate('/account/wishlist');
            }}
          >
            <Heart size={18} /> {t('wishlist')}
          </button>
          <button
            class="tab-btn"
            classList={{ active: activeTab() === 'addresses' }}
            onClick={() => {
              setActiveTab('addresses');
              navigate('/account');
            }}
          >
            <MapPin size={18} /> {t('saved_addresses')}
          </button>
          <button class="tab-btn" style={{ color: 'var(--danger)' }} onClick={handleLogout}>
            <LogOut size={18} /> {t('logout')}
          </button>
        </aside>

        {/* Inhoud per Tab */}
        <main class="tab-content">
          {/* Bestellingen Tab */}
          <Show when={activeTab() === 'orders'}>
            <div class="glass-panel" style={{ padding: '1.5rem' }}>
              <h3 style={{ "margin-bottom": "1rem" }}>{t('my_orders')}</h3>
              <Show when={!orders.loading} fallback={<p>{t('loading')}</p>}>
                <For each={orders()} fallback={<p style={{ color: "var(--text-muted)" }}>Nog geen bestellingen gevonden.</p>}>
                  {(o) => (
                    <div style={{ display: 'flex', 'justify-content': 'space-between', 'align-items': 'center', padding: '1rem', 'border-bottom': '1px solid var(--glass-border)' }}>
                      <div>
                        <strong>{o.order_number}</strong>
                        <div style={{ "font-size": "0.8rem", color: "var(--text-muted)" }}>
                          {new Date(o.created_at).toLocaleDateString(lang() === 'nl' ? 'nl-BE' : 'en-US')} • {formatPrice(o.total_amount)}
                        </div>
                      </div>
                      <a href={`/track/${o.order_number}`} class="glass-button secondary" style={{ padding: '0.4rem 0.8rem', 'font-size': '0.82rem' }}>
                        {t('track')}
                      </a>
                    </div>
                  )}
                </For>
              </Show>
            </div>
          </Show>

          {/* Favorieten Tab */}
          <Show when={activeTab() === 'wishlist'}>
            <div class="glass-panel" style={{ padding: '1.5rem' }}>
              <h3 style={{ "margin-bottom": "1rem" }}>{t('my_wishlist')}</h3>
              <Show when={!wishlist.loading} fallback={<p>{t('loading')}</p>}>
                <For each={wishlist()} fallback={<p style={{ color: "var(--text-muted)" }}>Nog geen favorieten toegevoegd.</p>}>
                  {() => (
                    <div style={{ display: 'grid', 'grid-template-columns': 'repeat(auto-fill, minmax(200px, 1fr))', gap: '1rem' }}>
                      <For each={wishlist()}>
                        {(item) => (
                          <div class="glass-card" style={{ padding: '0.8rem' }}>
                            <img src={item.product.primary_image || '/static/tshirt-placeholder.jpg'} alt="" style={{ width: '100%', 'aspect-ratio': '1', 'object-fit': 'cover', 'border-radius': '10px' }} />
                            <h4 style={{ "font-size": "0.95rem", "margin-top": "0.5rem" }}>{item.product.name}</h4>
                            <div style={{ "font-weight": 700, "margin-top": "0.2rem" }}>{formatPrice(item.product.base_price)}</div>
                            <a href={`/product/${item.product.slug}`} class="glass-button secondary" style={{ width: '100%', 'margin-top': '0.5rem', 'font-size': '0.8rem', padding: '0.4rem' }}>
                              {t('view')}
                            </a>
                          </div>
                        )}
                      </For>
                    </div>
                  )}
                </For>
              </Show>
            </div>
          </Show>

          {/* Adressen Tab */}
          <Show when={activeTab() === 'addresses'}>
            <div class="glass-panel" style={{ padding: '1.5rem' }}>
              <h3 style={{ "margin-bottom": "1rem" }}>{t('saved_addresses')}</h3>
              <For each={addresses()} fallback={<p style={{ color: "var(--text-muted)" }}>Nog geen adressen opgeslagen.</p>}>
                {(addr) => (
                  <div style={{ padding: '1rem', 'border-radius': '12px', background: 'var(--input-bg)', 'margin-bottom': '1rem' }}>
                    <strong>{addr.label} - {addr.full_name}</strong>
                    <p>{addr.street} {addr.house_number} {addr.bus || ''}, {addr.postal_code} {addr.city}</p>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </main>
      </div>
    </div>
  );
};