import { Component, Show } from 'solid-js';
import { A, useLocation } from '@solidjs/router';
import { X, Home, ShoppingBag, Heart, PackageSearch, User, Shield, LogOut, Layers, Sparkles } from 'lucide-solid';
import { useAuth } from '../store/auth';
import { useCart } from '../store/cart';
import { useI18n } from '../store/i18n';
import './AsideMenu.scss';

interface AsideMenuProps {
  isOpen: boolean;
  onClose: () => void;
}

export const AsideMenu: Component<AsideMenuProps> = (props) => {
  const { user, isAdmin, logout } = useAuth();
  const { totalCount, toggleCart } = useCart();
  const { lang, toggleLanguage, t } = useI18n();
  const location = useLocation();

  const isCurrent = (path: string) => {
    if (path === '/') return location.pathname === '/';
    return location.pathname.startsWith(path);
  };

  return (
    <Show when={props.isOpen}>
      <div class="aside-backdrop" onClick={props.onClose}>
        <aside class="boer-style-drawer" onClick={(e) => e.stopPropagation()}>
          {/* Header bovenaan de drawer */}
          <div class="drawer-top-banner">
            <div class="brand-group">
              <div class="brand-icon">A</div>
              <span class="brand-title">AESTHETIC</span>
            </div>
            <div class="top-controls">
              <button class="lang-btn" onClick={toggleLanguage} title="Switch language">
                <span>{lang() === 'nl' ? '🇳🇱' : '🇬🇧'}</span>
                <span>{lang().toUpperCase()}</span>
              </button>
              <button class="close-icon" onClick={props.onClose} aria-label="Sluit menu">
                <X size={22} />
              </button>
            </div>
          </div>

          {/* Account tegel (volle breedte) */}
          <A
            href={user() ? '/account' : '/login'}
            class="drawer-account-box"
            onClick={props.onClose}
          >
            <div class="avatar-wrap">
              <User size={19} />
            </div>
            <div class="account-meta">
              <span class="account-label">{t('account')}</span>
              <span class="account-name">
                {user() ? user()!.full_name : t('login_register')}
              </span>
            </div>
          </A>

          {/* Menulijst met 100% brede knoppen */}
          <nav class="drawer-nav-list">
            <A
              href="/"
              class="menu-item-link"
              classList={{ active: isCurrent('/') }}
              onClick={props.onClose}
            >
              <span class="item-icon"><Home size={20} /></span>
              <span>{t('home')}</span>
            </A>

            <A
              href="/shop"
              class="menu-item-link"
              classList={{ active: isCurrent('/shop') && !location.search }}
              onClick={props.onClose}
            >
              <span class="item-icon"><ShoppingBag size={20} /></span>
              <span>{t('all_products')}</span>
            </A>

            <A
              href="/shop?category=oversized"
              class="menu-item-link"
              classList={{ active: location.search.includes('oversized') }}
              onClick={props.onClose}
            >
              <span class="item-icon"><Sparkles size={20} /></span>
              <span>{t('oversized')}</span>
            </A>

            <A
              href="/shop?category=hoodies"
              class="menu-item-link"
              classList={{ active: location.search.includes('hoodies') }}
              onClick={props.onClose}
            >
              <span class="item-icon"><Layers size={20} /></span>
              <span>{t('hoodies')}</span>
            </A>

            <A
              href="/track"
              class="menu-item-link"
              classList={{ active: isCurrent('/track') }}
              onClick={props.onClose}
            >
              <span class="item-icon"><PackageSearch size={20} /></span>
              <span>{t('track_order')}</span>
            </A>

            <A
              href="/account/wishlist"
              class="menu-item-link"
              classList={{ active: location.pathname.includes('wishlist') }}
              onClick={props.onClose}
            >
              <span class="item-icon"><Heart size={20} /></span>
              <span>{t('wishlist')}</span>
            </A>

            <button
              class="menu-item-link"
              onClick={() => {
                props.onClose();
                toggleCart();
              }}
            >
              <span class="item-icon"><ShoppingBag size={20} /></span>
              <span>{t('cart')} ({totalCount()})</span>
            </button>

            <Show when={isAdmin()}>
              <A
                href="/admin"
                class="menu-item-link highlight-admin"
                classList={{ active: isCurrent('/admin') }}
                onClick={props.onClose}
              >
                <span class="item-icon"><Shield size={20} /></span>
                <span>{t('admin_panel')}</span>
              </A>
            </Show>
          </nav>

          {/* Uitloggen onderaan */}
          <Show when={user()}>
            <div class="drawer-footer">
              <button
                class="logout-btn"
                onClick={() => {
                  props.onClose();
                  logout();
                }}
              >
                <span class="item-icon"><LogOut size={20} /></span>
                <span>{t('logout')}</span>
              </button>
            </div>
          </Show>
        </aside>
      </div>
    </Show>
  );
};