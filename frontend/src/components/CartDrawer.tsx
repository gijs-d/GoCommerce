import { Component, Show, For } from 'solid-js';
import { A } from '@solidjs/router';
import { X, Trash2, Plus, Minus, ShoppingBag } from 'lucide-solid';
import { useCart } from '../store/cart';
import { useI18n } from '../store/i18n';
import { useCurrency } from '../store/currency';
import './CartDrawer.scss';

export const CartDrawer: Component = () => {
  const { isCartOpen, closeCart, items, formattedTotal, updateQuantity, removeItem, totalCount } = useCart();
  const { t } = useI18n();
  const { formatPrice } = useCurrency();

  return (
    <Show when={isCartOpen()}>
      <div class="cart-overlay" onClick={closeCart}>
        <div class="cart-slide" onClick={(e) => e.stopPropagation()}>
          <div class="cart-header">
            <h3>{t('cart')} ({totalCount()})</h3>
            <button class="close-btn" onClick={closeCart}>
              <X size={22} />
            </button>
          </div>

          <div class="cart-items">
            <Show
              when={items().length > 0}
              fallback={
                <div class="empty-cart">
                  <ShoppingBag size={48} />
                  <p>{t('cart_empty')}</p>
                  <button class="glass-button secondary" onClick={closeCart}>
                    {t('continue_shopping')}
                  </button>
                </div>
              }
            >
              <For each={items()}>
                {(item) => (
                  <div class="cart-card">
                    <img src={item.image || '/static/tshirt-placeholder.jpg'} alt={item.name} class="item-thumb" />
                    <div class="item-details">
                      <div>
                        <h4 class="item-name">{item.name}</h4>
                        <Show when={item.variantTitle}>
                          <span class="item-variant">{item.variantTitle}</span>
                        </Show>
                      </div>

                      <div class="item-bottom">
                        <span class="item-price">{formatPrice(item.price * item.quantity)}</span>
                        <div class="qty-selector">
                          <button onClick={() => updateQuantity(item.id, item.quantity - 1)}>
                            <Minus size={14} />
                          </button>
                          <span>{item.quantity}</span>
                          <button onClick={() => updateQuantity(item.id, item.quantity + 1)}>
                            <Plus size={14} />
                          </button>
                        </div>
                      </div>
                    </div>

                    <button class="remove-btn" onClick={() => removeItem(item.id)} title="Verwijderen">
                      <Trash2 size={16} />
                    </button>
                  </div>
                )}
              </For>
            </Show>
          </div>

          <Show when={items().length > 0}>
            <div class="cart-footer">
              <div class="subtotal-row">
                <span>{t('subtotal')}</span>
                <span>{formattedTotal()}</span>
              </div>
              <p style={{ "font-size": "0.8rem", color: "var(--text-muted)" }}>
                {t('free_shipping_notice')}
              </p>
              <A href="/checkout" class="glass-button checkout-btn" onClick={closeCart}>
                {t('checkout')}
              </A>
            </div>
          </Show>
        </div>
      </div>
    </Show>
  );
};