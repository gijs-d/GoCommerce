import { createSignal, createMemo, createEffect } from 'solid-js';
import { useCurrency } from './currency';

export interface CartItem {
  id: string; // unique item key: productId-variantId
  productId: string;
  variantId?: string;
  name: string;
  variantTitle?: string;
  price: number; // altijd in EUR in de backend
  quantity: number;
  image?: string;
  size?: string;
  color?: string;
}

const STORAGE_KEY = 'webshop_cart_v1';

const [cartItems, setCartItems] = createSignal<CartItem[]>(
  (() => {
    if (typeof window === 'undefined') return [];
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      return stored ? JSON.parse(stored) : [];
    } catch {
      return [];
    }
  })()
);

const [isCartOpen, setIsCartOpen] = createSignal<boolean>(false);

createEffect(() => {
  if (typeof window !== 'undefined') {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(cartItems()));
  }
});

export function useCart() {
  const { formatPrice } = useCurrency();

  const totalCount = createMemo(() =>
    cartItems().reduce((acc, item) => acc + item.quantity, 0)
  );

  const totalPrice = createMemo(() =>
    cartItems().reduce((acc, item) => acc + item.price * item.quantity, 0)
  );

  const formattedTotal = createMemo(() => {
    return formatPrice(totalPrice());
  });

  const addItem = (item: Omit<CartItem, 'id'>) => {
    const key = `${item.productId}-${item.variantId || 'default'}`;
    setCartItems((prev) => {
      const existingIndex = prev.findIndex((i) => i.id === key);
      if (existingIndex > -1) {
        const copy = [...prev];
        copy[existingIndex].quantity += item.quantity;
        return copy;
      }
      return [...prev, { ...item, id: key }];
    });
    setIsCartOpen(true);
  };

  const updateQuantity = (key: string, qty: number) => {
    if (qty <= 0) {
      removeItem(key);
      return;
    }
    setCartItems((prev) =>
      prev.map((i) => (i.id === key ? { ...i, quantity: qty } : i))
    );
  };

  const removeItem = (key: string) => {
    setCartItems((prev) => prev.filter((i) => i.id !== key));
  };

  const clearCart = () => {
    setCartItems([]);
  };

  return {
    items: cartItems,
    isCartOpen,
    openCart: () => setIsCartOpen(true),
    closeCart: () => setIsCartOpen(false),
    toggleCart: () => setIsCartOpen((prev) => !prev),
    totalCount,
    totalPrice,
    formattedTotal,
    addItem,
    updateQuantity,
    removeItem,
    clearCart,
  };
}