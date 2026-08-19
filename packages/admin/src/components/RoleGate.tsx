import { ReactNode } from 'react';
import { useAuthContext } from '../hooks/useAuth';

interface Props {
  roles: string[];
  children: ReactNode;
  fallback?: ReactNode;
}

export default function RoleGate({ roles, children, fallback }: Props) {
  const { user } = useAuthContext();
  if (!user || !roles.some(r => user.roles.includes(r as any))) return fallback ? <>{fallback}</> : null;
  return <>{children}</>;
}
