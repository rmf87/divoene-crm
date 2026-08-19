import { useState } from 'react';
import { X } from '@phosphor-icons/react';
import type { EventInfo, ContactPerson, AddOnItem } from '../hooks/usePipeline';
import { ADDON_CATALOG } from '../hooks/usePipeline';

const PRODUCTS = [
  { id: 'ensaio_fotografico', name: 'Ensaio Fotográfico' },
  { id: 'locacao_eventos', name: 'Eventos' },
  { id: 'corporativo', name: 'Corporativo' },
  { id: 'casamentos', name: 'Casamentos' },
  { id: 'buffet_infantil', name: 'Buffet Infantil' },
  { id: 'passeios_escolares', name: 'Passeios Escolares' },
];

const DAYS = ['mon','tue','wed','thu','fri','sat','sun'];
const DAY_LABELS: Record<string, string> = {
  mon: 'Seg', tue: 'Ter', wed: 'Qua', thu: 'Qui', fri: 'Sex', sat: 'Sáb', sun: 'Dom',
};

interface Props {
  leadName: string;
  initial?: { event?: EventInfo; contactPerson?: ContactPerson; addOns?: AddOnItem[] };
  onSave: (data: { event: EventInfo; contactPerson: ContactPerson; addOns: AddOnItem[] }) => void;
  onClose: () => void;
}

export default function ValidateLeadForm({ leadName, initial, onSave, onClose }: Props) {
  const [eventType, setEventType] = useState(initial?.event?.eventType || '');
  const [date1, setDate1] = useState(initial?.event?.possibleDates?.[0] || '');
  const [date2, setDate2] = useState(initial?.event?.possibleDates?.[1] || '');
  const [date3, setDate3] = useState(initial?.event?.possibleDates?.[2] || '');
  const [dayOfWeek, setDayOfWeek] = useState(initial?.event?.desiredDayOfWeek || '');
  const [duration, setDuration] = useState(initial?.event?.desiredDurationHours || 4);
  const [people, setPeople] = useState(initial?.event?.estimatedPeople || 50);
  const [contactName, setContactName] = useState(initial?.contactPerson?.name || '');
  const [contactWpp, setContactWpp] = useState(initial?.contactPerson?.whatsapp || '');
  const [contactRole, setContactRole] = useState(initial?.contactPerson?.role || '');
  const [selectedAddOns, setSelectedAddOns] = useState<Record<string, number>>(() => {
    const m: Record<string, number> = {};
    initial?.addOns?.forEach(a => { m[a.id] = a.quantity; });
    return m;
  });

  const [error, setError] = useState('');

  const catalog = eventType ? (ADDON_CATALOG[eventType] || []) : [];
  const hasDates = date1 || date2 || date3;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    if (!eventType) return setError('Selecione o tipo de evento.');
    if (!hasDates && !dayOfWeek) return setError('Informe ao menos uma data possível ou o dia da semana desejado.');
    if (!contactName.trim()) return setError('Nome do contato é obrigatório.');

    const dates = [date1, date2, date3].filter(Boolean);
    const addOns: AddOnItem[] = catalog
      .filter(a => (selectedAddOns[a.id] || 0) > 0)
      .map(a => ({ id: a.id, name: a.name, quantity: selectedAddOns[a.id] }));

    onSave({
      event: {
        possibleDates: dates,
        eventType,
        desiredDurationHours: duration,
        desiredDayOfWeek: hasDates ? undefined : dayOfWeek,
        estimatedPeople: people,
      },
      contactPerson: {
        name: contactName.trim(),
        whatsapp: contactWpp.trim(),
        role: contactRole.trim() || undefined,
      },
      addOns,
    });
  }

  return (
    <div id="validate-lead-modal-overlay" className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4"
      onClick={onClose}>
      <div id="validate-lead-modal" className="bg-white rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto"
        onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className="sticky top-0 bg-white border-b border-earth-200 px-6 py-4 flex items-center justify-between rounded-t-2xl">
          <h2 className="font-serif text-xl italic text-olive-900">Validar Lead — {leadName}</h2>
          <button onClick={onClose} className="text-olive-400 hover:text-olive-700 p-1" aria-label="Fechar">
            <X size={20} weight="bold" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {/* Event type */}
          <div>
            <label className="block text-sm font-medium text-olive-900 mb-1">Tipo de evento *</label>
            <select id="validate-event-type" value={eventType} onChange={e => setEventType(e.target.value)}
              className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm">
              <option value="">Selecione...</option>
              {PRODUCTS.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
            </select>
          </div>

          {/* Possible dates */}
          <div>
            <label className="block text-sm font-medium text-olive-900 mb-1">Datas possíveis</label>
            <div className="grid grid-cols-3 gap-2">
              <input id="validate-date-1" type="date" value={date1} onChange={e => setDate1(e.target.value)}
                className="border border-earth-200 rounded-lg px-3 py-2 text-sm" />
              <input id="validate-date-2" type="date" value={date2} onChange={e => setDate2(e.target.value)}
                className="border border-earth-200 rounded-lg px-3 py-2 text-sm" />
              <input id="validate-date-3" type="date" value={date3} onChange={e => setDate3(e.target.value)}
                className="border border-earth-200 rounded-lg px-3 py-2 text-sm" />
            </div>
          </div>

          {/* Day of week (if no dates) */}
          {!hasDates && (
            <div>
              <label className="block text-sm font-medium text-olive-900 mb-1">Dia da semana preferido</label>
              <select id="validate-day" value={dayOfWeek} onChange={e => setDayOfWeek(e.target.value)}
                className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm">
                <option value="">Qualquer dia</option>
                {DAYS.map(d => <option key={d} value={d}>{DAY_LABELS[d]}</option>)}
              </select>
            </div>
          )}

          {/* Duration + People */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-olive-900 mb-1">Duração (horas)</label>
              <input id="validate-duration" type="number" min={2} max={12} value={duration}
                onChange={e => setDuration(Number(e.target.value))}
                className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-olive-900 mb-1">Pessoas estimadas</label>
              <input id="validate-people" type="number" min={1} max={500} value={people}
                onChange={e => setPeople(Number(e.target.value))}
                className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
            </div>
          </div>

          {/* Contact person */}
          <fieldset className="border border-earth-200 rounded-lg p-4">
            <legend className="text-sm font-medium text-olive-900 px-1">Pessoa de contato *</legend>
            <div className="space-y-3">
              <input id="validate-contact-name" value={contactName} onChange={e => setContactName(e.target.value)}
                placeholder="Nome completo *" className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm" />
              <input id="validate-contact-wpp" value={contactWpp} onChange={e => setContactWpp(e.target.value)}
                placeholder="WhatsApp" className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm" />
              <input id="validate-contact-role" value={contactRole} onChange={e => setContactRole(e.target.value)}
                placeholder="Papel (ex: noiva, organizador, mãe)" className="w-full border border-earth-200 rounded-lg px-3 py-2 text-sm" />
            </div>
          </fieldset>

          {/* Add-ons */}
          {catalog.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-olive-900 mb-2">Adicionais disponíveis</label>
              <div className="space-y-2">
                {catalog.map(a => (
                  <div key={a.id} className="flex items-center gap-3">
                    <input id={`addon-${a.id}`} type="checkbox" checked={(selectedAddOns[a.id] || 0) > 0}
                      onChange={e => setSelectedAddOns(prev => ({ ...prev, [a.id]: e.target.checked ? 1 : 0 }))}
                      className="rounded border-earth-200" />
                    <span className="text-sm text-olive-700 flex-1">{a.name}</span>
                    {(selectedAddOns[a.id] || 0) > 0 && (
                      <input type="number" min={1} max={10} value={selectedAddOns[a.id]}
                        onChange={e => setSelectedAddOns(prev => ({ ...prev, [a.id]: Number(e.target.value) }))}
                        className="w-16 border border-earth-200 rounded-lg px-2 py-1 text-sm text-center" />
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {error && <p className="text-sm text-terracotta-500">{error}</p>}

          <div className="flex justify-end gap-3 pt-2 border-t border-earth-200">
            <button type="button" onClick={onClose}
              className="px-5 py-2.5 rounded-lg text-sm text-olive-500 hover:bg-earth-100 min-h-[44px]">
              Cancelar
            </button>
            <button id="validate-submit" type="submit"
              className="bg-olive-500 text-white px-6 py-2.5 rounded-lg text-sm font-semibold hover:bg-olive-700 min-h-[44px]">
              Salvar Validação
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
