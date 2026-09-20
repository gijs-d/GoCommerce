import { Component, Show, createSignal, onMount, onCleanup } from 'solid-js';
import { A, useLocation } from '@solidjs/router';
import { ShoppingBag, Sun, Moon, Menu, Heart, User, Shield, ChevronDown, Check } from 'lucide-solid';
import { useTheme } from '../store/theme';
import { useCart } from '../store/cart';
import { useAuth } from '../store/auth';
import { useWishlist } from '../store/wishlist';
import { useI18n, languages } from '../store/i18n';
import { useCurrency, currencies, CurrencyCode } from '../store/currency';
import './Header.scss';

interface HeaderProps {
  onOpenMenu: () => void;
}

export const Header: Component<HeaderProps> = (props) => {
  const { isDark, toggleTheme } = useTheme();
  const { totalCount, formattedTotal, toggleCart } = useCart();
  const { user, isAdmin } = useAuth();
  const { wishlistIds } = useWishlist();
  const { lang, setLanguage, t } = useI18n();
  const { currency, setCurrency } = useCurrency();
  const location = useLocation();

  const [isScrolled, setIsScrolled] = createSignal(false);
  const [showLocaleDropdown, setShowLocaleDropdown] = createSignal(false);

  const handleScroll = () => {
    setIsScrolled(window.scrollY > 30);
  };

  const closeDropdown = (e: MouseEvent) => {
    const target = e.target as HTMLElement;
    if (!target.closest('.locale-dropdown-container')) {
      setShowLocaleDropdown(false);
    }
  };

  onMount(() => {
    window.addEventListener('scroll', handleScroll, { passive: true });
    window.addEventListener('click', closeDropdown);
    handleScroll();
  });

  onCleanup(() => {
    window.removeEventListener('scroll', handleScroll);
    window.removeEventListener('click', closeDropdown);
  });

  const isHome = () => location.pathname === '/';
  const isTransparent = () => isHome() && !isScrolled();
  const isActive = (path: string) => (path === '/' ? location.pathname === '/' : location.pathname.startsWith(path));

  const currentFlag = () => (lang() === 'nl' ? '🇳🇱' : '🇬🇧');

  return (
    <header class="site-header" classList={{ 'header-transparent': isTransparent() }}>
      <div class="container header-inner">
        <div class="header-left">
          <A href="/" class="brand-logo">
            <span class="logo-emblem">A</span>
            <span>AESTHETIC</span>
          </A>

          <nav class="desktop-nav">
            <A href="/shop" class="nav-link" classList={{ active: isActive('/shop') && !location.search }}>
              {t('collection')}
            </A>
            <A href="/shop?category=oversized" class="nav-link" classList={{ active: location.search.includes('oversized') }}>
              {t('oversized')}
            </A>
            <A href="/shop?category=hoodies" class="nav-link" classList={{ active: location.search.includes('hoodies') }}>
              {t('hoodies')}
            </A>
            <A href="/track" class="nav-link nav-extra" classList={{ active: isActive('/track') }}>
              {t('track_order')}
            </A>
          </nav>
        </div>

        <div class="header-right">
          {/* Dropdown voor Taal & Valuta (EUR / USD) */}
          <div class="locale-dropdown-container">
            <button
              class="locale-btn"
              onClick={(e) => {
                e.stopPropagation();
                setShowLocaleDropdown(!showLocaleDropdown());
              }}
              title="Taal & Valuta kiezen"
            >
              <span class="lang-flag">{currentFlag()}</span>
              <span class="locale-text">{lang().toUpperCase()} • {currencies[currency()].symbol}</span>
              <ChevronDown size={14} />
            </button>

            <Show when={showLocaleDropdown()}>
              <div class="locale-popover glass-panel">
                <div class="popover-section">
                  <div class="section-title">{t('language')}</div>
                  <div class="options-list">
                    {languages.map((l) => (
                      <button
                        class="option-row"
                        classList={{ active: lang() === l.code }}
                        onClick={() => {
                          setLanguage(l.code);
                        }}
                      >
                        <span class="flag-icon">{l.flag}</span>
                        <span class="option-name">{l.name}</span>
                        <Show when={lang() === l.code}>
                          <Check size={16} class="check-icon" />
                        </Show>
                      </button>
                    ))}
                  </div>
                </div>

                <div class="popover-divider"></div>

                <div class="popover-section">
                  <div class="section-title">{t('currency')}</div>
                  <div class="options-list">
                    {(Object.keys(currencies) as CurrencyCode[]).map((c) => (
                      <button
                        class="option-row"
                        classList={{ active: currency() === c }}
                        onClick={() => {
                          setCurrency(c);
                        }}
                      >
                        <span class="curr-symbol">{currencies[c].symbol}</span>
                        <span class="option-name">{currencies[c].name}</span>
                        <Show when={currency() === c}>
                          <Check size={16} class="check-icon" />
                        </Show>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </Show>
          </div>

          {/* Admin link */}
          <Show when={isAdmin()}>
            <A href="/admin" class="icon-btn" title={t('admin_panel')}>
              <Shield size={23} />
            </A>
          </Show>

          {/* Favorieten met badge */}
          <A href="/account/wishlist" class="icon-btn" title={t('wishlist')}>
            <div style={{ position: 'relative', display: 'flex' }}>
              <Heart size={23} fill={wishlistIds().size > 0 ? "currentColor" : "none"} />
              <Show when={wishlistIds().size > 0}>
                <span class="glass-badge" style={{ position: 'absolute', top: '-8px', right: '-9px', 'font-size': '0.62rem', padding: '1px 4px' }}>
                  {wishlistIds().size}
                </span>
              </Show>
            </div>
          </A>

          {/* Account link */}
          <A href={user() ? "/account" : "/login"} class="icon-btn" title={user() ? user()!.full_name : t('login_register')}>
            <User size={23} />
          </A>

          {/* Dark / Light Toggle */}
          <button class="icon-btn" onClick={toggleTheme} title="Thema wisselen" aria-label="Toggle theme">
            <Show when={isDark()} fallback={<Moon size={23} />}>
              <Sun size={23} />
            </Show>
          </button>

          {/* Winkelmandje met prijs in gekozen valuta */}
          <button class="cart-trigger" onClick={toggleCart} title={t('cart')}>
            <div class="cart-icon-wrapper">
              <ShoppingBag size={23} />
              <Show when={totalCount() > 0}>
                <span class="cart-badge">{totalCount()}</span>
              </Show>
            </div>
            <span class="cart-price">{formattedTotal()}</span>
          </button>

          {/* Menuknop rechts */}
          <button class="mobile-menu-trigger" onClick={props.onOpenMenu} aria-label="Open menu">
            <Menu size={24} />
          </button>
        </div>
      </div>
    </header>
  );
};