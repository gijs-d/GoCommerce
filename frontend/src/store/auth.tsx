import { createSignal } from 'solid-js';
import { api } from '../lib/api';

export interface User {
  id: string;
  email: string;
  full_name: string;
  role: 'customer' | 'admin';
}

const [currentUser, setCurrentUser] = createSignal<User | null>(null);
const [isLoading, setIsLoading] = createSignal<boolean>(true);

export async function checkAuth() {
  setIsLoading(true);
  try {
    const user = await api.get<User>('/auth/me');
    setCurrentUser(user);
  } catch {
    setCurrentUser(null);
  } finally {
    setIsLoading(false);
  }
}

export function useAuth() {
  const login = async (email: string, password: string) => {
    const data = await api.post<{ user: User }>('/auth/login', { email, password });
    setCurrentUser(data.user);
    return data.user;
  };

  const register = async (fullName: string, email: string, password: string) => {
    const user = await api.post<User>('/auth/register', {
      full_name: fullName,
      email,
      password,
    });
    setCurrentUser(user);
    return user;
  };

  const logout = async () => {
    try {
      await api.post('/auth/logout');
    } finally {
      setCurrentUser(null);
    }
  };

  return {
    user: currentUser,
    isLoading,
    isAuthenticated: () => currentUser() !== null,
    isAdmin: () => currentUser()?.role === 'admin',
    login,
    register,
    logout,
    checkAuth,
  };
}