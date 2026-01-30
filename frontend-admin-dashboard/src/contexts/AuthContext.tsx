// ============================================
// Authentication Context
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { authApi, setTokens, clearTokens, getAccessToken } from '../services/api';
import type { User } from '../types';

interface AuthContextType {
    user: User | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    login: (email: string, password: string) => Promise<void>;
    logout: () => Promise<void>;
    refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = (): AuthContextType => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

interface AuthProviderProps {
    children: React.ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    const refreshUser = useCallback(async () => {
        try {
            const token = getAccessToken();
            if (!token) {
                setUser(null);
                return;
            }

            const response = await authApi.getProfile();
            if (response.success && response.data.user) {
                setUser(response.data.user);
            } else {
                setUser(null);
                clearTokens();
            }
        } catch (error) {
            console.error('Failed to refresh user:', error);
            setUser(null);
            clearTokens();
        }
    }, []);

    // Check authentication on mount
    useEffect(() => {
        const initAuth = async () => {
            setIsLoading(true);
            await refreshUser();
            setIsLoading(false);
        };
        initAuth();
    }, [refreshUser]);

    const login = async (email: string, password: string): Promise<void> => {
        const response = await authApi.login(email, password);

        if (response.success && response.data) {
            setTokens(response.data.access_token, response.data.refresh_token);
            await refreshUser();
        } else {
            throw new Error('Login failed');
        }
    };

    const logout = async (): Promise<void> => {
        try {
            await authApi.logout();
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            clearTokens();
            setUser(null);
        }
    };

    const value: AuthContextType = {
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
        refreshUser,
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};

export default AuthContext;
