import { useState, useEffect, useCallback } from 'react';
import { ArrowClockwise, CaretDown, DownloadSimple, Plus, Trash, UploadSimple, X } from '@phosphor-icons/react';
import { fetchWithAuth } from '../hooks/useAuth';

// ─── Types ────────────────────────────────────────────────────

interface ConfigSetting {
  key: string;
  value: string;
  masked_value: string;
  value_type: string;
  description: string;
  updated_at: string;
  updated_by: string;
}

interface User {
  id: string;
  email: string;
  name: string;
  roles: string[];
  active: boolean;
  created_at: string;
}

type Tab = 'secrets' | 'users' | 'backup';

interface SettingDef {
  key: string;
  label: string;
  category: string;
}

const SETTING_DEFS: SettingDef[] = [
  // Clicksign
  { key: 'clicksign_api_key', label: 'API Key', category: 'Clicksign' },
  { key: 'clicksign_base_url', label: 'Base URL', category: 'Clicksign' },
  { key: 'clicksign_webhook_secret', label: 'Webhook Secret', category: 'Clicksign' },
  { key: 'clicksign_template_key', label: 'Template Key', category: 'Clicksign' },
  // OpenPix
  { key: 'openpix_app_id', label: 'App ID', category: 'OpenPix / Woovi' },
  { key: 'openpix_base_url', label: 'Base URL', category: 'OpenPix / Woovi' },
  // WhatsApp
  { key: 'whatsapp_token', label: 'Access Token', category: 'WhatsApp' },
  { key: 'whatsapp_phone_number_id', label: 'Phone Number ID', category: 'WhatsApp' },
  { key: 'whatsapp_app_secret', label: 'App Secret', category: 'WhatsApp' },
  { key: 'whatsapp_webhook_verify_token', label: 'Webhook Verify Token', category: 'WhatsApp' },
  { key: 'whatsapp_base_url', label: 'Base URL', category: 'WhatsApp' },
];

const ROLE_LABELS: Record<string, string> = {
  manager: 'Administrador',
  seller: 'Vendedor',
  associate: 'Representante',
  guide: 'Guia',
};

const ROLE_OPTIONS = ['seller', 'associate', 'guide', 'manager'] as const;

const ROLE_BADGE_CLASSES: Record<string, string> = {
  manager: 'bg-purple-100 text-purple-700',
  seller: 'bg-blue-100 text-blue-700',
  guide: 'bg-green-100 text-green-700',
  associate: 'bg-earth-100 text-earth-700',
};

function RoleBadges({ roles }: { roles: string[] }) {
  return (
    <span className="inline-flex flex-wrap gap-1">
      {roles.map(role => (
        <span key={role} className={`text-xs px-2 py-0.5 rounded-full ${ROLE_BADGE_CLASSES[role] || 'bg-earth-100 text-earth-700'}`}>
          {ROLE_LABELS[role] || role}
        </span>
      ))}
    </span>
  );
}

function RolesCheckboxes({ value, onChange }: { value: string[]; onChange: (roles: string[]) => void }) {
  const toggle = (role: string) => {
    onChange(value.includes(role) ? value.filter(r => r !== role) : [...value, role]);
  };
  return (
    <div className="flex flex-wrap gap-3">
      {ROLE_OPTIONS.map(role => (
        <label key={role} className="flex items-center gap-1.5 text-sm text-olive-800 cursor-pointer">
          <input type="checkbox" checked={value.includes(role)} onChange={() => toggle(role)} className="accent-olive-600" />
          {ROLE_LABELS[role]}
        </label>
      ))}
    </div>
  );
}

// ─── Component ─────────────────────────────────────────────────

export default function Config() {
  const [activeTab, setActiveTab] = useState<Tab>('secrets');

  return (
    <div id="config-page" className="max-w-5xl mx-auto px-4 py-6">
      <h1 id="config-heading" className="font-serif text-2xl text-olive-900 italic mb-6">Configurações</h1>

      {/* Tabs */}
      <div id="config-tabs" className="flex gap-1 mb-6 border-b border-earth-300">
        <TabBtn id="config-tab-secrets" active={activeTab === 'secrets'} onClick={() => setActiveTab('secrets')}>
          Chaves API
        </TabBtn>
        <TabBtn id="config-tab-users" active={activeTab === 'users'} onClick={() => setActiveTab('users')}>
          Usuários
        </TabBtn>
        <TabBtn id="config-tab-backup" active={activeTab === 'backup'} onClick={() => setActiveTab('backup')}>
          Backup
        </TabBtn>
      </div>

      {activeTab === 'secrets' ? <SecretsPanel /> : activeTab === 'users' ? <UsersPanel /> : <BackupPanel />}
    </div>
  );
}

// ─── Tab Button ────────────────────────────────────────────────

function TabBtn({ active, onClick, children, id }: { active: boolean; onClick: () => void; children: React.ReactNode; id?: string }) {
  return (
    <button
      id={id}
      onClick={onClick}
      className={`px-4 py-2 text-sm font-medium rounded-t transition-colors ${
        active
          ? 'bg-white text-olive-900 border border-earth-300 border-b-white -mb-px'
          : 'text-olive-600 hover:text-olive-900'
      }`}
    >
      {children}
    </button>
  );
}

// ─── Secrets Panel ─────────────────────────────────────────────

function SecretsPanel() {
  const [settings, setSettings] = useState<ConfigSetting[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [saving, setSaving] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const fetchSettings = useCallback(async () => {
    try {
      const res = await fetchWithAuth('/api/config');
      if (!res.ok) throw new Error(`Erro ${res.status}`);
      const data = await res.json();
      setSettings(data);
      setError('');
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSettings();
  }, [fetchSettings]);

  const startEdit = (setting: ConfigSetting) => {
    setEditingKey(setting.key);
    setEditValue('');
    setShowPassword(false);
  };

  const cancelEdit = () => {
    setEditingKey(null);
    setEditValue('');
    setShowPassword(false);
  };

  const saveEdit = async (key: string) => {
    setSaving(true);
    try {
      const res = await fetchWithAuth(`/api/config/${key}`, {
        method: 'PUT',
        body: JSON.stringify({ value: editValue }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || `Erro ${res.status}`);
      }
      const updated = await res.json();
      setSettings(prev =>
        prev.map(s =>
          s.key === key ? { ...s, masked_value: updated.masked_value, value: updated.value, updated_at: updated.updated_at, updated_by: updated.updated_by } : s
        )
      );
      setEditingKey(null);
      setEditValue('');
    } catch (e: any) {
      alert(e.message);
    } finally {
      setSaving(false);
    }
  };

  // Group settings by category
  const categories = SETTING_DEFS.reduce<Record<string, SettingDef[]>>((acc, def) => {
    (acc[def.category] ||= []).push(def);
    return acc;
  }, {});

  if (loading) return <p className="text-olive-600 text-sm">Carregando...</p>;
  if (error) return <p className="text-red-600 text-sm">Erro: {error}</p>;

  return (
    <div className="space-y-6">
      {Object.entries(categories).map(([category, defs]) => (
        <section key={category}>
          <h2 className="text-sm font-semibold text-olive-800 mb-3 uppercase tracking-wide">{category}</h2>
          <div className="bg-white border border-earth-200 rounded-lg overflow-hidden">
            {defs.map((def) => {
              const setting = settings.find(s => s.key === def.key);
              const isEditing = editingKey === def.key;
              const isSecret = setting?.value_type === 'secret';
              const displayValue = setting?.masked_value || (setting?.value || '••••••••');

              return (
                <div key={def.key} className="flex items-center justify-between px-4 py-3 border-b border-earth-100 last:border-b-0 gap-4">
                  <div className="min-w-0 flex-1">
                    <p className="text-sm text-olive-900">{def.label}</p>
                    <p className="text-xs text-olive-500">{setting?.description || def.key}</p>
                  </div>

                  <div className="flex items-center gap-2">
                    {isEditing ? (
                      <div className="flex items-center gap-2">
                        <input
                          type={showPassword ? 'text' : 'password'}
                          value={editValue}
                          onChange={e => setEditValue(e.target.value)}
                          placeholder="Novo valor..."
                          className="text-sm border border-earth-300 rounded px-2 py-1 w-48 focus:outline-none focus:border-olive-500"
                          autoFocus
                        />
                        <button
                          type="button"
                          onClick={() => setShowPassword(!showPassword)}
                          className="text-xs text-olive-500 hover:text-olive-700"
                          title={showPassword ? 'Ocultar' : 'Mostrar'}
                        >
                          {showPassword ? '🙈' : '👁'}
                        </button>
                        <button
                          onClick={() => saveEdit(def.key)}
                          disabled={saving}
                          className="text-xs bg-olive-700 text-white px-3 py-1 rounded hover:bg-olive-800 disabled:opacity-50"
                        >
                          {saving ? '...' : 'Salvar'}
                        </button>
                        <button onClick={cancelEdit} className="text-xs text-olive-500 hover:text-olive-700">
                          Cancelar
                        </button>
                      </div>
                    ) : (
                      <>
                        <code className="text-xs bg-earth-100 px-2 py-1 rounded text-olive-700 max-w-[200px] truncate">
                          {displayValue}
                        </code>
                        <button
                          onClick={() => startEdit(setting!)}
                          className="text-xs text-olive-500 hover:text-olive-800 border border-earth-300 rounded px-2 py-1"
                        >
                          Editar
                        </button>
                      </>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </section>
      ))}
    </div>
  );
}

// ─── Users Panel ───────────────────────────────────────────────

function UsersPanel() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  // Create form
  const [newName, setNewName] = useState('');
  const [newEmail, setNewEmail] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [newRoles, setNewRoles] = useState<string[]>(['seller']);
  const [creating, setCreating] = useState(false);

  // Edit form
  const [editName, setEditName] = useState('');
  const [editRoles, setEditRoles] = useState<string[]>([]);
  const [editPassword, setEditPassword] = useState('');
  const [saving, setSaving] = useState(false);

  const fetchUsers = useCallback(async () => {
    try {
      const res = await fetchWithAuth('/api/users');
      if (!res.ok) throw new Error(`Erro ${res.status}`);
      const data = await res.json();
      setUsers(data);
      setError('');
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const handleCreate = async () => {
    if (!newName || !newEmail || !newPassword) {
      alert('Nome, email e senha são obrigatórios.');
      return;
    }
    setCreating(true);
    try {
      const res = await fetchWithAuth('/api/users', {
        method: 'POST',
        body: JSON.stringify({ name: newName, email: newEmail, password: newPassword, roles: newRoles }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || `Erro ${res.status}`);
      }
      setNewName('');
      setNewEmail('');
      setNewPassword('');
      setNewRoles(['seller']);
      setShowCreate(false);
      fetchUsers();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setCreating(false);
    }
  };

  const startEdit = (user: User) => {
    setEditingId(user.id);
    setEditName(user.name);
    setEditRoles(user.roles || []);
    setEditPassword('');
  };

  const cancelEdit = () => {
    setEditingId(null);
    setEditPassword('');
  };

  const handleSave = async (userId: string) => {
    setSaving(true);
    const body: Record<string, any> = { name: editName, roles: editRoles };
    if (editPassword) body.password = editPassword;

    try {
      const res = await fetchWithAuth(`/api/users/${userId}`, {
        method: 'PATCH',
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || `Erro ${res.status}`);
      }
      setEditingId(null);
      setEditPassword('');
      fetchUsers();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (user: User) => {
    if (!confirm(`Desativar usuário "${user.name}"?`)) return;
    try {
      const res = await fetchWithAuth(`/api/users/${user.id}`, { method: 'PATCH', body: JSON.stringify({ active: false }) });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || `Erro ${res.status}`);
      }
      fetchUsers();
    } catch (e: any) {
      alert(e.message);
    }
  };

  if (loading) return <p className="text-olive-600 text-sm">Carregando...</p>;
  if (error) return <p className="text-red-600 text-sm">Erro: {error}</p>;

  const activeUsers = users.filter(u => u.active);
  const inactiveUsers = users.filter(u => !u.active);

  return (
    <div className="space-y-6">
      {/* Create button */}
      <button
        id="user-create-toggle"
        onClick={() => setShowCreate(!showCreate)}
        className="text-sm bg-olive-700 text-white px-4 py-2 rounded hover:bg-olive-800 transition-colors"
      >
        {showCreate ? 'Cancelar' : '+ Novo Usuário'}
      </button>

      {/* Create form */}
      {showCreate && (
        <div id="user-create-form" className="bg-earth-50 border border-earth-200 rounded-lg p-4 space-y-3">
          <h3 className="text-sm font-semibold text-olive-900">Novo Usuário</h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <input id="user-create-input-name" value={newName} onChange={e => setNewName(e.target.value)} placeholder="Nome" className="text-sm border border-earth-300 rounded px-3 py-2 focus:outline-none focus:border-olive-500" />
            <input id="user-create-input-email" value={newEmail} onChange={e => setNewEmail(e.target.value)} placeholder="Email" type="email" className="text-sm border border-earth-300 rounded px-3 py-2 focus:outline-none focus:border-olive-500" />
            <input id="user-create-input-password" value={newPassword} onChange={e => setNewPassword(e.target.value)} placeholder="Senha" type="password" className="text-sm border border-earth-300 rounded px-3 py-2 focus:outline-none focus:border-olive-500" />
            <div id="user-create-roles" className="sm:col-span-2">
              <RolesCheckboxes value={newRoles} onChange={setNewRoles} />
            </div>
          </div>
          <button id="user-create-submit" onClick={handleCreate} disabled={creating || newRoles.length === 0} className="text-sm bg-terracotta-600 text-white px-4 py-2 rounded hover:bg-terracotta-700 disabled:opacity-50">
            {creating ? 'Criando...' : 'Criar Usuário'}
          </button>
        </div>
      )}

      {/* Users table */}
      <div id="users-table" className="bg-white border border-earth-200 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-earth-50 text-olive-800">
              <th className="text-left px-4 py-2 font-medium">Nome</th>
              <th className="text-left px-4 py-2 font-medium">Email</th>
              <th className="text-left px-4 py-2 font-medium">Perfil</th>
              <th className="text-left px-4 py-2 font-medium">Ativo</th>
              <th className="text-right px-4 py-2 font-medium">Ações</th>
            </tr>
          </thead>
          <tbody>
            {[...activeUsers, ...inactiveUsers].map(user => {
              const isEditing = editingId === user.id;
              return (
                <tr key={user.id} className={`border-t border-earth-100 ${!user.active ? 'opacity-50 bg-earth-50' : ''}`}>
                  <td className="px-4 py-2">
                    {isEditing ? (
                      <input value={editName} onChange={e => setEditName(e.target.value)} className="text-sm border border-earth-300 rounded px-2 py-1 w-full" />
                    ) : user.name}
                  </td>
                  <td className="px-4 py-2 text-olive-600">{user.email}</td>
                  <td className="px-4 py-2">
                    {isEditing ? (
                      <div id={`user-edit-roles-${user.id}`} className="min-w-[220px]">
                        <RolesCheckboxes value={editRoles} onChange={setEditRoles} />
                      </div>
                    ) : (
                      <RoleBadges roles={user.roles || []} />
                    )}
                  </td>
                  <td className="px-4 py-2">
                    <span className={`inline-block w-2 h-2 rounded-full ${user.active ? 'bg-green-500' : 'bg-red-400'}`} />
                  </td>
                  <td className="px-4 py-2 text-right space-x-2">
                    {isEditing ? (
                      <>
                        <input
                          value={editPassword}
                          onChange={e => setEditPassword(e.target.value)}
                          placeholder="Nova senha (opcional)"
                          type="password"
                          className="text-xs border border-earth-300 rounded px-2 py-1 w-36"
                        />
                        <button onClick={() => handleSave(user.id)} disabled={saving} className="text-xs text-green-700 hover:underline">
                          {saving ? '...' : 'Salvar'}
                        </button>
                        <button onClick={cancelEdit} className="text-xs text-olive-500 hover:underline">Cancelar</button>
                      </>
                    ) : (
                      <>
                        <button onClick={() => startEdit(user)} className="text-xs text-olive-600 hover:underline">Editar</button>
                        {user.active && (
                          <button onClick={() => handleDelete(user)} className="text-xs text-red-600 hover:underline">Desativar</button>
                        )}
                      </>
                    )}
                  </td>
                </tr>
              );
            })}
            {users.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-4 text-center text-olive-500">Nenhum usuário encontrado.</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ─── Backup Panel ──────────────────────────────────────────────

interface BackupFile {
  name: string;
  size: number;
  modified_at: string;
  last_loaded?: boolean;
}

interface BackupHistoryEntry {
  name: string;
  action: 'create' | 'upload' | 'restore';
  at: string;
}

interface BackupMeta {
  last_loaded: string;
  loaded_at?: string;
  history: BackupHistoryEntry[];
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function BackupPanel() {
  const [backups, setBackups] = useState<BackupFile[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [creating, setCreating] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [restoring, setRestoring] = useState<string | null>(null);
  const [selected, setSelected] = useState<string[]>([]);
  const [history, setHistory] = useState<BackupHistoryEntry[]>([]);
  const [showHistory, setShowHistory] = useState(false);
  const [restoreName, setRestoreName] = useState<string | null>(null);

  const fetchBackups = useCallback(async () => {
    try {
      const res = await fetchWithAuth('/api/backup');
      if (!res.ok) throw new Error(`Erro ${res.status}`);
      const data = await res.json();
      setBackups(data || []);
      setError('');
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchMeta = useCallback(async () => {
    try {
      const res = await fetchWithAuth('/api/backup/meta');
      if (!res.ok) return;
      const meta: BackupMeta = await res.json();
      setHistory(meta.history || []);
    } catch { /* non-fatal */ }
  }, []);

  const refresh = useCallback(() => {
    fetchBackups();
    fetchMeta();
  }, [fetchBackups, fetchMeta]);

  useEffect(() => {
    fetchBackups();
    fetchMeta();
  }, [fetchBackups, fetchMeta]);

  const handleCreate = async () => {
    setCreating(true);
    try {
      const res = await fetchWithAuth('/api/backup', { method: 'POST' });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || `Erro ${res.status}`);
      }
      refresh();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Excluir backup "${name}"?`)) return;
    try {
      const res = await fetchWithAuth(`/api/backup/${name}`, { method: 'DELETE' });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || `Erro ${res.status}`);
      }
      setSelected(prev => prev.filter(n => n !== name));
      refresh();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const saveBlob = (blob: Blob, filename: string) => {
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = filename;
    a.click();
    URL.revokeObjectURL(a.href);
  };

  const handleDownload = async (name: string) => {
    try {
      const res = await fetchWithAuth(`/api/backup/${name}`);
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || `Erro ${res.status}`);
      }
      const blob = await res.blob();
      saveBlob(blob, name);
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleDownloadSelected = async () => {
    if (selected.length === 0) return;
    setDownloading(true);
    try {
      const res = await fetchWithAuth('/api/backup/download', {
        method: 'POST',
        body: JSON.stringify({ names: selected }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || `Erro ${res.status}`);
      }
      const blob = await res.blob();
      const stamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
      saveBlob(blob, `divoene-backups-${stamp}.zip`);
    } catch (e: any) {
      alert(e.message);
    } finally {
      setDownloading(false);
    }
  };

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length === 0) return;
    setUploading(true);
    try {
      const formData = new FormData();
      files.forEach(f => formData.append('files', f));
      const res = await fetchWithAuth('/api/backup/upload', {
        method: 'POST',
        body: formData,
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || `Erro ${res.status}`);
      }
      refresh();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setUploading(false);
      e.target.value = '';
    }
  };

  const handleRestore = async (name: string) => {
    setRestoring(name);
    try {
      const res = await fetchWithAuth(`/api/backup/${name}/restore`, { method: 'POST' });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || `Erro ${res.status}`);
      }
      const data = await res.json();
      alert(data.message || 'Carregado. O servidor está reiniciando.');
      refresh();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setRestoring(null);
      setRestoreName(null);
    }
  };

  const toggleSelect = (name: string) => {
    setSelected(prev => prev.includes(name) ? prev.filter(n => n !== name) : [...prev, name]);
  };

  const toggleSelectAll = () => {
    setSelected(prev => (prev.length === backups.length ? [] : backups.map(b => b.name)));
  };

  const actionLabel: Record<string, string> = {
    create: 'Criação',
    upload: 'Upload',
    restore: 'Carga (snapshot)',
  };

  if (loading) return <p className="text-olive-600 text-sm">Carregando...</p>;
  if (error) return <p className="text-red-600 text-sm">Erro: {error}</p>;

  const allSelected = backups.length > 0 && selected.length === backups.length;

  return (
    <div id="backup-panel" className="space-y-6">
      <div id="backup-toolbar" className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 id="backup-heading" className="text-sm font-semibold text-olive-800 uppercase tracking-wide">Snapshots do Banco de Dados</h2>
          <p className="text-xs text-olive-500 mt-1">Backups locais (VACUUM INTO). Envie arquivos ou um .zip, baixe vários de uma vez ou carregue um como snapshot.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <label className={`text-sm bg-terracotta-600 text-white px-3 py-2 rounded-lg hover:bg-terracotta-700 cursor-pointer flex items-center gap-1.5 ${uploading ? 'opacity-50 pointer-events-none' : ''}`}>
            <UploadSimple size={16} weight="bold" />
            {uploading ? 'Enviando...' : 'Enviar Backup'}
            <input id="backup-upload-input" type="file" accept=".zip,.sqlite3,.db" multiple className="hidden" onChange={handleUpload} disabled={uploading} />
          </label>
          <button
            id="backup-download-selected"
            onClick={handleDownloadSelected}
            disabled={downloading || selected.length === 0}
            className="text-sm bg-olive-100 text-olive-800 px-3 py-2 rounded-lg hover:bg-olive-200 disabled:opacity-40 border border-earth-300 flex items-center gap-1.5"
          >
            <DownloadSimple size={16} weight="bold" />
            {downloading ? 'Baixando...' : `Baixar selecionados (${selected.length})`}
          </button>
          <button
            id="backup-create"
            onClick={handleCreate}
            disabled={creating}
            className="text-sm bg-olive-700 text-white px-3 py-2 rounded-lg hover:bg-olive-800 disabled:opacity-50 flex items-center gap-1.5"
          >
            <Plus size={16} weight="bold" />
            {creating ? 'Criando...' : 'Novo Backup'}
          </button>
        </div>
      </div>

      <div id="backup-table" className="bg-white border border-earth-200 rounded-lg overflow-hidden">
        {backups.length === 0 ? (
          <p className="px-4 py-6 text-center text-olive-500 text-sm">Nenhum backup disponível.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-earth-50 text-olive-800 sticky top-0">
                  <th className="px-4 py-2 w-10">
                    <input
                      type="checkbox"
                      id="backup-select-all"
                      checked={allSelected}
                      onChange={toggleSelectAll}
                      className="accent-olive-600"
                    />
                  </th>
                  <th className="text-left px-4 py-2 font-medium">Arquivo</th>
                  <th className="text-left px-4 py-2 font-medium">Tamanho</th>
                  <th className="text-left px-4 py-2 font-medium">Data</th>
                  <th className="text-left px-4 py-2 font-medium">Snapshot</th>
                  <th className="text-right px-4 py-2 font-medium">Ações</th>
                </tr>
              </thead>
              <tbody>
                {backups.map(f => (
                  <tr key={f.name} className={`border-t border-earth-100 ${f.last_loaded ? 'bg-olive-50' : ''}`}>
                    <td className="px-4 py-2">
                      <input
                        type="checkbox"
                        checked={selected.includes(f.name)}
                        onChange={() => toggleSelect(f.name)}
                        className="accent-olive-600"
                      />
                    </td>
                    <td className="px-4 py-2 font-mono text-xs text-olive-900 max-w-[280px] truncate" title={f.name}>{f.name}</td>
                    <td className="px-4 py-2 text-olive-600 whitespace-nowrap">{formatBytes(f.size)}</td>
                    <td className="px-4 py-2 text-olive-600 whitespace-nowrap">{new Date(f.modified_at).toLocaleString('pt-BR')}</td>
                    <td className="px-4 py-2">
                      {f.last_loaded && (
                        <span className="text-xs px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 font-medium">
                          Último carregado
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-2 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={() => handleDownload(f.name)}
                          title="Download"
                          aria-label={`Baixar ${f.name}`}
                          className="p-2 rounded-lg text-olive-600 hover:bg-olive-100 hover:text-olive-900 transition-colors min-h-[36px] min-w-[36px]"
                        >
                          <DownloadSimple size={18} />
                        </button>
                        <button
                          onClick={() => setRestoreName(f.name)}
                          disabled={restoring === f.name}
                          title="Carregar como snapshot"
                          aria-label={`Carregar ${f.name} como snapshot`}
                          className="p-2 rounded-lg text-amber-700 hover:bg-amber-100 transition-colors min-h-[36px] min-w-[36px] disabled:opacity-50"
                        >
                          <ArrowClockwise size={18} weight="bold" className={restoring === f.name ? 'animate-spin' : ''} />
                        </button>
                        <button
                          onClick={() => handleDelete(f.name)}
                          title="Excluir"
                          aria-label={`Excluir ${f.name}`}
                          className="p-2 rounded-lg text-red-600 hover:bg-red-50 transition-colors min-h-[36px] min-w-[36px]"
                        >
                          <Trash size={18} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div id="backup-history" className="border border-earth-200 rounded-lg bg-white">
        <button
          id="backup-history-toggle"
          onClick={() => setShowHistory(!showHistory)}
          className="w-full text-left px-4 py-3 text-sm font-medium text-olive-800 flex items-center justify-between"
          aria-expanded={showHistory}
        >
          <span>Histórico de backups e cargas</span>
          <CaretDown size={16} weight="bold" className={`text-olive-400 transition-transform ${showHistory ? 'rotate-180' : ''}`} />
        </button>
        {showHistory && (
          <div className="px-4 pb-3">
            {history.length === 0 ? (
              <p className="text-xs text-olive-500 py-2">Nenhum registro.</p>
            ) : (
              <ul className="divide-y divide-earth-100">
                {[...history].reverse().slice(0, 50).map((h, i) => (
                  <li key={i} className="py-1.5 flex items-center justify-between gap-3 text-xs">
                    <span className="font-mono text-olive-700 truncate">{h.name}</span>
                    <span className="flex items-center gap-2 shrink-0">
                      <span className={`px-2 py-0.5 rounded-full ${
                        h.action === 'restore' ? 'bg-amber-100 text-amber-800' :
                        h.action === 'upload' ? 'bg-blue-100 text-blue-700' :
                        'bg-earth-100 text-earth-700'
                      }`}>{actionLabel[h.action] || h.action}</span>
                      <span className="text-olive-400 whitespace-nowrap">{new Date(h.at).toLocaleString('pt-BR')}</span>
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </div>

      {/* Restore confirmation modal */}
      {restoreName && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4" onClick={() => setRestoreName(null)}>
          <div className="bg-white rounded-2xl shadow-2xl w-full max-w-md p-6" onClick={e => e.stopPropagation()} role="dialog" aria-modal="true">
            <div className="flex items-start justify-between gap-3 mb-3">
              <h3 className="font-semibold text-olive-900">Carregar snapshot</h3>
              <button onClick={() => setRestoreName(null)} aria-label="Fechar" className="text-olive-400 hover:text-olive-700 p-1">
                <X size={18} weight="bold" />
              </button>
            </div>
            <p className="text-sm text-olive-700 mb-1">
              Carregar <span className="font-mono text-olive-900 break-all">{restoreName}</span> como snapshot?
            </p>
            <p className="text-xs text-olive-500 mb-5">
              O banco de dados atual será substituído, este backup será marcado como o último carregado e o servidor reiniciará automaticamente.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setRestoreName(null)}
                className="px-4 py-2 rounded-lg text-sm text-olive-500 hover:bg-earth-100"
              >
                Cancelar
              </button>
              <button
                id="backup-restore-confirm"
                onClick={() => handleRestore(restoreName)}
                disabled={restoring === restoreName}
                className="px-4 py-2 rounded-lg text-sm bg-amber-600 text-white hover:bg-amber-700 disabled:opacity-50"
              >
                {restoring === restoreName ? 'Carregando...' : 'Carregar snapshot'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
