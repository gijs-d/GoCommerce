import { Component, createSignal, onMount } from 'solid-js';
import { Router, Route } from '@solidjs/router';
import { Header } from './components/Header';
import { AsideMenu } from './components/AsideMenu';
import { CartDrawer } from './components/CartDrawer';
import { Footer } from './components/Footer';
import { Home } from './pages/Home';
import { Shop } from './pages/Shop';
import { ProductDetail } from './pages/ProductDetail';
import { Checkout } from './pages/Checkout';
import { OrderTracking } from './pages/OrderTracking';
import { Account } from './pages/Account';
import { Auth } from './pages/Auth';
import { Admin } from './pages/Admin';
import { checkAuth } from './store/auth';
import { useWishlist } from './store/wishlist';

export const App: Component = () => {
  const [isMenuOpen, setIsMenuOpen] = createSignal(false);
  const { fetchWishlist } = useWishlist();

  onMount(() => {
    checkAuth().then(() => fetchWishlist());
  });

  return (
    <Router root={(props) => (
      <div style={{ display: 'flex', 'flex-direction': 'column', 'min-height': '100vh' }}>
        <Header onOpenMenu={() => setIsMenuOpen(true)} />
        <AsideMenu isOpen={isMenuOpen()} onClose={() => setIsMenuOpen(false)} />
        <CartDrawer />
        <main style={{ flex: 1, "min-height": "calc(100vh - 56px)", "padding-bottom": "50px" }}>
          {props.children}
        </main>
        <Footer />
      </div>
    )}>
      <Route path="/" component={Home} />
      <Route path="/shop" component={Shop} />
      <Route path="/product/:slug" component={ProductDetail} />
      <Route path="/checkout" component={Checkout} />
      <Route path="/track" component={OrderTracking} />
      <Route path="/track/:orderNumber" component={OrderTracking} />
      <Route path="/account" component={Account} />
      <Route path="/account/wishlist" component={Account} />
      <Route path="/login" component={Auth} />
      <Route path="/admin" component={Admin} />
    </Router>
  );
};