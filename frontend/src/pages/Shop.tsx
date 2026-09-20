import { Component, createSignal, createResource, For, Show } from 'solid-js';
import { useSearchParams, useNavigate } from '@solidjs/router';
import { Heart } from 'lucide-solid';
import { api } from '../lib/api';
import { useWishlist } from '../store/wishlist';
import { useI18n } from '../store/i18n';
import { useCurrency } from '../store/currency';
import './Shop.scss';

export const Shop: Component = () => {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { toggleWishlist, isWishlisted } = useWishlist();
  const { t } = useI18n();
  const { formatPrice } = useCurrency();

  const [selectedCategory, setSelectedCategory] = createSignal(params.category || '');
  const [selectedSize, setSelectedSize] = createSignal('');
  const [searchQuery, setSearchQuery] = createSignal('');
  const [sortBy, setSortBy] = createSignal('newest');

  const [categories] = createResource(() => api.get<any[]>('/categories'));

  const [productsData] = createResource(
    () => ({
      cat: selectedCategory(),
      size: selectedSize(),
      q: searchQuery(),
      sort: sortBy(),
    }),
    async (p) => {
      const qParams = new URLSearchParams();
      if (p.cat) qParams.set('category', p.cat);
      if (p.size) qParams.set('size', p.size);
      if (p.q) qParams.set('search', p.q);
      if (p.sort) qParams.set('sort', p.sort);
      return await api.get<{ products: any[]; total: number }>(`/products?${qParams.toString()}`);
    }
  );

  const sizes = ['XS', 'S', 'M', 'L', 'XL', 'XXL'];

  return (
    <div class="shop-page container">
      <div class="shop-header">
        <h1>{t('shop_title')}</h1>
        <p>{t('shop_subtitle')}</p>
      </div>

      <div class="shop-layout">
        {/* Filter zijbalk */}
        <aside class="shop-filters glass-panel">
          <div class="filter-group">
            <h4>{t('search')}</h4>
            <div style={{ position: 'relative' }}>
              <input
                type="text"
                placeholder={t('search_placeholder')}
                class="glass-input"
                value={searchQuery()}
                onInput={(e) => setSearchQuery(e.currentTarget.value)}
              />
            </div>
          </div>

          <div class="filter-group">
            <h4>{t('category')}</h4>
            <div class="filter-pills">
              <button
                class="pill"
                classList={{ active: selectedCategory() === '' }}
                onClick={() => setSelectedCategory('')}
              >
                {t('all')}
              </button>
              <For each={categories()}>
                {(cat) => (
                  <button
                    class="pill"
                    classList={{ active: selectedCategory() === cat.slug }}
                    onClick={() => setSelectedCategory(cat.slug)}
                  >
                    {cat.name}
                  </button>
                )}
              </For>
            </div>
          </div>

          <div class="filter-group">
            <h4>{t('size')}</h4>
            <div class="filter-pills">
              <button
                class="pill"
                classList={{ active: selectedSize() === '' }}
                onClick={() => setSelectedSize('')}
              >
                {t('all_sizes')}
              </button>
              <For each={sizes}>
                {(s) => (
                  <button
                    class="pill"
                    classList={{ active: selectedSize() === s }}
                    onClick={() => setSelectedSize(s)}
                  >
                    {s}
                  </button>
                )}
              </For>
            </div>
          </div>
        </aside>

        {/* Producten overzicht */}
        <main class="catalog-area">
          <div class="toolbar glass-panel" style={{ padding: '0.8rem 1.25rem' }}>
            <span style={{ "font-weight": 600, color: "var(--text-muted)" }}>
              {productsData()?.total ?? 0} {t('products_found')}
            </span>

            <div style={{ display: 'flex', 'align-items': 'center', gap: '0.5rem' }}>
              <span style={{ "font-size": "0.85rem", color: "var(--text-muted)" }}>{t('sort_label')}</span>
              <select
                class="glass-input"
                style={{ width: 'auto', padding: '0.35rem 0.75rem' }}
                value={sortBy()}
                onChange={(e) => setSortBy(e.currentTarget.value)}
              >
                <option value="newest">{t('sort_newest')}</option>
                <option value="price_asc">{t('sort_price_asc')}</option>
                <option value="price_desc">{t('sort_price_desc')}</option>
                <option value="name">{t('sort_name')}</option>
              </select>
            </div>
          </div>

          <div class="products-grid">
            <Show when={!productsData.loading} fallback={<p>{t('loading')}</p>}>
              <For each={productsData()?.products}>
                {(p) => (
                  <div
                    class="glass-card product-item-card"
                    style={{ cursor: 'pointer' }}
                    onClick={() => navigate(`/product/${p.slug}`)}
                  >
                    <div class="card-image-wrap">
                      <img src={p.primary_image || '/static/tshirt-placeholder.jpg'} alt={p.name} />
                      <button
                        class="wishlist-btn"
                        classList={{ active: isWishlisted(p.id) }}
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          toggleWishlist(p.id);
                        }}
                        title={t('wishlist')}
                      >
                        <Heart size={18} fill={isWishlisted(p.id) ? "currentColor" : "none"} />
                      </button>
                    </div>
                    <div class="card-info">
                      <span class="card-category">{p.category_name || 'Apparel'}</span>
                      <h3 class="card-title">{p.name}</h3>
                      <div class="card-bottom">
                        <span class="card-price">{formatPrice(p.base_price)}</span>
                        <span class="glass-button secondary mini">
                          {t('details')}
                        </span>
                      </div>
                    </div>
                  </div>
                )}
              </For>
            </Show>
          </div>
        </main>
      </div>
    </div>
  );
};