import { useState, useRef, useEffect, useCallback } from 'react';
import { CaretDown, CaretUp, Check, Checks, WhatsappLogo } from '@phosphor-icons/react';
import { fetchWithAuth } from '../hooks/useAuth';

interface Message {
  id: string;
  from: 'seller' | 'lead' | 'system';
  text: string;
  timestamp: string;
  status?: 'sent' | 'delivered' | 'read' | 'received';
}

interface Props {
  leadId: string;
  leadName: string;
  leadWhatsApp: string;
  product: string;
  sellerName?: string;
  stage?: string;
  collapsible?: boolean;
}

const TEMPLATES: Record<string, string> = {
  first_contact: 'Olá {lead_name}! Aqui é {seller_name} da Chácara Divoene. Vi que você se interessou por {product}. Como posso ajudar?',
  follow_up: 'Olá {lead_name}! Passando para saber se você viu o contrato que enviamos. Podemos conversar?',
  payment_reminder: 'Olá {lead_name}! O pagamento do sinal está pendente. Posso ajudar com o PIX?',
  visit_confirmation: 'Olá {lead_name}! Sua visita está confirmada para {date} às {time}. Nos vemos lá!',
};

function fillTemplate(template: string, vars: Record<string, string>): string {
  let result = template;
  for (const [key, value] of Object.entries(vars)) {
    result = result.replace(`{${key}}`, value);
  }
  return result;
}

function toMessage(raw: any): Message {
  // raw.id is an INTEGER (SQLite autoincrement) — always stringify it so id
  // checks (e.g. isMock -> startsWith) never throw.
  const rawId = raw.id ?? raw.wa_message_id ?? '';
  return {
    id: String(rawId || `msg-${Date.now()}`),
    from: raw.direction === 'lead' ? 'lead' : raw.direction === 'system' ? 'system' : 'seller',
    text: raw.body ?? '',
    timestamp: raw.sent_at ?? raw.created_at ?? new Date().toISOString(),
    status: raw.status ?? 'sent',
  };
}

export default function WhatsAppChat({ leadId, leadName, leadWhatsApp, product, sellerName = 'Seller', collapsible }: Props) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [template, setTemplate] = useState('first_contact');
  const [useTemplate, setUseTemplate] = useState(true);
  const [error, setError] = useState('');
  const [mockMode, setMockMode] = useState(false);
  const [leadMissing, setLeadMissing] = useState(false);
  const [open, setOpen] = useState(() => {
    if (!collapsible) return true;
    const v = localStorage.getItem(`divoene_chat_open_${leadId}`);
    return v === 'true';
  });
  useEffect(() => {
    if (collapsible) localStorage.setItem(`divoene_chat_open_${leadId}`, String(open));
  }, [open, leadId, collapsible]);
  const bottomRef = useRef<HTMLDivElement>(null);
  const isMock = (m: Message) => typeof m.id === 'string' && m.id.startsWith('wamid-mock-');

  const loadMessages = useCallback(async () => {
    try {
      const res = await fetchWithAuth(`/api/leads/${leadId}/messages`);
      if (res.status === 404) {
        setLeadMissing(true);
        setMessages([]);
        setError('');
        return;
      }
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const list = await res.json();
      const mapped = Array.isArray(list) ? list.map(toMessage) : [];
      setLeadMissing(false);
      setMessages(mapped);
      setMockMode(mapped.some(isMock));
      setError('');
    } catch (e: any) {
      setError(`Não foi possível carregar as mensagens${e && e.message ? ` (${e.message})` : ''}.`);
    }
  }, [leadId]);

  useEffect(() => {
    loadMessages();
  }, [leadId, loadMessages]);

  // Poll for inbound messages every 5 seconds.
  useEffect(() => {
    const t = setInterval(loadMessages, 5000);
    return () => clearInterval(t);
  }, [leadId, loadMessages]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  async function send(payload: any) {
    try {
      const res = await fetchWithAuth(`/api/leads/${leadId}/messages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || `HTTP ${res.status}`);
      }
      const created = await res.json();
      const msg = toMessage(created);
      setMessages(prev => [...prev, msg]);
      setMockMode(isMock(msg));
      setError('');
    } catch (e: any) {
      setError(e.message || 'Falha ao enviar a mensagem.');
    }
  }

  function handleSend() {
    const text = input.trim();
    if (!text) return;
    send({ body: text });
    setInput('');
    setUseTemplate(false);
  }

  function handleSendTemplate() {
    const vars: Record<string, string> = {
      lead_name: leadName,
      seller_name: sellerName,
      product: productLabel(product),
      date: new Date().toLocaleDateString('pt-BR'),
    };
    send({ template, lang: 'pt_BR', vars });
    setUseTemplate(false);
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  if (collapsible && !open) {
    return (
      <div id="whatsapp-chat" className="bg-white rounded-2xl border border-earth-200 overflow-hidden">
        <button onClick={() => setOpen(true)}
          className="w-full bg-[#075E54] px-4 py-3 flex items-center gap-3 hover:bg-[#064b40] transition-colors">
          <div className="w-8 h-8 rounded-full bg-[#25D366] flex items-center justify-center text-white font-bold text-sm">
            {leadName.charAt(0)}
          </div>
          <div className="flex-1 text-left">
            <p className="text-white font-semibold text-sm flex items-center gap-1.5"><WhatsappLogo size={15} weight="fill" /> WhatsApp</p>
            <p className="text-green-100 text-xs">Clique para abrir</p>
          </div>
          <span className="text-green-100 text-sm"><CaretDown size={16} weight="bold" /></span>
        </button>
      </div>
    );
  }

  return (
    <div id="whatsapp-chat" className="bg-white rounded-2xl border border-earth-200 overflow-hidden flex flex-col h-full">
      {/* Header */}
      <div className="bg-[#075E54] px-4 py-3 flex items-center gap-3">
        <div className="w-10 h-10 rounded-full bg-[#25D366] flex items-center justify-center text-white font-bold text-lg">
          {leadName.charAt(0)}
        </div>
        <div className="flex-1">
          <p className="text-white font-semibold text-sm">{leadName}</p>
          <p className="text-green-100 text-xs">{leadWhatsApp}</p>
        </div>
        {collapsible && (
          <button onClick={() => setOpen(false)} className="text-green-100 text-sm hover:text-white p-1" aria-label="Recolher">
            <CaretUp size={16} weight="bold" />
          </button>
        )}
        <span className="text-green-100 text-xs bg-[#128C7E] px-2 py-1 rounded-full">WhatsApp</span>
      </div>

      {/* Messages */}
      <div id="whatsapp-messages" className="flex-1 min-h-0 overflow-y-auto p-4 space-y-3 bg-[#ECE5DD]">
        {messages.length === 0 && !useTemplate && (
          <p className="text-gray-500 text-sm text-center py-8">
            Nenhuma mensagem ainda. Use o template abaixo para iniciar a conversa.
          </p>
        )}

        {messages.map((msg, i) => (
          <div
            key={`${msg.id}-${i}`}
            id={`whatsapp-msg-${i}`}
            className={`flex ${msg.from === 'seller' ? 'justify-end' : 'justify-start'}`}
          >
            <div
              className={`max-w-[80%] rounded-lg px-3 py-2 text-sm ${
                msg.from === 'seller'
                  ? 'bg-[#DCF8C6] text-gray-900'
                  : msg.from === 'system'
                  ? 'bg-yellow-100 text-yellow-800 italic'
                  : 'bg-white text-gray-900'
              }`}
            >
              <p>{msg.text}</p>
              <div className="flex items-center justify-end gap-1 mt-1">
                <span className="text-xs text-gray-500">
                  {new Date(msg.timestamp).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })}
                </span>
                {msg.from === 'seller' && msg.status && (
                  <span className="text-xs text-gray-400">
                    {msg.status === 'sent' ? <Check size={12} weight="bold" /> : <Checks size={12} weight="bold" />}
                  </span>
                )}
              </div>
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      {/* Template selector */}
      {useTemplate && messages.length === 0 && (
        <div className="px-4 py-3 bg-white border-t border-earth-100">
          <label className="text-xs text-olive-700 font-semibold block mb-2">Template de primeira mensagem:</label>
          <select
            id="whatsapp-template-select"
            value={template}
            onChange={e => setTemplate(e.target.value)}
            className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm mb-2"
          >
            <option value="first_contact">Primeiro contato</option>
            <option value="follow_up">Follow-up</option>
            <option value="payment_reminder">Lembrete de pagamento</option>
            <option value="visit_confirmation">Confirmação de visita</option>
          </select>
          <p id="whatsapp-template-preview" className="text-xs text-olive-500 mb-2 italic">
            {fillTemplate(TEMPLATES[template], {
              lead_name: leadName,
              seller_name: sellerName,
              product: productLabel(product),
              date: new Date().toLocaleDateString('pt-BR'),
            })}
          </p>
          <button
            id="whatsapp-send-template"
            onClick={handleSendTemplate}
            className="w-full bg-[#25D366] text-white py-2 rounded-lg text-sm font-semibold hover:bg-[#128C7E] transition-colors"
          >
            Enviar mensagem via WhatsApp
          </button>
        </div>
      )}

      {/* Message input */}
      <div className="px-3 py-2 bg-gray-100 flex items-center gap-2">
        <input
          id="whatsapp-message-input"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Digite sua mensagem..."
          className="flex-1 border border-earth-200 rounded-full px-4 py-2 text-sm focus:outline-none focus:border-[#25D366]"
        />
        <button
          id="whatsapp-send-btn"
          onClick={handleSend}
          disabled={!input.trim()}
          className="w-10 h-10 rounded-full bg-[#25D366] text-white flex items-center justify-center hover:bg-[#128C7E] disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>
        </button>
      </div>

      {/* Status banner */}
      {leadMissing && (
        <div id="whatsapp-lead-missing" className="px-4 py-3 bg-red-50 border-t border-red-200 text-center">
          <p className="text-xs text-red-700 mb-2">Este lead não existe no servidor (id local antigo).</p>
          <button
            onClick={() => window.location.reload()}
            className="text-xs bg-red-600 text-white px-3 py-1.5 rounded-lg hover:bg-red-700"
          >
            Recarregar pipeline
          </button>
        </div>
      )}
      {error && (
        <div className="px-4 py-2 bg-red-50 border-t border-red-200">
          <p className="text-xs text-red-700 text-center">{error}</p>
        </div>
      )}
      {mockMode && (
        <div id="whatsapp-mock-banner" className="px-4 py-2 bg-amber-50 border-t border-amber-200">
          <p className="text-xs text-amber-700 text-center">
            Modo mock — sem credenciais WhatsApp configuradas. Integração ativada ao configurar WHATSAPP_TOKEN.
          </p>
        </div>
      )}
    </div>
  );
}

function productLabel(id: string): string {
  const labels: Record<string, string> = {
    ensaio_fotografico: 'Ensaio Fotográfico',
    locacao_eventos: 'Locação para Eventos',
    corporativo: 'Corporativo',
    casamentos: 'Casamentos',
    buffet_infantil: 'Buffet Infantil',
    passeios_escolares: 'Passeios Escolares',
  };
  return labels[id] || id;
}
