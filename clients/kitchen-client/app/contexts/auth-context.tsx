import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { getMe, login, logout, register, type User } from "~/api/auth";
import { clearTokens, getRefreshToken, setTokens } from "~/utils/token";

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signUp: (email: string, username: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (!getRefreshToken()) {
      setIsLoading(false);
      return;
    }
    getMe()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setIsLoading(false));
  }, []);

  const signIn = useCallback(async (email: string, password: string) => {
    const tokens = await login(email, password);
    setTokens(tokens.access_token, tokens.refresh_token);
    const me = await getMe();
    setUser(me);
  }, []);

  const signUp = useCallback(
    async (email: string, username: string, password: string) => {
      const tokens = await register(email, username, password);
      setTokens(tokens.access_token, tokens.refresh_token);
      const me = await getMe();
      setUser(me);
    },
    []
  );

  const signOut = useCallback(async () => {
    const token = getRefreshToken();
    if (token) await logout(token).catch(() => {});
    clearTokens();
    setUser(null);
    window.location.href = "/auth/sign-in";
  }, []);

  return (
    <AuthContext.Provider value={{ user, isLoading, signIn, signUp, signOut }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
