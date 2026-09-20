import { Component, createResource, For, Show } from 'solid-js';
import { A, useNavigate } from '@solidjs/router';
import { Sparkles, Truck, RotateCcw, ShieldCheck, ArrowRight, Heart } from 'lucide-solid';
import { api } from '../lib/api';
import { useWishlist } from '../store/wishlist';
import { useI18n } from '../store/i18n';
import { useCurrency } from '../store/currency';
import './Home.scss';

export const Home: Component = () => {
  const navigate = useNavigate();
  const { toggleWishlist, isWishlisted } = useWishlist();
  const { t } = useI18n();
  const { formatPrice } = useCurrency();

  const [products] = createResource(async () => {
    const res = await api.get<{ products: any[] }>('/products?limit=8&featured=true');
    return res.products;
  });

  return (
    <div class="home-page container">
      {/* 100vh Schermvullende Hero */}
      <section class="fullscreen-hero">
        <div class="hero-content">
          <div class="hero-tag">
            <Sparkles size={16} />
            <span>{t('hero_badge')}</span>
          </div>
          <h1 class="hero-title">{t('hero_title')}</h1>
          <p class="hero-subtitle">{t('hero_subtitle')}</p>
          <div class="hero-cta">
            <A href="/shop" class="glass-button btn-large">
              {t('hero_cta_collection')} <ArrowRight size={18} />
            </A>
            <A href="/shop?category=oversized" class="glass-button btn-large btn-glass-white">
              {t('hero_cta_oversized')}
            </A>
          </div>
        </div>
      </section>

      {/* Trust Badges Widget */}
      <section class="trust-widget">
        <div class="trust-grid">
          <div class="trust-card glass-panel">
            <div class="trust-icon"><Truck size={28} /></div>
            <div>
              <h4>{t('trust_shipping_title')}</h4>
              <p>{t('trust_shipping_desc')}</p>
            </div>
          </div>
          <div class="trust-card glass-panel">
            <div class="trust-icon"><RotateCcw size={28} /></div>
            <div>
              <h4>{t('trust_returns_title')}</h4>
              <p>{t('trust_returns_desc')}</p>
            </div>
          </div>
          <div class="trust-card glass-panel">
            <div class="trust-icon"><ShieldCheck size={28} /></div>
            <div>
              <h4>{t('trust_eco_title')}</h4>
              <p>{t('trust_eco_desc')}</p>
            </div>
          </div>
        </div>
      </section>

      {/* Uitgelichte Producten */}
      <section>
        <div class="section-header">
          <div>
            <h2 class="section-title">{t('popular_items')}</h2>
            <p style={{ color: "var(--text-muted)" }}>{t('popular_subtitle')}</p>
          </div>
          <A href="/shop" class="see-all">
            {t('see_all')} <ArrowRight size={16} />
          </A>
        </div>

        <div class="products-grid">
          <Show when={!products.loading} fallback={<p>{t('loading')}</p>}>
            <For each={products()}>
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
                        {t('view')}
                      </span>
                    </div>
                  </div>
                </div>
              )}
            </For>
          </Show>
        </div>
      </section>

      {/* Nieuwsbrief Widget */}
      <section class="newsletter-widget glass-card">
        <h3>{t('newsletter_title')}</h3>
        <p>{t('newsletter_desc')}</p>
        <form class="nl-form" onSubmit={(e) => { e.preventDefault(); alert('Inschrijving geslaagd!'); }}>
          <input type="email" placeholder={t('email_placeholder')} class="glass-input" required />
          <button type="submit" class="glass-button">{t('subscribe')}</button>
        </form>
      </section>
    </div>
  );
};