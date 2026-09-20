import { Component, createResource, createSignal, For, Show } from 'solid-js';
import { useParams } from '@solidjs/router';
import { Check, FileText } from 'lucide-solid';
import { api } from '../lib/api';
import { useI18n } from '../store/i18n';
import { useCurrency } from '../store/currency';
import './OrderTracking.scss';

export const OrderTracking: Component = () => {
  const params = useParams();
  const { t, lang } = useI18n();
  const { formatPrice } = useCurrency();
  const [orderQuery, setOrderQuery] = createSignal(params.orderNumber || '');

  const [order, { refetch }] = createResource(
    () => orderQuery(),
    async (nr) => {
      if (!nr) return null;
      return await api.get<any>(`/orders/${nr}`);
    }
  );

  const handleSimulatePayment = async () => {
    const o = order();
    if (!o) return;
    await api.post(`/orders/${o.order_number}/pay-mock`);
    refetch();
  };

  return (
    <div class="track-page container">
      <Show when={!params.orderNumber}>
        <div class="search-card glass-panel">
          <input
            type="text"
            placeholder={t('track_input_placeholder')}
            class="glass-input"
            value={orderQuery()}
            onInput={(e) => setOrderQuery(e.currentTarget.value)}
          />
        </div>
      </Show>

      <Show when={order()} fallback={<p style={{ "text-align": "center", "margin-top": "3rem" }}>{t('track_not_found')}</p>}>
        {(o) => (
          <div class="tracking-card glass-panel">
            <div class="tracking-header">
              <div>
                <h2>{t('track_order_title')} {o().order_number}</h2>
                <span style={{ "font-size": "0.85rem", color: "var(--text-muted)" }}>
                  {t('placed_on')} {new Date(o().created_at).toLocaleDateString(lang() === 'nl' ? 'nl-BE' : 'en-US')}
                </span>
              </div>
              <div style={{ display: 'flex', 'align-items': 'center', gap: '0.75rem' }}>
                <a
                  href={`/api/orders/${o().order_number}/invoice`}
                  target="_blank"
                  class="glass-button secondary"
                  style={{ padding: '0.35rem 0.75rem', 'font-size': '0.8rem' }}
                >
                  <FileText size={15} /> Factuur (PDF)
                </a>
                <span class="glass-badge" style={{ 'text-transform': 'uppercase' }}>
                  {o().status}
                </span>
              </div>
            </div>

            {/* Testknop om betaling te voltooien indien unpaid */}
            <Show when={o().payment_status === 'unpaid'}>
              <div class="glass-panel" style={{ padding: '1rem', display: 'flex', 'justify-content': 'space-between', 'align-items': 'center', background: 'rgba(234, 179, 8, 0.1)', 'border-color': 'rgba(234, 179, 8, 0.3)' }}>
                <span>{t('track_payment_pending')}</span>
                <button class="glass-button" onClick={handleSimulatePayment}>
                  {t('track_simulate_pay')}
                </button>
              </div>
            </Show>

            {/* Live Tijdlijn */}
            <div class="timeline">
              <For each={o().timeline}>
                {(step: any) => (
                  <div class="timeline-step">
                    <div class="step-icon active">
                      <Check size={18} />
                    </div>
                    <div class="step-content">
                      <h4>{step.message}</h4>
                      <time>{new Date(step.created_at).toLocaleTimeString(lang() === 'nl' ? 'nl-BE' : 'en-US', { hour: '2-digit', minute: '2-digit' })} - {new Date(step.created_at).toLocaleDateString(lang() === 'nl' ? 'nl-BE' : 'en-US')}</time>
                    </div>
                  </div>
                )}
              </For>
            </div>

            {/* Bestelde artikelen */}
            <div class="order-items-box">
              <h4 style={{ "margin-bottom": "0.75rem" }}>{t('track_ordered_items')}</h4>
              <For each={o().items}>
                {(item: any) => (
                  <div style={{ display: 'flex', 'justify-content': 'space-between', 'padding': '0.4rem 0', 'font-size': '0.9rem' }}>
                    <span>{item.product_name} ({item.variant_title}) × {item.quantity}</span>
                    <strong>{formatPrice(item.total_price)}</strong>
                  </div>
                )}
              </For>
            </div>
          </div>
        )}
      </Show>
    </div>
  );
};