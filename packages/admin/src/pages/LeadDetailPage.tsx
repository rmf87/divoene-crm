// Full-page wrapper for /leads/:id route.
// 3-col on lg: LeadDetail card (left) + ProcessRouter (center) + WhatsAppChat (right).
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft } from '@phosphor-icons/react';
import { usePipeline } from '../hooks/usePipeline';
import { useAuthContext } from '../hooks/useAuth';
import LeadDetail from './LeadDetail';
import ProcessRouter from '../components/process/ProcessRouter';
import WhatsAppChat from '../components/WhatsAppChat';

export default function LeadDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { dispatch, getLeadById, patchValidation, createContract, createPayment, pollPayment } = usePipeline();
  const { user } = useAuthContext();

  const lead = id ? getLeadById(id) : undefined;
  if (!lead) return <div className="p-6 text-olive-700">Lead não encontrado.</div>;

  return (
    <div className="p-4 md:p-6">
      <button id="lead-detail-back" onClick={() => navigate('/pipeline')} className="text-olive-700 text-sm mb-4 hover:underline min-h-[40px] flex items-center gap-1.5">
        <ArrowLeft size={16} weight="bold" />
        Voltar ao Pipeline
      </button>
      <div className="flex flex-col lg:flex-row gap-6 items-start lg:h-[calc(100vh-140px)]">
        <LeadDetail leadId={lead.id} />
        <div className="flex-1 min-w-0 overflow-y-auto lg:h-full">
          <ProcessRouter lead={lead} dispatch={dispatch} patchValidation={patchValidation} createContract={createContract} createPayment={createPayment} pollPayment={pollPayment} />
        </div>
        <div className="w-full lg:w-[380px] flex-shrink-0 lg:h-full overflow-hidden">
          <WhatsAppChat
            leadId={lead.id}
            leadName={lead.name}
            leadWhatsApp={lead.whatsapp}
            product={lead.product}
            sellerName={user?.name || 'Seller'}
            stage={lead.stage}
            collapsible
          />
        </div>
      </div>
    </div>
  );
}
