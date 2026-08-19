import { useNavigate } from 'react-router-dom';
import { ArrowRight, WhatsappLogo } from '@phosphor-icons/react';
import { getNextStages, stageLabels } from '../hooks/usePipeline';

interface Props {
  id: string;
  name: string;
  product: string;
  whatsapp: string;
  stage: string;
  daysInStage: number;
  onDragStart: (e: React.DragEvent, id: string) => void;
  onPromote: (leadId: string, toStage: string) => boolean;
  onSelect?: (leadId: string) => void;
  selected?: boolean;
}

const productLabels: Record<string, string> = {
  buffet_infantil: 'Buffet Infantil',
  ensaio_fotografico: 'Ensaio',
  locacao_eventos: 'Eventos',
  corporativo: 'Corporativo',
  casamentos: 'Casamentos',
  passeios_escolares: 'Escolas',
};

export default function LeadCard({ id, name, product, whatsapp, stage, daysInStage, onDragStart, onPromote, onSelect, selected }: Props) {
  const navigate = useNavigate();
  const nextStages = getNextStages(stage).filter(s => s !== 'cancelled');

  return (
    <div
      id={`lead-card-${id}`}
      draggable
      onDragStart={e => onDragStart(e, id)}
      className={`bg-white rounded-xl border transition-colors ${selected ? 'ring-2 ring-olive-500 border-olive-500' : 'border-earth-200'}`}
    >
      <div
        onClick={() => onSelect ? onSelect(id) : navigate(`/leads/${id}`)}
        className="p-3 cursor-pointer"
      >
        <p className="font-semibold text-sm text-olive-900">{name}</p>
        <p className="text-xs text-olive-700 mt-0.5">{productLabels[product] || product}</p>
        <div className="flex justify-between items-center mt-2">
          <span className="text-xs text-olive-500">{whatsapp}</span>
          <span className="text-xs text-olive-500 bg-earth-100 px-1.5 py-0.5 rounded">{daysInStage}d</span>
        </div>
      </div>

      {/* Promote buttons — always visible for touch discoverability */}
      {nextStages.length > 0 && (
        <div className="border-t border-earth-100 px-2 py-1.5 flex flex-wrap gap-1">
          {nextStages.map(ns => (
            <button
              key={ns}
              id={`lead-card-promote-${id}-${ns}`}
              onClick={e => { e.stopPropagation(); onPromote(id, ns); }}
              className="text-xs bg-olive-100 text-olive-800 hover:bg-olive-500 hover:text-white px-2 py-1 rounded transition-colors min-h-[28px]"
              title={`Promover para ${stageLabels[ns] || ns}`}
            >
              <ArrowRight size={12} weight="bold" className="inline -mt-0.5 mr-0.5" />
              {stageLabels[ns] || ns}
            </button>
          ))}
        </div>
      )}

      {/* Open chat — WhatsApp happens inside the admin panel (US_LEAD_004) */}
      <button
        id={`lead-card-whatsapp-${id}`}
        onClick={e => { e.stopPropagation(); navigate(`/leads/${id}`); }}
        className="block w-full text-center text-xs text-olive-600 hover:text-olive-800 border-t border-earth-100 py-1.5"
        title="Abrir chat com o lead"
      >
        <WhatsappLogo size={14} weight="fill" className="inline -mt-0.5 mr-1 text-olive-600" />
        WhatsApp
      </button>
    </div>
  );
}
