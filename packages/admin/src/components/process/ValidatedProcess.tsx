// ValidatedProcess — Stage "validated": visit scheduling.
// Uses VisitScheduler modal internally.
import { useState } from 'react';
import { CalendarCheck, XCircle } from '@phosphor-icons/react';
import { stageLabels, type Lead } from '../../hooks/usePipeline';
import VisitScheduler, { type VisitData } from '../VisitScheduler';

interface Props {
  lead: Lead;
  dispatch: (action: any) => void;
}

export default function ValidatedProcess({ lead, dispatch }: Props) {
  const [showScheduler, setShowScheduler] = useState(false);

  function handleBook(leadId: string, visit: VisitData) {
    dispatch({ type: 'BOOK_VISIT', leadId, visit });
    setShowScheduler(false);
  }

  return (
    <div id="process-validated" className="flex-1 min-w-0 bg-white rounded-2xl border border-earth-200 overflow-hidden">
      <div className="px-5 py-4 border-b border-earth-100 flex items-center gap-3">
        <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-olive-500 text-white">{stageLabels.validated}</span>
        <h2 className="font-serif text-lg italic text-olive-900">Agendar Visita Técnica</h2>
      </div>
      <div className="p-5 space-y-3">
        <p className="text-sm text-olive-700">Lead validado. Agende a visita técnica para o cliente conhecer o espaço.</p>
        {lead.event && (
          <div className="bg-olive-50 rounded-lg p-3 text-sm text-olive-700 space-y-1">
            <p><strong>Evento:</strong> {lead.event.eventType} · {lead.event.desiredDurationHours}h · ~{lead.event.estimatedPeople} pessoas</p>
            {lead.event.possibleDates.length > 0 && <p><strong>Datas:</strong> {lead.event.possibleDates.join(', ')}</p>}
          </div>
        )}
        <button id="process-action-schedule" onClick={() => setShowScheduler(true)}
          className="w-full text-left bg-olive-500 text-white px-5 py-3 rounded-xl text-sm font-semibold hover:bg-olive-700 transition-colors min-h-[44px] flex items-center gap-2">
          <CalendarCheck size={18} weight="fill" />
          Agendar Visita
        </button>
        <button id="process-action-cancel" onClick={() => dispatch({ type: 'MOVE_LEAD', leadId: lead.id, toStage: 'cancelled' })}
          className="w-full text-left bg-earth-100 text-terracotta-600 px-5 py-3 rounded-xl text-sm hover:bg-earth-200 transition-colors min-h-[44px] flex items-center gap-2">
          <XCircle size={16} weight="bold" />
          Cancelar Lead
        </button>
      </div>
      {showScheduler && <VisitScheduler lead={lead} onBook={handleBook} onClose={() => setShowScheduler(false)} />}
    </div>
  );
}
