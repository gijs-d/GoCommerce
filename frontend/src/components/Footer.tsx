import { Component } from 'solid-js';
import { A } from '@solidjs/router';
import { useI18n } from '../store/i18n';

export const Footer: Component = () => {
  const { t } = useI18n();

  return (
    <footer style={{
      "margin-top": "auto",
      "background": "var(--card-bg)",
      "border-top": "1px solid var(--glass-border)",
      "padding": "3rem 0 2rem",
      "font-size": "0.88rem",
      "color": "var(--text-muted)"
    }}>
      <div class="container" style={{
        display: "grid",
        "grid-template-columns": "repeat(auto-fit, minmax(200px, 1fr))",
        gap: "2rem",
        "margin-bottom": "2rem"
      }}>
        <div>
          <div style={{ "font-size": "1.1rem", "font-weight": 800, color: "var(--text-main)", "margin-bottom": "0.5rem" }}>
            AESTHETIC
          </div>
          <p>{t('footer_about')}</p>
        </div>
        <div>
          <div style={{ "font-weight": 700, color: "var(--text-main)", "margin-bottom": "0.75rem" }}>{t('footer_shop')}</div>
          <p><A href="/shop">{t('all_products')}</A></p>
          <p><A href="/shop?category=oversized">{t('oversized')}</A></p>
          <p><A href="/shop?category=hoodies">{t('hoodies')}</A></p>
        </div>
        <div>
          <div style={{ "font-weight": 700, color: "var(--text-main)", "margin-bottom": "0.75rem" }}>{t('footer_service')}</div>
          <p><A href="/track">{t('track_order')}</A></p>
          <p><A href="/account/wishlist">{t('wishlist')}</A></p>
          <p><A href="/account">{t('account')}</A></p>
        </div>
        <div>
          <div style={{ "font-weight": 700, color: "var(--text-main)", "margin-bottom": "0.75rem" }}>{t('footer_security')}</div>
          <p>{t('footer_ssl')}</p>
          <p>{t('footer_shipping')}</p>
        </div>
      </div>
      <div class="container" style={{ "border-top": "1px solid var(--glass-border-subtle)", "padding-top": "1.5rem", "text-align": "center", "font-size": "0.8rem" }}>
        © 2026 AESTHETIC Studios. {t('footer_rights')}
      </div>
    </footer>
  );
};