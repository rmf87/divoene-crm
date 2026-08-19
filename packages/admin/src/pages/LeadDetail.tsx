// Compact reference card — NOT the primary process surface.
// ProcessPanel handles stage actions, checklists, and modals.
import { useState } from 'react';
import { Check, X } from '@phosphor-icons/react';
import { usePipeline, stageLabels } from '../hooks/usePipeline';

interface Props {
  leadId: string;
  onClose?: () => void;
}

export default function LeadDetail({ leadId, onClose }: Props) {
  const { getLeadById, addNote } = usePipeline();
  const [note, setNote] = useState('');
  const [showNotes, setShowNotes] = useState(false);

  const lead = getLeadById(leadId);
  if (!lead) return <div className="p-4 text-olive-500 text-sm">Lead não encontrado.</div>;

  const daysInStage = Math.floor((Date.now() - new Date(lead.lastStageChange).getTime()) / 86400000);

  function handleAddNote() {
    if (note.trim()) { addNote(lead!.id, note.trim()); setNote(''); }
  }

  return (
    <div id="lead-detail-card" className="bg-white rounded-2xl border border-earth-200 w-full max-w-[320px] flex-shrink-0 overflow-y-auto">
      {/* Header with close */}
      <div className="px-4 py-3 border-b border-earth-100 flex items-center justify-between">
        <div className="flex items-center gap-2 min-w-0">
          <span className={`w-2 h-2 rounded-full flex-shrink-0 ${
            lead.stage === 'cancelled' ? 'bg-terracotta-400' :
            lead.stage === 'completed' ? 'bg-olive-500' : 'bg-olive-400'
          }`} />
          <h2 className="font-serif text-base italic text-olive-900 truncate">{lead.name}</h2>
        </div>
        <div className="flex items-center gap-1">
          {onClose && (
            <button onClick={onClose} className="text-olive-400 hover:text-olive-700 p-1" aria-label="Fechar">
              <X size={18} weight="bold" />
            </button>
          )}
        </div>
      </div>

      {/* Key info */}
      <div className="px-4 py-3 space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-olive-500">WhatsApp</span>
          <span className="text-olive-900 font-medium">{lead.whatsapp}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-olive-500">Produto</span>
          <span className="text-olive-900">{lead.product}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-olive-500">Estágio</span>
          <span className="px-2 py-0.5 rounded-full text-xs font-semibold bg-olive-100 text-olive-700">
            {stageLabels[lead.stage]}
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-olive-500">No estágio</span>
          <span className="text-olive-700">{daysInStage}d</span>
        </div>
        {lead.desiredDate && (
          <div className="flex justify-between">
            <span className="text-olive-500">Data</span>
            <span className="text-olive-700">{lead.desiredDate}</span>
          </div>
        )}
        <div className="flex justify-between">
          <span className="text-olive-500">Origem</span>
          <span className="text-olive-700">{lead.source}</span>
        </div>
      </div>

      {/* Event info (if set) */}
      {lead.event && (
        <div className="border-t border-earth-100 px-4 py-3">
          <p className="text-xs font-semibold text-olive-500 uppercase mb-1">Evento</p>
          <div className="text-sm text-olive-700 space-y-0.5">
            <p>{lead.event.eventType} · {lead.event.desiredDurationHours}h · ~{lead.event.estimatedPeople}p</p>
            {lead.event.possibleDates.length > 0 && (
              <p className="text-xs text-olive-500">{lead.event.possibleDates.join(', ')}</p>
            )}
          </div>
        </div>
      )}

      {/* Contact person (if set) */}
      {lead.contactPerson && (
        <div className="border-t border-earth-100 px-4 py-3">
          <p className="text-xs font-semibold text-olive-500 uppercase mb-1">Contato</p>
          <p className="text-sm text-olive-700">{lead.contactPerson.name} · {lead.contactPerson.whatsapp}</p>
          {lead.contactPerson.role && <p className="text-xs text-olive-400">{lead.contactPerson.role}</p>}
        </div>
      )}

      {/* Add-ons (if set) */}
      {lead.addOns && lead.addOns.length > 0 && (
        <div className="border-t border-earth-100 px-4 py-3">
          <p className="text-xs font-semibold text-olive-500 uppercase mb-1">Adicionais</p>
          <div className="flex flex-wrap gap-1">
            {lead.addOns.map(a => (
              <span key={a.id} className="text-xs bg-earth-100 text-olive-700 px-1.5 py-0.5 rounded-full">{a.name} x{a.quantity}</span>
            ))}
          </div>
        </div>
      )}

      {/* Commission (if set) */}
      {lead.commission && (
        <div className="border-t border-earth-100 px-4 py-3">
          <p className="text-xs font-semibold text-olive-500 uppercase mb-1">Comissão</p>
          <p className="text-sm text-terracotta-500 font-medium">R$ {(lead.commission.amount / 100).toFixed(2)} — {lead.commission.status}</p>
        </div>
      )}

      {/* Visit (if set) */}
      {lead.visit && (
        <div className="border-t border-earth-100 px-4 py-3">
          <p className="text-xs font-semibold text-olive-500 uppercase mb-1">Visita</p>
          <div className="text-sm text-olive-700 space-y-0.5">
            <p>{lead.visit.date} às {lead.visit.timeSlot}</p>
            <p>Guia: {lead.visit.guideId === 'dev-guide-uid' ? 'Guia (dev)' : lead.visit.guideId}</p>
            <p className="flex items-center gap-1.5">{lead.visit.confirmed ? <><Check size={14} weight="bold" className="text-olive-600" />Confirmado</> : 'Pendente'}</p>
            {lead.visit.feedback && <p className="text-xs text-olive-500">{lead.visit.feedback.result} — {lead.visit.feedback.notes}</p>}
          </div>
        </div>
      )}

      {/* Notes — collapsible section */}
      <div className="border-t border-earth-100">
        <button onClick={() => setShowNotes(!showNotes)}
          className="w-full flex items-center justify-between px-4 py-3 text-sm text-olive-700 hover:bg-earth-50 transition-colors">
          <span>📝 Notas {lead.notes.length > 0 && `(${lead.notes.length})`}</span>
          <span className="text-xs text-olive-400">{showNotes ? '▲' : '▼'}</span>
        </button>
        {showNotes && (
          <div className="px-4 pb-3 space-y-2">
            {lead.notes.length === 0 && <p className="text-xs text-olive-400">Nenhuma nota.</p>}
            {lead.notes.map((n, i) => (
              <div key={i} className="text-xs text-olive-700 border-b border-earth-50 pb-1">
                <p>{n.text}</p>
                <p className="text-olive-400 mt-0.5">{new Date(n.createdAt).toLocaleString()}</p>
              </div>
            ))}
            <div className="flex gap-1 pt-1">
              <input value={note} onChange={e => setNote(e.target.value)}
                placeholder="Nova nota..." className="flex-1 border border-earth-200 rounded px-2 py-1 text-xs"
                onKeyDown={e => e.key === 'Enter' && handleAddNote()} />
              <button onClick={handleAddNote} className="bg-olive-500 text-white px-2 py-1 rounded text-xs hover:bg-olive-700">Salvar</button>
            </div>
          </div>
        )}
      </div>

    </div>
  );
}
