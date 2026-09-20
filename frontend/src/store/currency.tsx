import { createSignal } from 'solid-js';

export type CurrencyCode = 'EUR' | 'USD' | 'GBP';

export interface CurrencyConfig {
  code: CurrencyCode;
  symbol: string;
  name: string;
  rate: number; // t.o.v. EUR
}

export const currencies: Record<CurrencyCode, CurrencyConfig> = {
  EUR: { code: 'EUR', symbol: '€', name: 'Euro (€)', rate: 1.0 },
  USD: { code: 'USD', symbol: '$', name: 'US Dollar ($)', rate: 1.08 },
  GBP: { code: 'GBP', symbol: '£', name: 'British Pound (£)', rate: 0.85 },
};

const savedCurrency = (typeof window !== 'undefined' && (localStorage.getItem('preferred_currency') as CurrencyCode)) || 'EUR';
const [currentCurrency, setCurrentCurrency] = createSignal<CurrencyCode>(savedCurrency);

export function useCurrency() {
  const setCurrency = (c: CurrencyCode) => {
    setCurrentCurrency(c);
    if (typeof window !== 'undefined') {
      localStorage.setItem('preferred_currency', c);
    }
  };

  const formatPrice = (amountInEur: number) => {
    const config = currencies[currentCurrency()] || currencies.EUR;
    const converted = amountInEur * config.rate;
    const locale = currentCurrency() === 'EUR' ? 'nl-BE' : currentCurrency() === 'GBP' ? 'en-GB' : 'en-US';
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: config.code,
    }).format(converted);
  };

  return {
    currency: currentCurrency,
    setCurrency,
    formatPrice,
    currencies,
  };
}