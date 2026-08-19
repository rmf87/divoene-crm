import LeadCard from './LeadCard';
import { Lead, stageLabels } from '../hooks/usePipeline';

interface Props {
  stage: string;
  leads: Lead[];
  onDrop: (leadId: string, stage: string) => void;
  onPromote: (leadId: string, toStage: string) => boolean;
  onSelect?: (leadId: string) => void;
  selectedId?: string | null;
}

export default function StageColumn({ stage, leads, onDrop, onPromote, onSelect, selectedId }: Props) {
  function handleDragOver(e: React.DragEvent) {
    e.preventDefault();
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault();
    const leadId = e.dataTransfer.getData('text/plain');
    if (leadId) onDrop(leadId, stage);
  }

  function handleDragStart(e: React.DragEvent, leadId: string) {
    e.dataTransfer.setData('text/plain', leadId);
  }

  function daysSince(date: string): number {
    return Math.floor((Date.now() - new Date(date).getTime()) / 86400000);
  }

  return (
    <div
      id={`stage-${stage}`}
      onDragOver={handleDragOver}
      onDrop={handleDrop}
      data-dropzone={stage}
      className="bg-earth-100 rounded-2xl p-3 flex flex-col gap-2 min-h-0 h-full"
    >
      <div className="flex justify-between items-center px-1 shrink-0">
        <h3 className="font-semibold text-sm text-olive-900">{stageLabels[stage] || stage}</h3>
        <span className="text-xs text-olive-500 bg-white px-2 py-0.5 rounded-full">{leads.length}</span>
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto space-y-2 pr-0.5">
        {leads.map(lead => (
          <LeadCard
            key={lead.id}
            id={lead.id}
            name={lead.name}
            product={lead.product}
            whatsapp={lead.whatsapp}
            stage={lead.stage}
            daysInStage={daysSince(lead.lastStageChange)}
            onDragStart={handleDragStart}
            onPromote={onPromote}
            onSelect={onSelect}
            selected={selectedId === lead.id}
          />
        ))}
      </div>
    </div>
  );
}
