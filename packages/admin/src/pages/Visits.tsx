import { usePipeline } from '../hooks/usePipeline';
import { useAuthContext, fetchWithAuth } from '../hooks/useAuth';
import { useState } from 'react';
import WhatsAppChat from '../components/WhatsAppChat';

const FEEDBACK_LABELS: Record<string, string> = { liked: 'Gostou', disliked: 'Não gostou', maybe: 'Talvez' };
const FEEDBACK_COLORS: Record<string, string> = {
  liked: 'bg-olive-100 text-olive-700',
  disliked: 'bg-red-50 text-red-600',
  maybe: 'bg-amber-100 text-amber-800',
};

export default function Visits() {
  const { user } = useAuthContext();
  const { allLeads, confirmVisitLocal, setVisitFeedback } = usePipeline();
  const [feedback, setFeedback] = useState<Record<string, { result: string; notes: string }>>({});

  const visits = allLeads.filter(l =>
    l.visit && l.visit.guideId === user?.uid && (l.stage === 'visit_scheduled' || l.stage === 'visit_done')
  );

  const today = new Date().toISOString().split('T')[0];
  const todayVisits = visits.filter(v => v.visit?.date === today);
  const upcomingVisits = visits.filter(v => v.visit?.date !== today);

  async function confirmVisit(leadId: string): Promise<boolean> {
    const lead = allLeads.find(l => l.id === leadId);
    const visitId = lead?.visit?.visitId;
    if (!visitId) return false;
    try {
      const res = await fetchWithAuth(`/api/visits/${visitId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: 'confirmed' }),
      });
      if (res.ok) { confirmVisitLocal(leadId); return true; }
    } catch { /* best-effort */ }
    return false;
  }

  async function submitFeedback(leadId: string) {
    const fb = feedback[leadId];
    if (!fb?.result) return;
    const lead = allLeads.find(l => l.id === leadId);
    const visitId = lead?.visit?.visitId;
    if (!visitId) return;
    try {
      const res = await fetchWithAuth(`/api/visits/${visitId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: 'done', feedback: { result: fb.result, notes: fb.notes || '' } }),
      });
      if (res.ok) {
        setVisitFeedback(leadId, fb.result, fb.notes || '');
        setFeedback(prev => ({ ...prev, [leadId]: { result: '', notes: '' } }));
      }
    } catch { /* best-effort */ }
  }

  return (
    <div id="page-visits" className="max-w-6xl mx-auto p-4 md:p-6">
      <h1 id="visits-heading" className="font-serif text-2xl italic text-olive-900 mb-6">Minhas Visitas</h1>

      {visits.length === 0 && <p id="visits-empty" className="text-olive-500">Nenhuma visita agendada.</p>}

      {todayVisits.length > 0 && (
        <div className="mb-8">
          <h2 className="font-semibold text-olive-900 mb-3">Hoje</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {todayVisits.map(lead => (
              <VisitCard key={lead.id} lead={lead} onConfirm={() => confirmVisit(lead.id)} feedback={feedback[lead.id]} setFeedback={f => setFeedback(prev => ({ ...prev, [lead.id]: f }))} onSubmitFeedback={() => submitFeedback(lead.id)} />
            ))}
          </div>
        </div>
      )}

      {upcomingVisits.length > 0 && (
        <div>
          <h2 className="font-semibold text-olive-900 mb-3">Próximas</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {upcomingVisits.map(lead => (
              <VisitCard key={lead.id} lead={lead} onConfirm={() => confirmVisit(lead.id)} feedback={feedback[lead.id]} setFeedback={f => setFeedback(prev => ({ ...prev, [lead.id]: f }))} onSubmitFeedback={() => submitFeedback(lead.id)} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function VisitCard({ lead, feedback, setFeedback, onConfirm, onSubmitFeedback }: {
  lead: any; onConfirm: () => Promise<boolean>; feedback?: { result: string; notes: string };
  setFeedback: (f: { result: string; notes: string }) => void; onSubmitFeedback: () => void;
}) {
  const [showConfirmChat, setShowConfirmChat] = useState(false);
  const showFeedback = lead.stage === 'visit_scheduled';

  async function handleConfirm() {
    const ok = await onConfirm();
    if (ok) setShowConfirmChat(true);
  }

  return (
    <div id={`visit-card-${lead.id}`} className="bg-white rounded-xl p-4 border border-earth-200 mb-3">
      <p className="font-semibold text-olive-900">{lead.name}</p>
      <p className="text-sm text-olive-700">{lead.product} · {lead.whatsapp}</p>
      <p className="text-sm text-olive-700">{lead.visit?.date} às {lead.visit?.timeSlot}</p>

      {lead.stage === 'visit_scheduled' && !lead.visit?.confirmed && (
        <button id={`visit-confirm-${lead.id}`} onClick={handleConfirm} className="mt-3 bg-olive-500 text-white px-4 py-2 rounded-lg text-sm hover:bg-olive-700 min-h-[40px]">Confirmar presença</button>
      )}

      {lead.visit?.confirmed && showFeedback && (
        <div className="mt-3 space-y-2">
          <p className="text-sm text-olive-500">Registrar resultado:</p>
          <div className="flex gap-2">
            {['liked', 'disliked', 'maybe'].map(r => (
              <button key={r} id={`visit-feedback-${r}-${lead.id}`} onClick={() => setFeedback({ ...(feedback || { notes: '' }), result: r })}
                className={`px-3 py-1 rounded-full text-xs min-h-[32px] ${feedback?.result === r ? 'bg-olive-500 text-white' : 'bg-earth-100 text-olive-700'}`}>
                {r === 'liked' ? 'Gostou' : r === 'disliked' ? 'Não gostou' : 'Talvez'}
              </button>
            ))}
          </div>
          <input
            id={`visit-feedback-input-${lead.id}`}
            placeholder="Notas (opcional)"
            value={feedback?.notes || ''}
            onChange={e => setFeedback({ ...(feedback || { result: '' }), notes: e.target.value })}
            className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm"
          />
          <button id={`visit-feedback-submit-${lead.id}`} onClick={onSubmitFeedback} className="bg-olive-500 text-white px-4 py-2 rounded-lg text-sm hover:bg-olive-700 min-h-[40px]">Enviar feedback</button>
        </div>
      )}

      {lead.visit?.feedback && (
        <div className="mt-3 text-sm text-olive-700 bg-earth-100 rounded-lg p-3">
          <p className="flex items-center gap-2">
            <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${FEEDBACK_COLORS[lead.visit.feedback.result] || 'bg-earth-200 text-earth-700'}`}>
              {FEEDBACK_LABELS[lead.visit.feedback.result] || lead.visit.feedback.result}
            </span>
            <span>Resultado da visita</span>
          </p>
          {lead.visit.feedback.notes && <p className="mt-1.5">{lead.visit.feedback.notes}</p>}
        </div>
      )}

      {showConfirmChat && (
        <div id={`visit-chat-${lead.id}`} className="mt-3 border-t border-earth-100 pt-3">
          <WhatsAppChat
            leadId={lead.id}
            leadName={lead.name}
            leadWhatsApp={lead.whatsapp}
            product={lead.product}
            stage={lead.stage}
            collapsible={false}
          />
        </div>
      )}
    </div>
  );
}
