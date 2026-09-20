import { Component, createSignal, createResource, For, Show, createEffect } from 'solid-js';
import { MapPin, CreditCard, ShieldCheck, Ticket, Building2, Truck, Check } from 'lucide-solid';
import { useCart } from '../store/cart';
import { useAuth } from '../store/auth';
import { useI18n } from '../store/i18n';
import { useCurrency } from '../store/currency';
import { api } from '../lib/api';
import './Checkout.scss';

export const Checkout: Component = () => {
  const { items, totalPrice, clearCart } = useCart();
  const { user } = useAuth();
  const { t } = useI18n();
  const { formatPrice } = useCurrency();

  // Klant- en Adresgegevens
  const [email, setEmail] = createSignal(user()?.email || '');
  const [fullName, setFullName] = createSignal(user()?.full_name || '');
  const [street, setStreet] = createSignal('');
  const [houseNumber, setHouseNumber] = createSignal('');
  const [bus, setBus] = createSignal('');
  const [postalCode, setPostalCode] = createSignal('');
  const [city, setCity] = createSignal('');
  const [country, setCountry] = createSignal('BE');

  // B2B & Btw-verlegging
  const [isB2B, setIsB2B] = createSignal(false);
  const [vatNumber, setVatNumber] = createSignal('');
  const [viesCompany, setViesCompany] = createSignal<string | null>(null);

  // Kortingscode
  const [couponInput, setCouponInput] = createSignal('');
  const [appliedCoupon, setAppliedCoupon] = createSignal<{ code: string; discount_amount: number } | null>(null);

  // Geselecteerde verzendmethode & betaalmethode
  const [selectedShippingId, setSelectedShippingId] = createSignal<string>('default');
  const [selectedShippingCost, setSelectedShippingCost] = createSignal<number>(4.95);
  const [selectedPaymentMethod, setSelectedPaymentMethod] = createSignal<string>('mock');

  const [isSubmitting, setIsSubmitting] = createSignal(false);
  const [errorMsg, setErrorMsg] = createSignal('');

  // 1. Betaalmethoden ophalen
  const [paymentMethods] = createResource(() => api.get<any[]>('/payment-methods'));

  // 2. Verzendopties ophalen op basis van adres & mandje
  const [shippingQuotes, { refetch: refetchShipping }] = createResource(
    () => ({
      country: country(),
      postcode: postalCode(),
      subtotal: totalPrice(),
    }),
    async (p) => {
      try {
        const quotes = await api.post<any[]>('/shipping-quotes', {
          destination: { country: p.country, postcode: p.postcode, city: city() },
          items: items().map((it) => ({ variant_id: it.variantId || '', pod_variant_id: 'gelato_tee', quantity: it.quantity })),
          subtotal: p.subtotal,
        });
        if (quotes && quotes.length > 0) {
          setSelectedShippingId(quotes[0].id);
          setSelectedShippingCost(quotes[0].cost);
        }
        return quotes;
      } catch {
        return [];
      }
    }
  );

  // Btw Berekening resource
  const [taxData, { refetch: refetchTax }] = createResource(
    () => ({
      country: country(),
      subtotal: totalPrice(),
      vat: isB2B() ? vatNumber() : '',
    }),
    async (p) => {
      try {
        return await api.post<any>('/taxes/calculate', {
          country_code: p.country,
          subtotal: p.subtotal,
          vat_number: p.vat,
        });
      } catch {
        return null;
      }
    }
  );

  // Kortingscode toepassen
  const handleApplyCoupon = async (e: Event) => {
    e.preventDefault();
    if (!couponInput()) return;
    try {
      const res = await api.post<any>('/coupons/validate', {
        code: couponInput(),
        subtotal: totalPrice(),
      });
      setAppliedCoupon({ code: res.code, discount_amount: res.discount_amount });
      alert(`Kortingscode ${res.code} toegepast (-${formatPrice(res.discount_amount)})`);
    } catch (err: any) {
      alert('Kortingscode ongeldig: ' + err.message);
    }
  };

  // VIES B2B validatie
  const handleVerifyVat = async () => {
    if (!vatNumber()) return;
    try {
      const res = await api.post<any>('/taxes/validate-vies', { vat_number: vatNumber() });
      if (res.isValid) {
        setViesCompany(res.name);
        refetchTax();
        alert(`Btw-nummer geverifieerd: ${res.name}. Btw wordt verlegd naar 0%.`);
      } else {
        alert('Ongeldig btw-nummer volgens EU VIES database');
      }
    } catch (err: any) {
      alert('Fout bij controleren btw-nummer: ' + err.message);
    }
  };

  // Bereken totalen
  const discountAmount = () => (appliedCoupon() ? appliedCoupon()!.discount_amount : 0);
  const finalSubtotal = () => Math.max(0, totalPrice() - discountAmount());
  const finalTotal = () => finalSubtotal() + selectedShippingCost();

  // Bestelling definitief plaatsen & doorsturen naar betaalprovider
  const handlePlaceOrder = async (e: Event) => {
    e.preventDefault();
    setErrorMsg('');

    if (items().length === 0) {
      setErrorMsg(t('checkout_empty_err'));
      return;
    }

    setIsSubmitting(true);

    const shippingAddress = {
      full_name: fullName(),
      street: street(),
      house_number: houseNumber(),
      bus: bus() || null,
      postal_code: postalCode(),
      city: city(),
      country: country(),
      vat_number: isB2B() ? vatNumber() : null,
      company_name: viesCompany(),
    };

    const orderPayload = {
      guest_email: user() ? undefined : email(),
      shipping_address: shippingAddress,
      billing_address: shippingAddress,
      items: items().map((i) => ({
        product_id: i.productId,
        variant_id: i.variantId,
        quantity: i.quantity,
      })),
      shipping_cost: selectedShippingCost(),
      payment_provider: selectedPaymentMethod(),
    };

    try {
      const order = await api.post<any>('/orders', orderPayload);
      clearCart();

      // Start de betaling (bij Mollie, Stripe of Mock)
      const payResult = await api.post<any>('/payments/initiate', {
        order_number: order.order_number,
        method: selectedPaymentMethod(),
      });

      if (payResult.redirect_url) {
        window.location.href = payResult.redirect_url;
      } else {
        window.location.href = `/track/${order.order_number}`;
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'Fout bij plaatsen van bestelling');
      setIsSubmitting(false);
    }
  };

  return (
    <div class="checkout-page container">
      <div class="checkout-grid">
        {/* Adres, Bezorging & Betaalmethoden */}
        <form class="form-section" onSubmit={handlePlaceOrder}>
          {/* 1. Verzendadres */}
          <div class="section-block glass-panel">
            <h3><MapPin size={20} color="var(--accent)" /> {t('checkout_shipping_title')}</h3>
            
            <Show when={errorMsg()}>
              <div style={{ color: 'var(--danger)', 'font-size': '0.9rem', 'font-weight': 700 }}>
                {errorMsg()}
              </div>
            </Show>

            <div class="form-row">
              <div class="form-field">
                <label>{t('checkout_full_name')} *</label>
                <input type="text" required class="glass-input" value={fullName()} onInput={(e) => setFullName(e.currentTarget.value)} />
              </div>
              <div class="form-field">
                <label>{t('checkout_email')} *</label>
                <input type="email" required class="glass-input" value={email()} onInput={(e) => setEmail(e.currentTarget.value)} />
              </div>
            </div>

            <div class="form-row three-cols">
              <div class="form-field">
                <label>{t('checkout_street')} *</label>
                <input type="text" required class="glass-input" value={street()} onInput={(e) => setStreet(e.currentTarget.value)} />
              </div>
              <div class="form-field">
                <label>{t('checkout_number')} *</label>
                <input type="text" required class="glass-input" value={houseNumber()} onInput={(e) => setHouseNumber(e.currentTarget.value)} />
              </div>
              <div class="form-field">
                <label>{t('checkout_bus')}</label>
                <input type="text" class="glass-input" value={bus()} onInput={(e) => setBus(e.currentTarget.value)} />
              </div>
            </div>

            <div class="form-row three-cols">
              <div class="form-field">
                <label>{t('checkout_postal')} *</label>
                <input type="text" required class="glass-input" value={postalCode()} onInput={(e) => { setPostalCode(e.currentTarget.value); refetchShipping(); }} />
              </div>
              <div class="form-field">
                <label>{t('checkout_city')} *</label>
                <input type="text" required class="glass-input" value={city()} onInput={(e) => setCity(e.currentTarget.value)} />
              </div>
              <div class="form-field">
                <label>Land *</label>
                <select class="glass-input" value={country()} onChange={(e) => { setCountry(e.currentTarget.value); refetchShipping(); refetchTax(); }}>
                  <option value="BE">België</option>
                  <option value="NL">Nederland</option>
                  <option value="DE">Duitsland</option>
                  <option value="FR">Frankrijk</option>
                  <option value="US">Verenigde Staten</option>
                </select>
              </div>
            </div>

            {/* B2B VIES Btw-verlegging optie */}
            <div class="b2b-box">
              <div style={{ display: 'flex', 'align-items': 'center', gap: '0.6rem' }}>
                <input type="checkbox" id="b2b_cb" checked={isB2B()} onChange={(e) => { setIsB2B(e.currentTarget.checked); refetchTax(); }} style={{ width: '16px', height: '16px' }} />
                <label for="b2b_cb" style={{ "font-weight": 700, "font-size": "0.85rem" }}>Zakelijke aankoop (Btw-vrijstelling)</label>
              </div>
              <Show when={isB2B()}>
                <div style={{ display: 'flex', gap: '0.5rem', 'margin-top': '0.3rem' }}>
                  <input type="text" placeholder="Btw-nummer (bijv. NL123456789B01)" class="glass-input" value={vatNumber()} onInput={(e) => setVatNumber(e.currentTarget.value)} />
                  <button type="button" class="glass-button secondary" onClick={handleVerifyVat}>Verifieer</button>
                </div>
                <Show when={viesCompany()}>
                  <span style={{ "font-size": "0.8rem", color: "var(--success)", "font-weight": 700 }}>✓ Geverifieerd: {viesCompany()} (0% Btw)</span>
                </Show>
              </Show>
            </div>
          </div>

          {/* 2. Kies Verzendmethode */}
          <div class="section-block glass-panel">
            <h3><Truck size={20} color="var(--accent)" /> Kies Bezorgoptie</h3>
            <div class="options-grid">
              <For each={shippingQuotes()} fallback={<p style={{ color: "var(--text-muted)", "font-size": "0.85rem" }}>Verzendopties berekenen...</p>}>
                {(sq) => (
                  <div
                    class="option-card"
                    classList={{ selected: selectedShippingId() === sq.id }}
                    onClick={() => {
                      setSelectedShippingId(sq.id);
                      setSelectedShippingCost(sq.cost);
                    }}
                  >
                    <div class="opt-left">
                      <input type="radio" checked={selectedShippingId() === sq.id} />
                      <div>
                        <div class="opt-title">{sq.title}</div>
                        <div class="opt-desc">{sq.description}</div>
                      </div>
                    </div>
                    <div class="opt-price">
                      {sq.cost === 0 ? t('checkout_free') : formatPrice(sq.cost)}
                    </div>
                  </div>
                )}
              </For>
            </div>
          </div>

          {/* 3. Kies Betaalmethode */}
          <div class="section-block glass-panel">
            <h3><CreditCard size={20} color="var(--accent)" /> {t('checkout_payment_title')}</h3>
            <div class="options-grid">
              <For each={paymentMethods()}>
                {(pm) => (
                  <div
                    class="option-card"
                    classList={{ selected: selectedPaymentMethod() === pm.id }}
                    onClick={() => setSelectedPaymentMethod(pm.id)}
                  >
                    <div class="opt-left">
                      <input type="radio" checked={selectedPaymentMethod() === pm.id} />
                      <div>
                        <div class="opt-title">{pm.title}</div>
                        <div class="opt-desc">{pm.description}</div>
                      </div>
                    </div>
                  </div>
                )}
              </For>
            </div>

            <button type="submit" class="glass-button" disabled={isSubmitting()} style={{ height: '52px', 'font-size': '1.05rem', 'margin-top': '0.75rem' }}>
              <ShieldCheck size={20} /> {isSubmitting() ? t('checkout_btn_processing') : `${t('checkout_btn_place')} (${formatPrice(finalTotal())})`}
            </button>
          </div>
        </form>

        {/* Besteloverzicht & Kortingsbon */}
        <div class="summary-section glass-panel">
          <h3>{t('checkout_summary_title')} ({items().length} {t('checkout_items_count')})</h3>
          
          {/* Kortingscode Formulier */}
          <form onSubmit={handleApplyCoupon} class="coupon-box">
            <input
              type="text"
              placeholder="KORTINGSCODE..."
              class="glass-input"
              value={couponInput()}
              onInput={(e) => setCouponInput(e.currentTarget.value)}
            />
            <button type="submit" class="glass-button secondary">Toepassen</button>
          </form>

          <div class="items-list">
            <For each={items()}>
              {(item) => (
                <div class="summary-item">
                  <img src={item.image || '/static/tshirt-placeholder.jpg'} alt={item.name} />
                  <div class="info">
                    <div class="name">{item.name}</div>
                    <div class="meta">{item.variantTitle || 'Standard'} × {item.quantity}</div>
                  </div>
                  <div class="price">{formatPrice(item.price * item.quantity)}</div>
                </div>
              )}
            </For>
          </div>

          <div class="cost-breakdown">
            <div class="row">
              <span>{t('subtotal')}</span>
              <span>{formatPrice(totalPrice())}</span>
            </div>

            <Show when={appliedCoupon()}>
              <div class="row discount">
                <span>Korting ({appliedCoupon()!.code})</span>
                <span>-{formatPrice(appliedCoupon()!.discount_amount)}</span>
              </div>
            </Show>

            <div class="row">
              <span>{t('checkout_shipping_cost')}</span>
              <span>{selectedShippingCost() === 0 ? t('checkout_free') : formatPrice(selectedShippingCost())}</span>
            </div>

            <Show when={taxData()}>
              <div class="row" style={{ "font-size": "0.82rem" }}>
                <span>{taxData()?.tax_name}</span>
                <span>{taxData()?.is_reverse_charge ? '0% (Btw verlegd)' : formatPrice(taxData()?.tax_amount ?? 0)}</span>
              </div>
            </Show>

            <div class="row total">
              <span>{t('checkout_total')}</span>
              <span>{formatPrice(finalTotal())}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};