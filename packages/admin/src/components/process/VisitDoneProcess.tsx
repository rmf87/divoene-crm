// VisitDoneProcess — Stage "visit_done": send contract to client.
import { PaperPlaneTilt, XCircle } from '@phosphor-icons/react';
import { stageLabels, type Lead } from '../../hooks/usePipeline';

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

export default function VisitDoneProcess({ lead, dispatch }: Props) {
  return (
    <div id="process-visit-done" className="flex-1 min-w-0 bg-white rounded-2xl border border-earth-200 overflow-hidden">
      <div className="px-5 py-4 border-b border-earth-100 flex items-center gap-3">
        <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-olive-500 text-white">{stageLabels.visit_done}</span>
        <h2 className="font-serif text-lg italic text-olive-900">Enviar Contrato</h2>
      </div>
      <div className="p-5 space-y-3">
        <p className="text-sm text-olive-700">Visita concluída. Envie o contrato para o cliente formalizar.</p>
        {lead.visit?.feedback && (
          <div className="bg-olive-50 rounded-lg p-3 text-sm text-olive-700">
            <p className="flex items-center gap-2">
              <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${FEEDBACK_COLORS[lead.visit.feedback.result] || 'bg-earth-200 text-earth-700'}`}>
                {FEEDBACK_LABELS[lead.visit.feedback.result] || lead.visit.feedback.result}
              </span>
              <span>Feedback da visita</span>
            </p>
            {lead.visit.feedback.notes && <p className="text-olive-500 mt-0.5">{lead.visit.feedback.notes}</p>}
          </div>
        )}
        <button id="process-action-contract" onClick={() => dispatch({ type: 'MOVE_LEAD', leadId: lead.id, toStage: 'contract' })}
          className="w-full text-left bg-olive-500 text-white px-5 py-3 rounded-xl text-sm font-semibold hover:bg-olive-700 min-h-[44px] flex items-center gap-2">
          <PaperPlaneTilt size={18} weight="fill" />
          Enviar Contrato
        </button>
        <button id="process-action-cancel" onClick={() => dispatch({ type: 'MOVE_LEAD', leadId: lead.id, toStage: 'cancelled' })}
          className="w-full text-left bg-earth-100 text-terracotta-600 px-5 py-3 rounded-xl text-sm hover:bg-earth-200 transition-colors min-h-[44px] flex items-center gap-2">
          <XCircle size={16} weight="bold" />
          Cancelar Lead
        </button>
      </div>
    </div>
  );
}
