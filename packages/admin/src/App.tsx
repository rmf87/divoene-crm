import { Routes, Route, Navigate } from 'react-router-dom';
import { getStoredAuth, AuthContext } from './hooks/useAuth';
import { PipelineUndoProvider, initialPipelineState } from './hooks/usePipeline';
import AuthGate from './components/AuthGate';
import RoleGate from './components/RoleGate';
import AdminHeader from './components/AdminHeader';
import Login from './pages/Login';
import Pipeline from './pages/Pipeline';
import LeadDetailPage from './pages/LeadDetailPage';
import MyLeads from './pages/MyLeads';
import Visits from './pages/Visits';
import ManageVisits from './pages/ManageVisits';
import Reports from './pages/Reports';
import Availability from './pages/Availability';
import Config from './pages/Config';
import Help from './pages/Help';

function roleHome(roles?: string[]): string {
  if (!roles) return '/pipeline';
  if (roles.includes('manager') || roles.includes('seller')) return '/pipeline';
  if (roles.includes('guide')) return '/visits';
  if (roles.includes('associate')) return '/my-leads';
  return '/pipeline';
}

function Page({ children }: { children: React.ReactNode }) {
  return (
    <>
      <AdminHeader />
      {children}
    </>
  );
}

export default function App() {
  const auth = getStoredAuth();

  return (
    <div id="app-root">
    <AuthContext.Provider value={auth}>
      <PipelineUndoProvider initialState={initialPipelineState}>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/pipeline" element={<AuthGate><Page><RoleGate roles={['seller', 'manager']} fallback={<AccessDenied />}><Pipeline /></RoleGate></Page></AuthGate>} />
        <Route path="/leads/:id" element={<AuthGate><Page><RoleGate roles={['seller', 'manager']} fallback={<AccessDenied />}><LeadDetailPage /></RoleGate></Page></AuthGate>} />
        <Route path="/my-leads" element={<AuthGate><Page><RoleGate roles={['associate', 'manager']} fallback={<AccessDenied />}><MyLeads /></RoleGate></Page></AuthGate>} />
        <Route path="/visits" element={<AuthGate><Page><RoleGate roles={['guide']} fallback={<AccessDenied />}><Visits /></RoleGate></Page></AuthGate>} />
        <Route path="/manage-visits" element={<AuthGate><Page><RoleGate roles={['manager']} fallback={<AccessDenied />}><ManageVisits /></RoleGate></Page></AuthGate>} />
        <Route path="/reports" element={<AuthGate><Page><RoleGate roles={['manager']} fallback={<AccessDenied />}><Reports /></RoleGate></Page></AuthGate>} />
        <Route path="/availability" element={<AuthGate><Page><RoleGate roles={['guide', 'manager']} fallback={<AccessDenied />}><Availability /></RoleGate></Page></AuthGate>} />
        <Route path="/config" element={<AuthGate><Page><RoleGate roles={['manager']} fallback={<AccessDenied />}><Config /></RoleGate></Page></AuthGate>} />
        <Route path="/ajuda" element={<AuthGate><Page><Help /></Page></AuthGate>} />
        <Route path="*" element={<Navigate to={roleHome(auth.user?.roles)} replace />} />
      </Routes>
      </PipelineUndoProvider>
    </AuthContext.Provider>
    </div>
  );
}

function AccessDenied() {
  return (
    <div className="min-h-[60vh] flex items-center justify-center">
      <div className="text-center">
        <p className="font-serif text-2xl text-olive-900 italic">Acesso negado</p>
        <p className="text-olive-700 mt-2">Seu perfil não tem acesso a esta página.</p>
      </div>
    </div>
  );
}
