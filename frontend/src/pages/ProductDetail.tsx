import { Component, createResource, createSignal, For, Show } from 'solid-js';
import { useParams, A } from '@solidjs/router';
import { ShoppingBag, Heart, Check, ArrowLeft } from 'lucide-solid';
import { api } from '../lib/api';
import { useCart } from '../store/cart';
import { useWishlist } from '../store/wishlist';
import { useI18n } from '../store/i18n';
import { useCurrency } from '../store/currency';
import './ProductDetail.scss';

export const ProductDetail: Component = () => {
  const params = useParams();
  const { addItem } = useCart();
  const { toggleWishlist, isWishlisted } = useWishlist();
  const { t } = useI18n();
  const { formatPrice } = useCurrency();

  const [product] = createResource(() => params.slug, async (slug) => {
    return await api.get<any>(`/products/${slug}`);
  });

  const [selectedVariantId, setSelectedVariantId] = createSignal<string | undefined>(undefined);
  const [activeImage, setActiveImage] = createSignal<string | null>(null);

  const selectedVariant = () => {
    const p = product();
    if (!p || !p.variants || p.variants.length === 0) return null;
    return p.variants.find((v: any) => v.id === selectedVariantId()) || p.variants[0];
  };

  const handleAddToCart = () => {
    const p = product();
    if (!p) return;
    const v = selectedVariant();

    const price = v && v.price_override ? v.price_override : p.base_price;
    const image = (v && v.image_url) || p.primary_image;

    addItem({
      productId: p.id,
      variantId: v?.id,
      name: p.name,
      variantTitle: v ? `${v.size || ''} ${v.color || ''}`.trim() : undefined,
      price: price,
      quantity: 1,
      image: image,
      size: v?.size,
      color: v?.color,
    });
  };

  return (
    <div class="pdp-page container">
      <Show when={!product.loading && product()} fallback={<p>{t('loading')}</p>}>
        {(p) => {
          const mainImg = () => activeImage() || p().primary_image || '/static/tshirt-placeholder.jpg';

          return (
            <div>
              {/* Bovenbalk met Vorige-knop en categorie */}
              <div class="pdp-top-nav">
                <button class="back-link" onClick={() => window.history.back()} title={t('back_to_overview')}>
                  <ArrowLeft size={18} /> {t('back_to_overview')}
                </button>
                <span class="breadcrumb-text">
                  <A href="/shop">{t('collection')}</A> / <span>{p().category_name || 'T-Shirts'}</span>
                </span>
              </div>

              <div class="pdp-layout">
                {/* Afbeeldingen Galerij */}
                <div class="gallery-section">
                  <div class="main-image-wrap glass-panel">
                    <img src={mainImg()} alt={p().name} />
                  </div>
                  <Show when={p().images && p().images.length > 1}>
                    <div class="thumbs-row">
                      <For each={p().images}>
                        {(img: any) => (
                          <button
                            class="thumb-btn glass-panel"
                            classList={{ active: mainImg() === img.url }}
                            onClick={() => setActiveImage(img.url)}
                          >
                            <img src={img.url} alt={img.alt_text || p().name} />
                          </button>
                        )}
                      </For>
                    </div>
                  </Show>
                </div>

                {/* Product Info & Aankoop */}
                <div class="details-section">
                  <span class="pdp-category">{p().category_name || 'Apparel'}</span>
                  <h1 class="pdp-title">{p().name}</h1>
                  <div class="pdp-price">
                    {formatPrice(selectedVariant()?.price_override ?? p().base_price)}
                  </div>

                  {/* Varianten / Maat selectie */}
                  <Show when={p().variants && p().variants.length > 0}>
                    <div class="variant-picker">
                      <span class="picker-label">{t('select_size')}</span>
                      <div class="pills-list">
                        <For each={p().variants}>
                          {(v: any) => (
                            <button
                              class="size-pill"
                              classList={{ selected: (selectedVariantId() || p().variants[0]?.id) === v.id }}
                              onClick={() => {
                                setSelectedVariantId(v.id);
                                if (v.image_url) setActiveImage(v.image_url);
                              }}
                            >
                              {v.size || v.title}
                            </button>
                          )}
                        </For>
                      </div>
                    </div>
                  </Show>

                  {/* Beschrijving */}
                  <div class="glass-panel" style={{ padding: '1.25rem', 'margin-top': '0.5rem' }}>
                    <p style={{ "line-height": 1.7, color: "var(--text-muted)" }}>
                      {p().description}
                    </p>
                  </div>

                  {/* CTA Knoppen */}
                  <div class="cta-row">
                    <button class="glass-button add-to-cart" onClick={handleAddToCart}>
                      <ShoppingBag size={20} /> {t('add_to_cart')}
                    </button>
                    <button
                      class="glass-button secondary wish-btn"
                      classList={{ active: isWishlisted(p().id) }}
                      onClick={() => toggleWishlist(p().id)}
                      title={isWishlisted(p().id) ? "Verwijderen uit favorieten" : "Opslaan in favorieten"}
                      aria-label="Favoriet"
                    >
                      <Heart size={20} fill={isWishlisted(p().id) ? "currentColor" : "none"} />
                    </button>
                  </div>

                  {/* Voordelen bullet points */}
                  <div style={{ display: 'flex', 'flex-direction': 'column', gap: '0.6rem', 'margin-top': '1rem' }}>
                    <div style={{ display: 'flex', 'align-items': 'center', gap: '0.6rem', 'font-size': '0.85rem' }}>
                      <Check size={16} color="var(--success)" /> {t('cotton_benefit')}
                    </div>
                    <div style={{ display: 'flex', 'align-items': 'center', gap: '0.6rem', 'font-size': '0.85rem' }}>
                      <Check size={16} color="var(--success)" /> {t('delivery_benefit')}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          );
        }}
      </Show>
    </div>
  );
};