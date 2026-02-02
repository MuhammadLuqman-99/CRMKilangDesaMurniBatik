// ============================================
// Authentication Context
// Production-Ready Auth State Management
// ============================================

import {
    createContext,
    useContext,
    useState,
    useEffect,
    useCallback,
    type ReactNode,
} from 'react';
import { authService } from '../services';
import type { User, LoginRequest, RegisterRequest, ProfileUpdateRequest } from '../types';

interface AuthContextType {
    user: User | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    error: string | null;
    login: (credentials: LoginRequest) => Promise<void>;
    register: (data: RegisterRequest) => Promise<void>;
    logout: () => Promise<void>;
    clearError: () => void;
    refreshUser: () => Promise<void>;
    updateProfile: (data: ProfileUpdateRequest) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
    children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const isAuthenticated = !!user;

    // Load user on mount if token exists
    useEffect(() => {
        const loadUser = async () => {
            if (authService.isAuthenticated()) {
                try {
                    const userData = await authService.getProfile();
                    setUser(userData);
                } catch (err) {
                    console.error('Failed to load user:', err);
                    authService.clearAuth();
                }
            }
            setIsLoading(false);
        };

        loadUser();
    }, []);

    const login = useCallback(async (credentials: LoginRequest) => {
        setIsLoading(true);
        setError(null);

        try {
            const response = await authService.login(credentials);
            setUser(response.user);
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Login failed';
            setError(message);
            throw err;
        } finally {
            setIsLoading(false);
        }
    }, []);

    const register = useCallback(async (data: RegisterRequest) => {
        setIsLoading(true);
        setError(null);

        try {
            await authService.register(data);
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Registration failed';
            setError(message);
            throw err;
        } finally {
            setIsLoading(false);
        }
    }, []);

    const logout = useCallback(async () => {
        setIsLoading(true);
        try {
            await authService.logout();
        } finally {
            setUser(null);
            setIsLoading(false);
        }
    }, []);

    const clearError = useCallback(() => {
        setError(null);
    }, []);

    const refreshUser = useCallback(async () => {
        if (authService.isAuthenticated()) {
            try {
                const userData = await authService.getProfile();
                setUser(userData);
            } catch (err) {
                console.error('Failed to refresh user:', err);
            }
        }
    }, []);

    const updateProfile = useCallback(async (data: ProfileUpdateRequest) => {
        try {
            const updatedUser = await authService.updateProfile(data);
            setUser(updatedUser);
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to update profile';
            throw new Error(message);
        }
    }, []);

    const value: AuthContextType = {
        user,
        isAuthenticated,
        isLoading,
        error,
        login,
        register,
        logout,
        clearError,
        refreshUser,
        updateProfile,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
    const context = useContext(AuthContext);

    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }

    return context;
}

export default AuthContext;
