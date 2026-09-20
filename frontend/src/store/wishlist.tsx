import { createSignal, createEffect } from 'solid-js';
import { api } from '../lib/api';
import { useAuth } from './auth';

const [wishlistIds, setWishlistIds] = createSignal<Set<string>>(new Set());

export function useWishlist() {
  const { isAuthenticated } = useAuth();

  const fetchWishlist = async () => {
    if (!isAuthenticated()) {
      setWishlistIds(new Set<string>());
      return;
    }
    try {
      const items = await api.get<Array<{ product_id: string }>>('/user/wishlist');
      setWishlistIds(new Set(items.map((i) => i.product_id)));
    } catch {
      // sessie niet actief
    }
  };

  const toggleWishlist = async (productId: string) => {
    if (!isAuthenticated()) {
      alert('Log eerst in om favorieten op te slaan.');
      return;
    }

    const current = new Set(wishlistIds());
    if (current.has(productId)) {
      current.delete(productId);
      setWishlistIds(new Set(current));
      await api.delete(`/user/wishlist/${productId}`);
    } else {
      current.add(productId);
      setWishlistIds(new Set(current));
      await api.post(`/user/wishlist/${productId}`);
    }
  };

  const isWishlisted = (productId: string) => wishlistIds().has(productId);

  return {
    wishlistIds,
    fetchWishlist,
    toggleWishlist,
    isWishlisted,
  };
}