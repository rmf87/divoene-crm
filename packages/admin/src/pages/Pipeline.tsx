import { useState } from 'react';
import { CaretDown, CheckCircle, Warning, X, XCircle } from '@phosphor-icons/react';
import { STAGES } from '../hooks/usePipeline';
import { usePipeline } from '../hooks/usePipeline';
import StageColumn from '../components/StageColumn';
import UndoRedoControls from '../components/UndoRedoControls';

const ACTIVE_STAGES = STAGES.filter(s => s !== 'completed' && s !== 'cancelled');

export default function Pipeline() {
  const { leads, allLeads, moveLead, apiError, toast, dismissToast, undo, redo, canUndo, canRedo } = usePipeline();
  const [showTerminal, setShowTerminal] = useState(false);

  function handleDrop(leadId: string, toStage: string) { moveLead(leadId, toStage); }

  const grouped: Record<string, typeof leads> = {};
  STAGES.forEach(s => { grouped[s] = []; });
  leads.forEach(l => { if (grouped[l.stage]) grouped[l.stage].push(l); });

  const terminalLeads = [...(grouped['completed'] || []), ...(grouped['cancelled'] || [])];
  const hasTerminal = terminalLeads.length > 0;
  const completedCount = (grouped['completed'] || []).length;
  const cancelledCount = (grouped['cancelled'] || []).length;

  return (
    <div id="page-pipeline" className="h-screen flex flex-col">
      <div className="px-4 md:px-6 py-3 flex items-center justify-between gap-3 border-b border-earth-200 bg-white/80 backdrop-blur shrink-0">
        <h1 id="pipeline-heading" className="font-serif text-xl italic text-olive-900">Pipeline</h1>
        <UndoRedoControls undo={undo} redo={redo} canUndo={canUndo} canRedo={canRedo} />
      </div>

      {apiError && (
        <div id="pipeline-api-error" className="mx-4 md:mx-6 mt-3 px-3 py-2 bg-earth-100 border border-terracotta-400 text-terracotta-600 rounded-lg text-sm flex items-center gap-2 shrink-0">
          <Warning weight="fill" className="shrink-0" size={18} />
          <span>API de leads indisponível. Exibindo apenas dados mock.</span>
        </div>
      )}

      {toast && (
        <div className="fixed bottom-4 right-4 md:top-4 md:bottom-auto z-50 bg-olive-900 text-white px-4 py-3 rounded-xl shadow-lg text-sm flex gap-3 items-center">
          <span>{toast}</span>
          <button onClick={dismissToast} className="text-white/70 hover:text-white" aria-label="Fechar">
            <X size={16} weight="bold" />
          </button>
        </div>
      )}

      {/* Kanban board — full height, horizontal scroll */}
      <div id="pipeline-board" className="flex-1 min-h-0 overflow-x-auto px-4 md:px-6 py-4">
        <div className="flex gap-3 h-full min-w-max items-stretch">
          {ACTIVE_STAGES.map(stage => (
            <div key={stage} className="w-72 shrink-0 flex flex-col">
              <StageColumn stage={stage} leads={grouped[stage] || []} onDrop={handleDrop} onPromote={moveLead} />
            </div>
          ))}
        </div>
      </div>

      {/* Terminal stages: compact collapsible strip */}
      {hasTerminal && (
        <div id="pipeline-terminal" className="border-t border-earth-200 bg-white/80 shrink-0">
          <button
            id="pipeline-terminal-toggle"
            onClick={() => setShowTerminal(!showTerminal)}
            className="w-full flex items-center gap-3 px-4 md:px-6 py-2 text-sm text-olive-700 hover:bg-earth-50"
            aria-expanded={showTerminal}
          >
            <span className="font-semibold">Finalizados</span>
            <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-olive-50 text-olive-700">
              <CheckCircle weight="fill" size={14} className="text-olive-600" />
              {completedCount} concluído{completedCount === 1 ? '' : 's'}
            </span>
            <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-red-50 text-red-600">
              <XCircle weight="fill" size={14} />
              {cancelledCount} cancelado{cancelledCount === 1 ? '' : 's'}
            </span>
            <CaretDown
              size={16}
              weight="bold"
              className={`ml-auto text-olive-400 transition-transform ${showTerminal ? 'rotate-180' : ''}`}
            />
          </button>
          {showTerminal && (
            <div className="px-4 md:px-6 pb-3 flex gap-2 overflow-x-auto">
              {terminalLeads.map(lead => {
                const isCompleted = lead.stage === 'completed';
                return (
                  <div
                    key={lead.id}
                    className={`flex items-center gap-2 text-xs px-3 py-1.5 rounded-full border shrink-0 ${
                      isCompleted ? 'border-olive-300 bg-olive-50 text-olive-700' : 'border-terracotta-400 bg-red-50 text-red-600'
                    }`}
                  >
                    {isCompleted
                      ? <CheckCircle weight="fill" size={14} className="text-olive-600" />
                      : <XCircle weight="fill" size={14} />}
                    <span className="font-medium">{lead.name}</span>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
