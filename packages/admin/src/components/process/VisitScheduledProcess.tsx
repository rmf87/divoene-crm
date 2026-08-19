// VisitScheduledProcess — Stage "visit_scheduled": confirm with feedback or reject scheduled visit.
import { useState } from 'react';
import { CheckCircle, X, XCircle } from '@phosphor-icons/react';
import { stageLabels, type Lead } from '../../hooks/usePipeline';
import { fetchWithAuth } from '../../hooks/useAuth';

const FEEDBACK_LABELS: Record<string, string> = { liked: 'Gostou', disliked: 'Não gostou', maybe: 'Talvez' };
const FEEDBACK_COLORS: Record<string, string> = {
  liked: 'bg-olive-100 text-olive-700',
  disliked: 'bg-red-50 text-red-600',
  maybe: 'bg-amber-100 text-amber-800',
};

interface Props {
  lead: Lead;
  dispatch: (action: any) => void;
}

export default function VisitScheduledProcess({ lead, dispatch }: Props) {
  const [showConfirmForm, setShowConfirmForm] = useState(false);
  const [feedbackResult, setFeedbackResult] = useState('');
  const [feedbackNotes, setFeedbackNotes] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [confirmError, setConfirmError] = useState('');

  function resetForm() {
    setShowConfirmForm(false);
    setFeedbackResult('');
    setFeedbackNotes('');
    setConfirmError('');
  }

  async function handleConfirm() {
    if (!feedbackResult) return;
    setSubmitting(true);
    setConfirmError('');
    try {
      const visitId = lead.visit?.visitId;
      if (visitId) {
        const res = await fetchWithAuth(`/api/visits/${visitId}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ status: 'done', feedback: { result: feedbackResult, notes: feedbackNotes } }),
        });
        if (!res.ok) {
          const body = await res.json().catch(() => ({}));
          throw new Error(body.error || 'Falha ao registrar feedback');
        }
      }
      dispatch({ type: 'SET_VISIT_FEEDBACK', leadId: lead.id, result: feedbackResult, notes: feedbackNotes });
      dispatch({ type: 'MOVE_LEAD', leadId: lead.id, toStage: 'visit_done' });
    } catch (e: any) {
      setConfirmError(e.message || 'Erro ao confirmar');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div id="process-visit-scheduled" className="flex-1 min-w-0 bg-white rounded-2xl border border-earth-200 overflow-hidden">
      <div className="px-5 py-4 border-b border-earth-100 flex items-center gap-3">
        <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-olive-500 text-white">{stageLabels.visit_scheduled}</span>
        <h2 className="font-serif text-lg italic text-olive-900">Visita Agendada</h2>
      </div>
      <div className="p-5 space-y-3">
        {lead.visit && (
          <div className="bg-olive-50 rounded-lg p-3 text-sm text-olive-700 space-y-1">
            <p><strong>Data:</strong> {lead.visit.date} às {lead.visit.timeSlot}</p>
            <p><strong>Guia:</strong> {lead.visit.guideId === 'dev-guide-uid' ? 'Guia (dev)' : lead.visit.guideId === 'dev-guide-uid-2' ? 'Guia 2 (dev)' : lead.visit.guideId}</p>
            <p><strong>Confirmação:</strong> {lead.visit.confirmed ? 'Confirmado' : 'Pendente'}</p>
            {lead.visit.feedback && (
              <p className="text-xs text-olive-500 mt-1 flex items-center gap-1.5">
                <span className={`px-2 py-0.5 rounded-full font-semibold ${FEEDBACK_COLORS[lead.visit.feedback.result] || 'bg-earth-200 text-earth-700'}`}>
                  {FEEDBACK_LABELS[lead.visit.feedback.result] || lead.visit.feedback.result}
                </span>
                {lead.visit.feedback.notes ? ` — ${lead.visit.feedback.notes}` : ''}
              </p>
            )}
          </div>
        )}

        {!lead.visit?.confirmed && !showConfirmForm && (
          <button id="process-action-confirm" onClick={() => setShowConfirmForm(true)}
            className="w-full text-left bg-olive-500 text-white px-5 py-3 rounded-xl text-sm font-semibold hover:bg-olive-700 min-h-[44px] flex items-center gap-2">
            <CheckCircle size={18} weight="fill" />
            Confirmar Visita Realizada
          </button>
        )}

        {showConfirmForm && (
          <div className="bg-olive-50 rounded-xl p-4 space-y-3">
            <h3 className="text-sm font-semibold text-olive-900">Resultado da Visita</h3>

            {/* Feedback buttons */}
            <div className="flex gap-2">
              {[
                { value: 'liked', label: 'Gostou' },
                { value: 'disliked', label: 'Não gostou' },
                { value: 'maybe', label: 'Talvez' },
              ].map(r => (
                <button key={r.value}
                  id={`visit-confirm-feedback-${r.value}`}
                  onClick={() => setFeedbackResult(r.value)}
                  className={`px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
                    feedbackResult === r.value
                      ? 'bg-olive-500 text-white'
                      : 'bg-white text-olive-700 border border-earth-200 hover:bg-olive-100'
                  }`}>
                  {r.label}
                </button>
              ))}
            </div>

            {/* Notes */}
            <textarea
              id="visit-confirm-notes"
              value={feedbackNotes}
              onChange={e => setFeedbackNotes(e.target.value)}
              placeholder="Notas adicionais (opcional)..."
              className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm resize-none"
              rows={2}
            />

            {confirmError && <p className="text-xs text-terracotta-500">{confirmError}</p>}

            <div className="flex gap-2">
              <button id="visit-confirm-submit" onClick={handleConfirm}
                disabled={!feedbackResult || submitting}
                className="flex-1 bg-olive-500 text-white px-4 py-2 rounded-lg text-sm font-semibold hover:bg-olive-700 disabled:opacity-50 disabled:cursor-not-allowed min-h-[40px]">
                {submitting ? 'Confirmando...' : 'Confirmar'}
              </button>
              <button id="visit-confirm-cancel" onClick={resetForm}
                className="px-4 py-2 rounded-lg text-sm text-olive-500 hover:bg-earth-100 min-h-[40px]">
                Cancelar
              </button>
            </div>
          </div>
        )}

        <button id="process-action-cancel" onClick={() => dispatch({ type: 'MOVE_LEAD', leadId: lead.id, toStage: 'cancelled' })}
          className="w-full text-left bg-earth-100 text-terracotta-600 px-5 py-3 rounded-xl text-sm hover:bg-earth-200 transition-colors min-h-[44px] flex items-center gap-2">
          <X size={16} weight="bold" />
          Cancelar Lead
        </button>
      </div>
    </div>
  );
}
