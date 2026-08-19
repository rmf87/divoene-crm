// TerminalProcess — Stages "completed" and "cancelled": read-only summary.
import { CheckCircle, Medal, XCircle } from '@phosphor-icons/react';
import { stageLabels, type Lead } from '../../hooks/usePipeline';

interface Props {
  lead: Lead;
}

export default function TerminalProcess({ lead }: Props) {
  const isCompleted = lead.stage === 'completed';
  const isCancelled = lead.stage === 'cancelled';

  return (
    <div id="process-terminal" className="flex-1 min-w-0 bg-white rounded-2xl border border-earth-200 overflow-hidden">
      <div className="px-5 py-4 border-b border-earth-100 flex items-center gap-3">
        <span className={`px-2.5 py-1 rounded-full text-xs font-semibold ${
          isCompleted ? 'bg-olive-100 text-olive-700' : 'bg-earth-100 text-olive-400'
        }`}>
          {stageLabels[lead.stage]}
        </span>
        <h2 className="font-serif text-lg italic text-olive-900">
          {isCompleted ? 'Lead Concluído' : 'Lead Cancelado'}
        </h2>
      </div>
      <div className="p-5 space-y-3">
        <div className={`rounded-lg p-4 text-sm ${
          isCompleted ? 'bg-olive-50 text-olive-700' : 'bg-red-50 text-terracotta-600'
        }`}>
          <p className="font-semibold text-lg mb-1 flex items-center gap-2">
            {isCompleted
              ? <><Medal size={22} weight="fill" className="text-olive-600" /> Evento realizado com sucesso</>
              : <><XCircle size={22} weight="fill" /> Lead cancelado</>}
          </p>
          <p>{isCompleted ? 'Este lead foi concluído. Nenhuma ação pendente.' : 'Este lead foi cancelado e não requer mais ações.'}</p>
        </div>
        {lead.stageHistory && lead.stageHistory.length > 0 && (
          <div className="space-y-1">
            <p className="text-xs font-semibold text-olive-500 uppercase">Histórico de estágios</p>
            {lead.stageHistory.map((h, i) => (
              <div key={i} className="text-xs text-olive-700 flex justify-between">
                <span>{stageLabels[h.stage] || h.stage}</span>
                <span className="text-olive-400">{new Date(h.changedAt).toLocaleDateString('pt-BR')}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
