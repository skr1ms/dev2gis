'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getStoredToken, getUserFromToken, removeStoredToken, isTokenExpired } from '@/utils/auth';

interface User {
  id: string;
  email: string;
  name: string;
  role: string;
}

export function useAuth(requireAuth: boolean = false, requireAdmin: boolean = false) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [token, setToken] = useState<string | null>(null);
  const router = useRouter();
  
  useEffect(() => {
    const checkAuth = () => {
      const storedToken = getStoredToken();
      
      if (!storedToken) {
        setLoading(false);
        if (requireAuth) {
          router.push('/auth/login');
        }
        return;
      }
      
      if (isTokenExpired(storedToken)) {
        removeStoredToken();
        setLoading(false);
        if (requireAuth) {
          router.push('/auth/login');
        }
        return;
      }
      
      const userData = getUserFromToken(storedToken);
      
      if (!userData) {
        removeStoredToken();
        setLoading(false);
        if (requireAuth) {
          router.push('/auth/login');
        }
        return;
      }
      
      if (requireAdmin && userData.role !== 'admin') {
        setLoading(false);
        router.push('/');
        return;
      }
      
      setUser(userData);
      setToken(storedToken);
      setLoading(false);
    };
    
    checkAuth();
  }, [requireAuth, requireAdmin, router]);
  
  const logout = () => {
    removeStoredToken();
    setUser(null);
    setToken(null);
    router.push('/auth/login');
  };
  
  const isAdmin = user?.role === 'admin';
  const isAuthenticated = !!user && !!token;
  
  return {
    user,
    token,
    loading,
    isAuthenticated,
    isAdmin,
    logout,
  };
}

