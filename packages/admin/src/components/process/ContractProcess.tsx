// ContractProcess — Stage "contract": generate and send contract from validated lead data.
import { useState } from 'react';
import { Calendar, CurrencyCircleDollar, FileText, Hourglass, User, XCircle } from '@phosphor-icons/react';
import { stageLabels, type Lead } from '../../hooks/usePipeline';

interface Props {
  lead: Lead;
  dispatch: (action: any) => void;
  createContract: (leadId: string, data: any) => Promise<any>;
}

const PRODUCT_PRICES: Record<string, number> = {
  ensaio_fotografico: 80000,   // R$  800,00
  locacao_eventos:   300000,   // R$3.000,00
  corporativo:       400000,   // R$4.000,00
  casamentos:        600000,   // R$6.000,00
  buffet_infantil:   500000,   // R$5.000,00
  passeios_escolares: 50000,   // R$  500,00
};

const ADDON_PRICES: Record<string, number> = {
  monitor: 15000, brinquedos: 30000, mesa_doces: 25000, decoracao_tematica: 80000,
  dj_banda: 120000, buffet_completo: 200000, decoracao_floral: 100000, fotografo_extra: 60000,
  projetor: 20000, coffee_break: 15000, recepcionista: 25000, estacionamento_vip: 10000,
  seguranca: 30000, limpeza_extra: 15000, gerador: 20000, iluminacao: 40000,
  maquiador: 20000, figurino: 15000, assistente: 10000,
  lanche_incluso: 8000, transporte: 50000,
};

export default function ContractProcess({ lead, dispatch, createContract }: Props) {
  const [paymentCond, setPaymentCond] = useState('50% sinal + 50% na véspera');
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [sent, setSent] = useState(false);

  const event = lead.event;
  const contact = lead.contactPerson;
  const addOns = lead.addOns || [];

  const basePrice = PRODUCT_PRICES[lead.product] || 0;
  const addOnTotal = addOns.reduce((sum, a) => sum + (ADDON_PRICES[a.id] || 0) * a.quantity, 0);
  const totalAmount = basePrice + addOnTotal;

  async function handleSend() {
    setError('');
    setLoading(true);
    try {
      await createContract(lead.id, {
        lead_name: lead.name,
        lead_email: '',  // will be asked if empty
        lead_whatsapp: contact?.whatsapp || lead.whatsapp,
        amount: totalAmount,
        product: lead.product,
        event_type: event?.eventType || lead.product,
        event_date: event?.possibleDates?.[0] || lead.desiredDate,
        event_duration: event?.desiredDurationHours || 4,
        estimated_people: event?.estimatedPeople || 50,
        contact_name: contact?.name || lead.name,
        contact_whatsapp: contact?.whatsapp || lead.whatsapp,
        contact_role: contact?.role || '',
        add_ons: addOns.map(a => ({ name: a.name, quantity: a.quantity, unit_price: ADDON_PRICES[a.id] || 0 })),
        payment_conditions: paymentCond,
        notes: notes || undefined,
      });
      setSent(true);
    } catch (e: any) {
      setError(e.message || 'Erro ao enviar contrato');
    } finally {
      setLoading(false);
    }
  }

  if (sent) {
    return (
      <div className="flex-1 min-w-0 bg-olive-50 rounded-2xl border border-olive-200 p-6 flex items-center justify-center">
        <div className="text-center">
          <FileText size={40} weight="fill" className="mx-auto text-olive-500 mb-2" />
          <p className="font-serif text-lg italic text-olive-900">Contrato enviado!</p>
          <p className="text-sm text-olive-700 mt-1">Aguardando assinatura do cliente via Clicksign.</p>
        </div>
      </div>
    );
  }

  return (
    <div id="process-contract" className="flex-1 min-w-0 bg-white rounded-2xl border border-earth-200 overflow-hidden">
      <div className="px-5 py-4 border-b border-earth-100 flex items-center gap-3">
        <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-olive-500 text-white">{stageLabels.contract}</span>
        <h2 className="font-serif text-lg italic text-olive-900">Enviar Contrato</h2>
      </div>
      <div className="p-5 space-y-4">
        {/* Event summary */}
        {event && (
          <div className="bg-earth-50 rounded-lg p-3 text-sm space-y-1">
            <p className="font-medium text-olive-900 flex items-center gap-1.5"><Calendar size={15} weight="fill" className="text-olive-600" /> {event.eventType && PRODUCT_LABELS[event.eventType] || lead.product}</p>
            <p className="text-olive-700">Data: {event.possibleDates?.[0] || lead.desiredDate}</p>
            <p className="text-olive-700">Duração: {event.desiredDurationHours}h · ~{event.estimatedPeople} pessoas</p>
          </div>
        )}

        {/* Contact */}
        {contact && (
          <div className="bg-earth-50 rounded-lg p-3 text-sm space-y-1">
            <p className="font-medium text-olive-900 flex items-center gap-1.5"><User size={15} weight="fill" className="text-olive-600" /> Contratante</p>
            <p className="text-olive-700">{contact.name} · {contact.whatsapp}</p>
            {contact.role && <p className="text-olive-500 text-xs">{contact.role}</p>}
          </div>
        )}

        {/* Add-ons + total */}
        <div className="bg-earth-50 rounded-lg p-3 text-sm space-y-1">
          <p className="font-medium text-olive-900 flex items-center gap-1.5"><CurrencyCircleDollar size={15} weight="fill" className="text-olive-600" /> Serviços</p>
          <div className="flex justify-between text-olive-700">
            <span>{PRODUCT_LABELS[lead.product] || lead.product}</span>
            <span>R$ {(basePrice / 100).toFixed(2)}</span>
          </div>
          {addOns.map(a => {
            const price = (ADDON_PRICES[a.id] || 0) * a.quantity;
            return (
              <div key={a.id} className="flex justify-between text-olive-600 text-xs">
                <span>+ {a.name} x{a.quantity}</span>
                <span>R$ {(price / 100).toFixed(2)}</span>
              </div>
            );
          })}
          <div className="flex justify-between font-bold text-olive-900 pt-1 border-t border-earth-200">
            <span>Total</span>
            <span>R$ {(totalAmount / 100).toFixed(2)}</span>
          </div>
        </div>

        {/* Payment conditions */}
        <div>
          <label className="block text-sm font-medium text-olive-900 mb-1">Condições de pagamento</label>
          <select id="contract-payment-cond" value={paymentCond} onChange={e => setPaymentCond(e.target.value)}
            className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm">
            <option value="50% sinal + 50% na véspera">50% sinal + 50% na véspera</option>
            <option value="50% sinal + 50% pós-evento">50% sinal + 50% pós-evento</option>
            <option value="100% à vista">100% à vista</option>
            <option value="30% sinal + 70% na véspera">30% sinal + 70% na véspera</option>
          </select>
        </div>

        {/* Notes */}
        <div>
          <label className="block text-sm font-medium text-olive-900 mb-1">Observações (opcional)</label>
          <textarea id="contract-notes" value={notes} onChange={e => setNotes(e.target.value)}
            rows={2} className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm"
            placeholder="Observações sobre o contrato..." />
        </div>

        {/* Error */}
        {error && <p className="text-sm text-terracotta-500 bg-terracotta-50 rounded-lg p-2">{error}</p>}

        {/* Actions */}
        <div className="space-y-2 pt-3 border-t border-earth-200">
          <button id="process-action-send-contract"
            onClick={handleSend}
            disabled={loading}
            className="w-full bg-olive-500 text-white px-5 py-3 rounded-xl text-sm font-semibold hover:bg-olive-700 disabled:opacity-50 min-h-[44px]">
            {loading ? <><Hourglass size={18} weight="bold" className="inline -mt-0.5 mr-2" />Enviando...</> : <><FileText size={18} weight="fill" className="inline -mt-0.5 mr-2" />Gerar e Enviar Contrato</>}
          </button>
          <button id="process-action-cancel"
            onClick={() => dispatch({ type: 'MOVE_LEAD', leadId: lead.id, toStage: 'cancelled' })}
            className="w-full bg-earth-100 text-terracotta-600 px-5 py-3 rounded-xl text-sm hover:bg-earth-200 min-h-[44px] flex items-center justify-center gap-2">
            <XCircle size={16} weight="bold" />
            Cancelar Lead
          </button>
        </div>
      </div>
    </div>
  );
}

const PRODUCT_LABELS: Record<string, string> = {
  ensaio_fotografico: 'Ensaio Fotográfico',
  locacao_eventos: 'Locação para Eventos',
  corporativo: 'Corporativo',
  casamentos: 'Casamentos',
  buffet_infantil: 'Buffet Infantil',
  passeios_escolares: 'Passeios Escolares',
};
