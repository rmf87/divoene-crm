import { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuthContext } from '../hooks/useAuth';

export default function AuthGate({ children }: { children: ReactNode }) {
  const { user, loading } = useAuthContext();
  if (loading) return <div id="auth-loading" className="p-8 text-olive-900">Carregando...</div>;
  if (!user) return <Navigate to="/login" replace />;
  return <>{children}</>;
}
