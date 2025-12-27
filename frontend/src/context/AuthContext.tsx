import React, { createContext, useContext, useState, useEffect } from 'react';
import type { User, AuthState } from '../types';
import { mockApi } from '../services/mockApi';

interface AuthContextType extends AuthState {
  login: (email: string, pass: string) => Promise<void>;
  register: (email: string, pass: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, setState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    isLoading: true,
  });

  useEffect(() => {
    // Check for existing session
    const initAuth = async () => {
      try {
        const user = await mockApi.auth.getCurrentUser();
        if (user) {
          setState({ user, isAuthenticated: true, isLoading: false });
        } else {
          setState({ user: null, isAuthenticated: false, isLoading: false });
        }
      } catch (error) {
        setState({ user: null, isAuthenticated: false, isLoading: false });
      }
    };
    initAuth();
  }, []);

  const login = async (email: string, pass: string) => {
    setState(prev => ({ ...prev, isLoading: true }));
    try {
      const user = await mockApi.auth.login(email, pass);
      setState({ user, isAuthenticated: true, isLoading: false });
    } catch (error) {
      setState(prev => ({ ...prev, isLoading: false }));
      throw error;
    }
  };

  const register = async (email: string, pass: string) => {
    setState(prev => ({ ...prev, isLoading: true }));
    try {
      const user = await mockApi.auth.register(email, pass);
      setState({ user, isAuthenticated: true, isLoading: false });
    } catch (error) {
      setState(prev => ({ ...prev, isLoading: false }));
      throw error;
    }
  };

  const logout = async () => {
    await mockApi.auth.logout();
    setState({ user: null, isAuthenticated: false, isLoading: false });
  };

  return (
    <AuthContext.Provider value={{ ...state, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
