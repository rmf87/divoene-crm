import { useState } from 'react';
import { login, getHomePath } from '../hooks/useAuth';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const auth = await login(email, password);
      window.location.href = getHomePath(auth.user!.roles);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div id="page-login" className="min-h-screen flex items-center justify-center bg-earth-50 px-4">
      <div className="max-w-sm w-full space-y-8">
        <div className="text-center">
          <h1 id="login-heading" className="font-serif text-3xl italic text-olive-900">
            Chácara Divoene
          </h1>
          <p id="login-subtitle" className="mt-2 text-olive-700">Painel de vendas</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="login-email" className="block text-sm font-medium text-olive-700 mb-1">
              Email
            </label>
            <input
              id="login-email"
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
              autoFocus
              className="w-full px-4 py-3 rounded-xl border border-earth-200 bg-white text-olive-900 focus:outline-none focus:ring-2 focus:ring-olive-500"
              placeholder="seu@email.com"
            />
          </div>

          <div>
            <label htmlFor="login-password" className="block text-sm font-medium text-olive-700 mb-1">
              Senha
            </label>
            <input
              id="login-password"
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              className="w-full px-4 py-3 rounded-xl border border-earth-200 bg-white text-olive-900 focus:outline-none focus:ring-2 focus:ring-olive-500"
              placeholder="••••••••"
            />
          </div>

          {error && (
            <p id="login-error" className="text-red-600 text-sm text-center">{error}</p>
          )}

          <button
            id="login-submit"
            type="submit"
            disabled={loading}
            className="w-full bg-olive-600 text-white font-semibold py-3 rounded-xl hover:bg-olive-700 disabled:opacity-50 transition-colors"
          >
            {loading ? 'Entrando...' : 'Entrar'}
          </button>
        </form>

        {import.meta.env.DEV && (
          <p className="text-xs text-olive-500 text-center mt-4">
            Dev: admin@divoene.test / admin
          </p>
        )}
      </div>
    </div>
  );
}
