// BookedProcess — Stage "booked": finalize the process.
import { CheckCircle, XCircle } from '@phosphor-icons/react';
import { stageLabels, type Lead } from '../../hooks/usePipeline';

interface Props {
  lead: Lead;
  dispatch: (action: any) => void;
}

export default function BookedProcess({ lead, dispatch }: Props) {
  return (
    <div id="process-booked" className="flex-1 min-w-0 bg-white rounded-2xl border border-earth-200 overflow-hidden">
      <div className="px-5 py-4 border-b border-earth-100 flex items-center gap-3">
        <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-olive-500 text-white">{stageLabels.booked}</span>
        <h2 className="font-serif text-lg italic text-olive-900">Finalizar Processo</h2>
      </div>
      <div className="p-5 space-y-3">
        <p className="text-sm text-olive-700">Reserva confirmada! Após o evento, marque como concluído.</p>
        <button id="process-action-complete" onClick={() => dispatch({ type: 'MOVE_LEAD', leadId: lead.id, toStage: 'completed' })}
          className="w-full text-left bg-olive-500 text-white px-5 py-3 rounded-xl text-sm font-semibold hover:bg-olive-700 min-h-[44px] flex items-center gap-2">
          <CheckCircle size={18} weight="fill" />
          Concluir Evento
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
