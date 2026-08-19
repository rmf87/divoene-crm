import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { List, Moon, SignOut, Sun, X } from '@phosphor-icons/react';
import { useAuthContext, logout } from '../hooks/useAuth';
import { useTheme } from '../hooks/useTheme';

interface NavItem {
  to: string;
  label: string;
  roles: string[];
}

const ALL_ROLES = ['manager', 'seller', 'associate', 'guide'];

const navItems: NavItem[] = [
  { to: '/pipeline', label: 'Pipeline', roles: ['seller', 'manager'] },
  { to: '/my-leads', label: 'Meus Leads', roles: ['associate', 'manager'] },
  { to: '/visits', label: 'Minhas Visitas', roles: ['guide'] },
  { to: '/manage-visits', label: 'Gerenciar Visitas', roles: ['manager'] },
  { to: '/availability', label: 'Disponibilidade', roles: ['guide', 'manager'] },
  { to: '/reports', label: 'Relatórios', roles: ['manager'] },
  { to: '/config', label: 'Configurações', roles: ['manager'] },
  { to: '/ajuda', label: 'Ajuda', roles: ALL_ROLES },
];

const roleLabels: Record<string, string> = {
  manager: 'Administrador',
  seller: 'Vendedor',
  associate: 'Representante',
  guide: 'Guia',
};

export default function AdminHeader() {
  const { user } = useAuthContext();
  const location = useLocation();
  const [menuOpen, setMenuOpen] = useState(false);
  const { theme, toggle } = useTheme();

  if (!user) return null;

  const visible = navItems.filter(item => item.roles.some(r => user.roles.includes(r as any)));
  const roleLabel = user.roles.map(r => roleLabels[r] || r).join(', ');

  const linkClass = (to: string, base: string) =>
    `${base} transition-colors ${location.pathname === to ? 'bg-olive-800 text-white' : 'text-white/70 hover:text-white hover:bg-olive-800/60'}`;

  return (
    <header id="admin-header" className="bg-olive-900 text-white sticky top-0 z-40">
      <nav className="w-full px-4 md:px-6 py-3 flex items-center justify-between gap-4">
        <div className="flex items-center gap-4 min-w-0">
          <Link to="/" className="font-serif text-lg italic font-bold shrink-0" onClick={() => setMenuOpen(false)}>Divoene</Link>
          <div className="hidden lg:flex gap-1 text-sm">
            {visible.map(item => (
              <Link
                key={item.to}
                id={`nav-${item.to.replace('/', '')}`}
                to={item.to}
                className={`px-3 py-1.5 rounded-lg ${linkClass(item.to, 'font-medium')}`}
              >
                {item.label}
              </Link>
            ))}
          </div>
        </div>

        <div className="flex items-center gap-3 shrink-0">
          <button
            id="admin-theme-toggle"
            onClick={toggle}
            aria-label={theme === 'dark' ? 'Ativar tema claro' : 'Ativar tema escuro'}
            title={theme === 'dark' ? 'Tema claro' : 'Tema escuro'}
            className="p-2 rounded-lg text-white/70 hover:text-white hover:bg-olive-800 transition-colors"
          >
            {theme === 'dark' ? <Sun size={18} weight="bold" /> : <Moon size={18} weight="bold" />}
          </button>
          <span id="admin-user-role" className="text-xs text-white/50 hidden sm:block truncate max-w-[180px]">
            {roleLabel}
          </span>
          <button
            id="admin-logout"
            onClick={logout}
            className="text-xs text-white/70 hover:text-terracotta-400 transition-colors border border-white/30 rounded-full px-3 py-1.5 flex items-center gap-1.5 hover:border-terracotta-400"
          >
            <SignOut size={14} weight="bold" />
            <span className="hidden sm:inline">Sair</span>
          </button>
          <button
            id="admin-menu-toggle"
            onClick={() => setMenuOpen(!menuOpen)}
            className="lg:hidden p-2 rounded-lg hover:bg-olive-800 text-white"
            aria-label="Menu"
            aria-expanded={menuOpen}
          >
            {menuOpen ? <X size={20} weight="bold" /> : <List size={20} weight="bold" />}
          </button>
        </div>
      </nav>

      {menuOpen && (
        <div id="admin-mobile-menu" className="lg:hidden border-t border-olive-800 px-4 py-2 flex flex-col gap-1">
          {visible.map(item => (
            <Link
              key={item.to}
              to={item.to}
              onClick={() => setMenuOpen(false)}
              className={`px-3 py-2 rounded-lg text-sm ${linkClass(item.to, 'font-medium')}`}
            >
              {item.label}
            </Link>
          ))}
        </div>
      )}
    </header>
  );
}
