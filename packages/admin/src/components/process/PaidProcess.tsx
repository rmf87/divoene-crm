// PaidProcess — Stage "paid": generate PIX charge, show QR code + copia-e-cola + link.
import { useState, useEffect, useRef } from 'react';
import { Check, Copy, CreditCard, CurrencyCircleDollar, DeviceMobile, Hourglass, LinkSimple, Package, XCircle } from '@phosphor-icons/react';
import { stageLabels, type Lead } from '../../hooks/usePipeline';

interface Props {
  lead: Lead;
  dispatch: (action: any) => void;
  createPayment: (leadId: string, contractId: string, data: any) => Promise<any>;
  pollPayment: (txID: string, onConfirmed: () => void) => () => void;
}

export default function PaidProcess({ lead, dispatch, createPayment, pollPayment }: Props) {
  const [chargeType, setChargeType] = useState<'pix' | 'pix_credit'>('pix');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [payment, setPayment] = useState<any>(null);
  const [copied, setCopied] = useState(false);
  const stopPoll = useRef<(() => void) | null>(null);

  // Cleanup polling on unmount
  useEffect(() => () => { stopPoll.current?.(); }, []);

  async function handleGenerate() {
    setError('');
    setLoading(true);
    try {
      const contact = lead.contactPerson;
      const cp = lead.contractingParty;
      const contractAmount = 100000; // default — should come from contract
      const result = await createPayment(lead.id, lead.contractId || '', {
        amount: contractAmount,
        description: `Evento - ${contact?.name || lead.name}`,
        type: chargeType === 'pix_credit' ? 'total' : 'sinal',
        charge_type: chargeType,
        payer_email: '',  // will be asked if empty
        payer_name: contact?.name || lead.name,
        payer_taxid: cp?.taxID,
        payer_phone: contact?.whatsapp || lead.whatsapp,
      });
      setPayment(result);

      // Start polling for confirmation
      if (result.openpix_transaction_id) {
        stopPoll.current = pollPayment(result.openpix_transaction_id, () => {
          dispatch({ type: 'MOVE_LEAD', leadId: lead.id, toStage: 'paid' });
        });
      }
    } catch (e: any) {
      setError(e.message || 'Erro ao gerar cobrança');
    } finally {
      setLoading(false);
    }
  }

  function handleCopyBrCode() {
    if (payment?.br_code) {
      navigator.clipboard.writeText(payment.br_code);
      setCopied(true);
      setTimeout(() => setCopied(false), 3000);
    }
  }

  if (payment) {
    return (
      <div id="process-paid" className="flex-1 min-w-0 bg-white rounded-2xl border border-earth-200 overflow-hidden">
        <div className="px-5 py-4 border-b border-earth-100 flex items-center gap-3">
          <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-olive-500 text-white">{stageLabels.paid}</span>
          <h2 className="font-serif text-lg italic text-olive-900">Cobrança PIX</h2>
        </div>
        <div className="p-5 space-y-4">
          {/* QR Code */}
          {payment.br_code && (
            <div className="text-center space-y-3">
              <p className="text-sm text-olive-700">Escaneie o QR Code ou use o PIX copia-e-cola</p>
              <div className="bg-white inline-block p-4 rounded-xl border border-earth-200">
                {/* QR Code placeholder — in production, render QR from brCode */}
                <div className="w-48 h-48 bg-earth-100 rounded-lg mx-auto flex items-center justify-center text-olive-400">
                  <DeviceMobile size={56} weight="regular" />
                </div>
              </div>
              <div className="bg-earth-50 rounded-lg p-3">
                <p className="text-xs text-olive-500 mb-1">PIX Copia-e-Cola</p>
                <p className="text-xs text-olive-700 break-all font-mono">{payment.br_code}</p>
                <button onClick={handleCopyBrCode}
                  className="mt-2 text-xs bg-olive-500 text-white px-4 py-1.5 rounded-lg hover:bg-olive-700 min-h-[36px] flex items-center gap-1.5 mx-auto">
                  {copied ? <><Check size={14} weight="bold" />Copiado!</> : <><Copy size={14} weight="bold" />Copiar PIX</>}
                </button>
              </div>
            </div>
          )}

          {/* Payment link */}
          {payment.payment_link_url && (
            <div className="text-center">
              <a href={payment.payment_link_url} target="_blank" rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 text-sm bg-earth-100 text-olive-700 px-5 py-2.5 rounded-lg hover:bg-earth-200 min-h-[44px]">
                <LinkSimple size={16} weight="bold" />
                Abrir link de pagamento
              </a>
            </div>
          )}

          {/* Status */}
          <div className="bg-olive-50 rounded-lg p-3 text-center text-sm text-olive-700 flex items-center justify-center gap-2">
            <Hourglass size={16} weight="bold" className="text-olive-500" />
            <p>Aguardando pagamento...</p>
            {payment.openpix_transaction_id && (
              <p className="text-xs text-olive-400 mt-1">ID: {payment.openpix_transaction_id}</p>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div id="process-paid" className="flex-1 min-w-0 bg-white rounded-2xl border border-earth-200 overflow-hidden">
      <div className="px-5 py-4 border-b border-earth-100 flex items-center gap-3">
        <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-olive-500 text-white">{stageLabels.paid}</span>
        <h2 className="font-serif text-lg italic text-olive-900">Gerar Cobrança PIX</h2>
      </div>
      <div className="p-5 space-y-4">
        {/* Contract summary */}
        <div className="bg-earth-50 rounded-lg p-3 text-sm space-y-1">
          <p className="font-medium text-olive-900 flex items-center gap-1.5"><Check size={16} weight="fill" className="text-olive-600" /> Contrato assinado</p>
          <p className="text-olive-700">{lead.name}</p>
          {lead.contactPerson && (
            <p className="text-olive-600 text-xs">{lead.contactPerson.name} · {lead.contactPerson.whatsapp}</p>
          )}
        </div>

        {/* Charge type */}
        <div>
          <label className="block text-sm font-medium text-olive-900 mb-2">Tipo de cobrança</label>
          <div className="grid grid-cols-2 gap-3">
            <button
              type="button"
              onClick={() => setChargeType('pix')}
              className={`p-4 rounded-xl border-2 text-sm font-medium transition-colors min-h-[80px] ${
                chargeType === 'pix'
                  ? 'border-olive-500 bg-olive-50 text-olive-900'
                  : 'border-earth-200 text-olive-500 hover:border-olive-300'
              }`}
            >
              <span className="block text-xl mb-1 flex justify-center"><CreditCard size={26} weight="fill" className="text-olive-600" /></span>
              PIX à vista
            </button>
            <button
              type="button"
              onClick={() => setChargeType('pix_credit')}
              className={`p-4 rounded-xl border-2 text-sm font-medium transition-colors min-h-[80px] ${
                chargeType === 'pix_credit'
                  ? 'border-olive-500 bg-olive-50 text-olive-900'
                  : 'border-earth-200 text-olive-500 hover:border-olive-300'
              }`}
            >
              <span className="block text-xl mb-1 flex justify-center"><Package size={26} weight="fill" className="text-olive-600" /></span>
              PIX Parcelado
            </button>
          </div>
          {chargeType === 'pix_credit' && (
            <p className="text-xs text-olive-400 mt-2">
              O cliente poderá escolher o número de parcelas no link de pagamento. Requer CPF + endereço na validação.
            </p>
          )}
        </div>

        {/* Error */}
        {error && <p className="text-sm text-terracotta-500 bg-terracotta-50 rounded-lg p-2">{error}</p>}

        {/* Actions */}
        <div className="space-y-2 pt-3 border-t border-earth-200">
          <button id="process-action-generate-pix"
            onClick={handleGenerate}
            disabled={loading}
            className="w-full bg-olive-500 text-white px-5 py-3 rounded-xl text-sm font-semibold hover:bg-olive-700 disabled:opacity-50 min-h-[44px]">
            {loading ? <><Hourglass size={18} weight="bold" className="inline -mt-0.5 mr-2" />Gerando...</> : <><CurrencyCircleDollar size={18} weight="fill" className="inline -mt-0.5 mr-2" />Gerar Cobrança PIX</>}
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
