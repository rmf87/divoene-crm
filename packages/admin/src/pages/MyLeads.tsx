import { useState } from 'react';
import { usePipeline } from '../hooks/usePipeline';
import { useNavigate } from 'react-router-dom';
import { useAuthContext } from '../hooks/useAuth';
import UndoRedoControls from '../components/UndoRedoControls';

const PRODUCTS = [
  { id: 'ensaio_fotografico', name: 'Ensaio Fotográfico' },
  { id: 'locacao_eventos', name: 'Eventos' },
  { id: 'corporativo', name: 'Corporativo' },
  { id: 'casamentos', name: 'Casamentos' },
  { id: 'buffet_infantil', name: 'Buffet Infantil' },
  { id: 'passeios_escolares', name: 'Passeios Escolares' },
];

const SOURCES = [
  'instagram', 'facebook', 'google', 'linkedin', 'indicacao', 'evento', 'visita', 'outro',
];

const stageLabels: Record<string, string> = {
  lead: 'Lead', validated: 'Validado', visit_scheduled: 'Visita Agendada', visit_done: 'Visita Feita',
  contract: 'Contrato', paid: 'Pago', booked: 'Reservado', completed: 'Concluído', cancelled: 'Cancelado',
};

export default function MyLeads() {
  const { user } = useAuthContext();
  const { leads: myLeads, allLeads, createLead, undo, redo, canUndo, canRedo } = usePipeline();
  const navigate = useNavigate();
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: '', whatsapp: '', product: '', desiredDate: '', source: 'indicacao' });
  const [error, setError] = useState('');
  const [submitted, setSubmitted] = useState(false);

  const totalCommission = allLeads
    .filter(l => l.commission?.status === 'pending_payout' && l.createdBy === user?.uid)
    .reduce((sum, l) => sum + (l.commission?.amount || 0), 0);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    if (!form.name.trim()) return setError('Nome é obrigatório.');
    if (!form.whatsapp.trim() || form.whatsapp.replace(/\D/g, '').length < 10) return setError('WhatsApp inválido.');
    if (!form.product) return setError('Selecione um produto.');

    const cleanWhatsapp = form.whatsapp.replace(/\D/g, '');
    await createLead({
      name: form.name.trim(),
      whatsapp: cleanWhatsapp,
      product: form.product,
      desiredDate: form.desiredDate || '',
      source: form.source,
      stage: 'lead',
      assignedSeller: '',
    });

    setSubmitted(true);
    setShowForm(false);
    setForm({ name: '', whatsapp: '', product: '', desiredDate: '', source: 'indicacao' });
    setTimeout(() => setSubmitted(false), 4000);
  }

  function formatWhatsApp(value: string): string {
    const digits = value.replace(/\D/g, '');
    if (digits.length <= 2) return digits;
    if (digits.length <= 7) return `(${digits.slice(0, 2)}) ${digits.slice(2)}`;
    return `(${digits.slice(0, 2)}) ${digits.slice(2, 7)}-${digits.slice(7, 11)}`;
  }

  return (
    <div id="page-my-leads" className="max-w-6xl mx-auto p-4 md:p-6">
      <div className="flex justify-between items-start mb-6">
        <div>
          <h1 id="myleads-heading" className="font-serif text-2xl italic text-olive-900">Meus Leads</h1>
          <p id="myleads-commission" className="text-olive-700">Comissão a receber: <span className="font-bold text-terracotta-500">R$ {(totalCommission / 100).toFixed(2)}</span></p>
          <UndoRedoControls undo={undo} redo={redo} canUndo={canUndo} canRedo={canRedo} />
        </div>
        <button
          id="myleads-new-lead"
          onClick={() => setShowForm(!showForm)}
          className="bg-olive-500 text-white px-4 py-2 rounded-lg text-sm font-semibold hover:bg-olive-700 transition-colors"
        >
          {showForm ? 'Cancelar' : '+ Novo Lead'}
        </button>
      </div>

      {submitted && (
        <div id="myleads-submitted" className="bg-green-100 text-green-800 rounded-xl p-4 mb-4 text-sm">
          Lead cadastrado com sucesso! Um vendedor vai entrar em contato.
        </div>
      )}

      {showForm && (
        <form id="myleads-form" onSubmit={handleSubmit} className="bg-white rounded-2xl p-6 border border-earth-200 mb-6 space-y-4">
          <h2 className="font-semibold text-olive-900">Novo Lead</h2>

          <input id="myleads-form-nome" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="Nome do cliente *" className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm" />

          <input id="myleads-form-whatsapp" value={formatWhatsApp(form.whatsapp)} onChange={e => setForm({ ...form, whatsapp: e.target.value })} placeholder="WhatsApp (DDD + número) *" className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm" />

          <select id="myleads-form-produto" value={form.product} onChange={e => setForm({ ...form, product: e.target.value })} className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm">
            <option value="">Produto de interesse *</option>
            {PRODUCTS.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
          </select>

          <input id="myleads-form-data" type="date" value={form.desiredDate} onChange={e => setForm({ ...form, desiredDate: e.target.value })} className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm" />

          <select id="myleads-form-origem" value={form.source} onChange={e => setForm({ ...form, source: e.target.value })} className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm">
            <option value="">Origem</option>
            {SOURCES.map(s => <option key={s} value={s}>{s}</option>)}
          </select>

          {error && <p id="myleads-form-error" className="text-red-600 text-sm">{error}</p>}

          <button id="myleads-form-submit" type="submit" className="w-full bg-olive-500 text-white py-2.5 rounded-lg text-sm font-semibold hover:bg-olive-700">Cadastrar Lead</button>
        </form>
      )}

      {myLeads.length === 0 && !showForm && <p className="text-olive-500">Nenhum lead cadastrado.</p>}

      <div id="myleads-list" className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {myLeads.map(lead => (
          <div
            key={lead.id}
            id={`mylead-${lead.id}`}
            onClick={() => navigate(`/leads/${lead.id}`)}
            className="bg-white rounded-xl p-4 border border-earth-200 cursor-pointer hover:shadow-md transition-shadow"
          >
            <div className="flex justify-between items-start">
              <div>
                <p className="font-semibold text-olive-900">{lead.name}</p>
                <p className="text-sm text-olive-700">{lead.product} · {lead.source}</p>
              </div>
              <span className="text-xs text-olive-500 bg-earth-100 px-2 py-1 rounded-full">{stageLabels[lead.stage]}</span>
            </div>
            {lead.commission && (
              <p className="text-xs text-terracotta-500 mt-2">Comissão: R$ {(lead.commission.amount / 100).toFixed(2)} — {lead.commission.status}</p>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
