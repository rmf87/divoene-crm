// LeadProcess — Stage "lead": tabbed validation form with v-* checkmark IDs.
// Three tabs: Evento / Contato / Adicionais.
// Each tab is independently validatable. Data persists to reducer on each save.
// Final "Validar Lead" button appears only after all 3 groups are saved.
import { useState } from 'react';
import { Check, CheckCircle } from '@phosphor-icons/react';
import { ADDON_CATALOG, stageLabels, type Lead, type EventInfo, type ContactPerson, type AddOnItem, type ContractingParty } from '../../hooks/usePipeline';

interface Props {
  lead: Lead;
  dispatch: (action: any) => void;
  patchValidation: (leadId: string, payload: { event?: EventInfo; contactPerson?: ContactPerson; addOns?: AddOnItem[] }) => void;
}

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

type Tab = 'evento' | 'contato' | 'adicionais' | 'contratante';

// ── per-lead localStorage helpers ────────────────────────────────

function loadData(leadId: string, key: string, fallback: any) {
  try {
    const v = localStorage.getItem(`divoene_${key}_${leadId}`);
    return v ? JSON.parse(v) : fallback;
  } catch { return fallback; }
}

function saveData(leadId: string, key: string, data: any) {
  localStorage.setItem(`divoene_${key}_${leadId}`, JSON.stringify(data));
}

// ── component ────────────────────────────────────────────────────

export default function LeadProcess({ lead, dispatch, patchValidation }: Props) {
  const [tab, setTab] = useState<Tab>('evento');
  const [submitted, setSubmitted] = useState(false);

  // ── per-group "saved" flags (which tabs have been validated) ────
  type GroupKey = 'evento' | 'contato' | 'adicionais' | 'contratante';
  const [savedGroups, setSavedGroups] = useState<GroupKey[]>(() =>
    loadData(lead.id, 'groups', [])
  );

  function markGroupSaved(g: GroupKey) {
    setSavedGroups(prev => {
      const next = prev.includes(g) ? prev : [...prev, g];
      saveData(lead.id, 'groups', next);
      return next;
    });
  }

  const gruposSalvos = new Set(savedGroups);
  const allSaved = gruposSalvos.has('evento') && gruposSalvos.has('contato') && gruposSalvos.has('adicionais') && gruposSalvos.has('contratante');

  // ── confirmed ✓ checkmarks per field ────────────────────────────
  const [confirmed, setConfirmed] = useState<string[]>(() =>
    loadData(lead.id, 'v', [])
  );

  function toggleV(key: string) {
    setConfirmed(prev => {
      const next = prev.includes(key) ? prev.filter(k => k !== key) : [...prev, key];
      saveData(lead.id, 'v', next);
      return next;
    });
  }

  // ── form state (initialised from lead or localStorage) ──────────
  const [eventType, setEventType] = useState(() =>
    lead.event?.eventType || loadData(lead.id, 'eventType', lead.product || '')
  );
  const [date1, setDate1] = useState(() =>
    lead.event?.possibleDates?.[0] || loadData(lead.id, 'date1', lead.desiredDate || '')
  );
  const [date2, setDate2] = useState(() =>
    lead.event?.possibleDates?.[1] || loadData(lead.id, 'date2', '')
  );
  const [date3, setDate3] = useState(() =>
    lead.event?.possibleDates?.[2] || loadData(lead.id, 'date3', '')
  );
  const [dayOfWeek, setDayOfWeek] = useState(() =>
    lead.event?.desiredDayOfWeek || loadData(lead.id, 'dayOfWeek', '')
  );
  const [duration, setDuration] = useState(() =>
    lead.event?.desiredDurationHours || loadData(lead.id, 'duration', 4)
  );
  const [people, setPeople] = useState(() =>
    lead.event?.estimatedPeople || loadData(lead.id, 'people', 50)
  );
  const [contactName, setContactName] = useState(() =>
    lead.contactPerson?.name || loadData(lead.id, 'contactName', lead.name || '')
  );
  const [contactWpp, setContactWpp] = useState(() =>
    lead.contactPerson?.whatsapp || loadData(lead.id, 'contactWpp', lead.whatsapp || '')
  );
  const [contactRole, setContactRole] = useState(() =>
    lead.contactPerson?.role || loadData(lead.id, 'contactRole', '')
  );
  const [taxID, setTaxID] = useState(() =>
    lead.contractingParty?.taxID || loadData(lead.id, 'taxID', '')
  );
  const [zipcode, setZipcode] = useState(() =>
    lead.contractingParty?.address?.zipcode || loadData(lead.id, 'zipcode', '')
  );
  const [addrStreet, setAddrStreet] = useState(() =>
    lead.contractingParty?.address?.street || loadData(lead.id, 'addrStreet', '')
  );
  const [addrNumber, setAddrNumber] = useState(() =>
    lead.contractingParty?.address?.number || loadData(lead.id, 'addrNumber', '')
  );
  const [addrNeighborhood, setAddrNeighborhood] = useState(() =>
    lead.contractingParty?.address?.neighborhood || loadData(lead.id, 'addrNeighborhood', '')
  );
  const [addrCity, setAddrCity] = useState(() =>
    lead.contractingParty?.address?.city || loadData(lead.id, 'addrCity', '')
  );
  const [addrState, setAddrState] = useState(() =>
    lead.contractingParty?.address?.state || loadData(lead.id, 'addrState', 'SP')
  );
  const [addrComplement, setAddrComplement] = useState(() =>
    lead.contractingParty?.address?.complement || loadData(lead.id, 'addrComplement', '')
  );

  const [selectedAddOns, setSelectedAddOns] = useState<Record<string, number>>(() => {
    const stored = loadData(lead.id, 'addOns', null);
    if (stored) return stored;
    const m: Record<string, number> = {};
    lead.addOns?.forEach(a => { m[a.id] = a.quantity; });
    return m;
  });

  const [error, setError] = useState('');

  const catalog = eventType ? (ADDON_CATALOG[eventType] || []) : [];
  const hasDates = date1 || date2 || date3;

  // ── build current validation payload ────────────────────────────
  function buildPayload() {
    const dates = [date1, date2, date3].filter(Boolean);
    const addOns = catalog
      .filter(a => (selectedAddOns[a.id] || 0) > 0)
      .map(a => ({ id: a.id, name: a.name, quantity: selectedAddOns[a.id] }));
    return {
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
    };
  }

  // ── per-tab "all confirmed" checks ────────────────────────────────
  const eventoFields = ['evento-tipo', 'duracao', 'pessoas', hasDates ? 'datas' : 'dia-semana'];
  const eventoAllV = eventoFields.every(f => confirmed.includes(f));

  const contatoFields = ['contato-nome', 'contato-wpp', 'contato-role'];
  const contratanteFields = ['contratante-cpf', 'contratante-cep', 'contratante-rua', 'contratante-numero', 'contratante-bairro', 'contratante-cidade', 'contratante-estado'];
  const contratanteAllV = contratanteFields.every(f => confirmed.includes(f));
  const contatoAllV = contatoFields.every(f => confirmed.includes(f));

  const addonSelected = catalog.filter(a => (selectedAddOns[a.id] || 0) > 0);
  const adicionaisAllV = addonSelected.length === 0 || addonSelected.every(a => confirmed.includes(`adicionais-${a.id}`));

  // ── per-tab save handlers ───────────────────────────────────────
  function handleSaveEvento() {
    setError('');
    if (!eventType) { setError('Selecione o tipo de evento.'); return; }
    if (!hasDates && !dayOfWeek) { setError('Informe ao menos uma data possível ou o dia da semana.'); return; }

    // persist form values so they survive refresh
    saveData(lead.id, 'eventType', eventType);
    saveData(lead.id, 'date1', date1);
    saveData(lead.id, 'date2', date2);
    saveData(lead.id, 'date3', date3);
    saveData(lead.id, 'dayOfWeek', dayOfWeek);
    saveData(lead.id, 'duration', duration);
    saveData(lead.id, 'people', people);

    const payload = buildPayload();
    dispatch({ type: 'UPDATE_EVENT', leadId: lead.id, ...payload });
    patchValidation(lead.id, { event: payload.event });
    markGroupSaved('evento');
  }

  function handleSaveContato() {
    setError('');
    if (!contactName.trim()) { setError('Nome do contato é obrigatório.'); return; }

    saveData(lead.id, 'contactName', contactName);
    saveData(lead.id, 'contactWpp', contactWpp);
    saveData(lead.id, 'contactRole', contactRole);

    const payload = buildPayload();
    dispatch({ type: 'UPDATE_EVENT', leadId: lead.id, ...payload });
    patchValidation(lead.id, { contactPerson: payload.contactPerson });
    markGroupSaved('contato');
  }

  function handleSaveAdicionais() {
    setError('');
    saveData(lead.id, 'addOns', selectedAddOns);

    const payload = buildPayload();
    dispatch({ type: 'UPDATE_EVENT', leadId: lead.id, ...payload });
    patchValidation(lead.id, { addOns: payload.addOns });
    markGroupSaved('adicionais');
  }

  // ── save contratante ─────────────────────────────────────────
  function handleSaveContratante() {
    setError('');
    if (!taxID.trim()) { setError('CPF/CNPJ é obrigatório.'); return; }
    saveData(lead.id, 'taxID', taxID);
    saveData(lead.id, 'zipcode', zipcode);
    saveData(lead.id, 'addrStreet', addrStreet);
    saveData(lead.id, 'addrNumber', addrNumber);
    saveData(lead.id, 'addrNeighborhood', addrNeighborhood);
    saveData(lead.id, 'addrCity', addrCity);
    saveData(lead.id, 'addrState', addrState);
    saveData(lead.id, 'addrComplement', addrComplement);

    const cp: ContractingParty = {
      taxID: taxID.trim(),
      address: {
        zipcode: zipcode.trim(),
        street: addrStreet.trim(),
        number: addrNumber.trim(),
        neighborhood: addrNeighborhood.trim(),
        city: addrCity.trim(),
        state: addrState.trim(),
        complement: addrComplement.trim() || undefined,
      },
    };
    dispatch({ type: 'UPDATE_CONTRACTING_PARTY', leadId: lead.id, contractingParty: cp });
    markGroupSaved('contratante');
  }

  // ── final promote
  // ── final promote (all groups saved) ────────────────────────────
  function handleFinalValidate() {
    dispatch({ type: 'MOVE_LEAD', leadId: lead.id, toStage: 'validated' });
    // clean up localStorage
    ['groups','v','eventType','date1','date2','date3','dayOfWeek','duration','people','contactName','contactWpp','contactRole','addOns','taxID','zipcode','addrStreet','addrNumber','addrNeighborhood','addrCity','addrState','addrComplement'].forEach(k =>
      localStorage.removeItem(`divoene_${k}_${lead.id}`)
    );
    setSubmitted(true);
  }

  // ── UI helpers ──────────────────────────────────────────────────
  const tabs: { key: Tab; label: string }[] = [
    { key: 'evento', label: 'Evento' },
    { key: 'contato', label: 'Contato' },
    { key: 'adicionais', label: 'Adicionais' },
    { key: 'contratante', label: 'Contratante' },
  ];

  const VMark = ({ field }: { field: string }) => {
    const isConfirmed = confirmed.includes(field);
    return (
      <button
        id={`v-confirm-${field}`}
        type="button"
        onClick={() => toggleV(field)}
        className={`flex-shrink-0 w-7 h-7 rounded-full border-2 flex items-center justify-center text-sm font-bold transition-colors min-h-[32px] ${
          isConfirmed ? 'bg-olive-500 border-olive-500 text-white' : 'border-earth-200 text-transparent hover:border-olive-300'
        }`}
        title={isConfirmed ? 'Confirmado ✓' : 'Clique para confirmar'}
      >
        {isConfirmed ? <Check size={16} weight="bold" /> : null}
      </button>
    );
  };

  function tabBadge(g: GroupKey) {
    if (!gruposSalvos.has(g)) return null;
    return <Check size={14} weight="bold" className="ml-1 inline text-olive-500" />;
  }

  // ── submitted state ─────────────────────────────────────────────
  if (submitted) {
    return (
      <div className="flex-1 min-w-0 bg-olive-50 rounded-2xl border border-olive-200 p-6 flex items-center justify-center">
        <div className="text-center">
          <CheckCircle size={40} weight="fill" className="mx-auto text-olive-500 mb-2" />
          <p className="font-serif text-lg italic text-olive-900">Lead validado!</p>
          <p className="text-sm text-olive-700 mt-1">{lead.name} → {stageLabels.validated}</p>
        </div>
      </div>
    );
  }

  // ── render ──────────────────────────────────────────────────────
  return (
    <div id="process-lead" className="flex-1 min-w-0 bg-white rounded-2xl border border-earth-200 overflow-hidden">
      {/* Header */}
      <div className="px-5 py-4 border-b border-earth-100 flex items-center gap-3">
        <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-olive-500 text-white">
          {stageLabels.lead}
        </span>
        <h2 className="font-serif text-lg italic text-olive-900">Validar Lead</h2>
        <span className="text-xs text-olive-400 ml-auto">{savedGroups.length}/4 grupos salvos</span>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-earth-200" role="tablist">
        {tabs.map(({ key, label }) => (
          <button
            key={key}
            id={`v-tab-${key}`}
            role="tab"
            aria-selected={tab === key}
            onClick={() => setTab(key)}
            className={`flex-1 px-4 py-3 text-sm font-medium transition-colors min-h-[44px] ${
              tab === key
                ? 'text-olive-900 border-b-2 border-olive-500 bg-olive-50'
                : 'text-olive-400 hover:text-olive-700 hover:bg-earth-50'
            }`}
          >
            {label}{tabBadge(key)}
          </button>
        ))}
      </div>

      <div className="p-5 space-y-4">
        {/* ── Tab: Evento ─────────────────────────────────── */}
        {tab === 'evento' && (
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="flex-1">
                <label className="block text-sm font-medium text-olive-900 mb-1">Tipo de evento</label>
                <select id="validate-event-type" value={eventType} onChange={e => setEventType(e.target.value)}
                  className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm">
                  <option value="">Selecione...</option>
                  {PRODUCTS.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </div>
              <div className="pt-6"><VMark field="evento-tipo" /></div>
            </div>

            <div className="flex items-start gap-3">
              <div className="flex-1">
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
              <div className="pt-6"><VMark field="datas" /></div>
            </div>

            {!hasDates && (
              <div className="flex items-start gap-3">
                <div className="flex-1">
                  <label className="block text-sm font-medium text-olive-900 mb-1">Dia da semana preferido</label>
                  <select id="validate-day" value={dayOfWeek} onChange={e => setDayOfWeek(e.target.value)}
                    className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm">
                    <option value="">Qualquer dia</option>
                    {DAYS.map(d => <option key={d} value={d}>{DAY_LABELS[d]}</option>)}
                  </select>
                </div>
                <div className="pt-6"><VMark field="dia-semana" /></div>
              </div>
            )}

            <div className="flex items-start gap-3">
              <div className="flex-1">
                <label className="block text-sm font-medium text-olive-900 mb-1">Duração (horas)</label>
                <input id="validate-duration" type="number" min={2} max={12} value={duration}
                  onChange={e => setDuration(Number(e.target.value))}
                  className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
              </div>
              <div className="pt-6"><VMark field="duracao" /></div>
            </div>

            <div className="flex items-start gap-3">
              <div className="flex-1">
                <label className="block text-sm font-medium text-olive-900 mb-1">Pessoas estimadas</label>
                <input id="validate-people" type="number" min={1} max={500} value={people}
                  onChange={e => setPeople(Number(e.target.value))}
                  className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
              </div>
              <div className="pt-6"><VMark field="pessoas" /></div>
            </div>

            <div className="flex justify-end gap-3 pt-3 border-t border-earth-200">
              <button
                id="v-save-evento"
                type="button"
                onClick={handleSaveEvento}
                disabled={!eventoAllV}
                className="bg-olive-500 text-white px-5 py-2.5 rounded-lg text-sm font-semibold hover:bg-olive-700 disabled:opacity-50 disabled:cursor-not-allowed min-h-[44px]"
              >
                {gruposSalvos.has('evento') ? <><Check size={16} weight="bold" className="inline -mt-0.5 mr-1" />Salvo</> : 'Salvar Evento'}
              </button>
            </div>
          </div>
        )}

        {/* ── Tab: Contato ────────────────────────────────── */}
        {tab === 'contato' && (
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="flex-1">
                <label className="block text-sm font-medium text-olive-900 mb-1">Nome do contato</label>
                <input id="validate-contact-name" value={contactName} onChange={e => setContactName(e.target.value)}
                  placeholder="Nome completo" className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
              </div>
              <div className="pt-6"><VMark field="contato-nome" /></div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-1">
                <label className="block text-sm font-medium text-olive-900 mb-1">WhatsApp do contato</label>
                <input id="validate-contact-wpp" value={contactWpp} onChange={e => setContactWpp(e.target.value)}
                  placeholder="55..." className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
              </div>
              <div className="pt-6"><VMark field="contato-wpp" /></div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-1">
                <label className="block text-sm font-medium text-olive-900 mb-1">Papel / relação</label>
                <input id="validate-contact-role" value={contactRole} onChange={e => setContactRole(e.target.value)}
                  placeholder="Ex: noiva, organizador, mãe" className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
              </div>
              <div className="pt-6"><VMark field="contato-role" /></div>
            </div>

            <div className="flex justify-end gap-3 pt-3 border-t border-earth-200">
              <button
                id="v-save-contato"
                type="button"
                onClick={handleSaveContato}
                disabled={!contatoAllV}
                className="bg-olive-500 text-white px-5 py-2.5 rounded-lg text-sm font-semibold hover:bg-olive-700 disabled:opacity-50 disabled:cursor-not-allowed min-h-[44px]"
              >
                {gruposSalvos.has('contato') ? <><Check size={16} weight="bold" className="inline -mt-0.5 mr-1" />Salvo</> : 'Salvar Contato'}
              </button>
            </div>
          </div>
        )}

        {/* ── Tab: Adicionais ──────────────────────────────── */}
        {tab === 'adicionais' && (
          <div className="space-y-4">
            {catalog.length === 0 && (
              <p className="text-sm text-olive-400 text-center py-4">Selecione um tipo de evento na aba "Evento" para ver os adicionais disponíveis.</p>
            )}
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
                <VMark field={`adicionais-${a.id}`} />
              </div>
            ))}

            <div className="flex justify-end gap-3 pt-3 border-t border-earth-200">
              <button
                id="v-save-adicionais"
                type="button"
                onClick={handleSaveAdicionais}
                disabled={!adicionaisAllV}
                className="bg-olive-500 text-white px-5 py-2.5 rounded-lg text-sm font-semibold hover:bg-olive-700 disabled:opacity-50 disabled:cursor-not-allowed min-h-[44px]"
              >
                {gruposSalvos.has('adicionais') ? <><Check size={16} weight="bold" className="inline -mt-0.5 mr-1" />Salvo</> : 'Salvar Adicionais'}
              </button>
            </div>
          </div>
        )}

        {/* ── Tab: Contratante ────────────────────────────── */}
        {tab === 'contratante' && (
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="flex-1">
                <label className="block text-sm font-medium text-olive-900 mb-1">CPF/CNPJ do contratante</label>
                <input id="validate-taxid" value={taxID} onChange={e => setTaxID(e.target.value)}
                  placeholder="000.000.000-00" className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
              </div>
              <div className="pt-6"><VMark field="contratante-cpf" /></div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="flex items-start gap-3">
                <div className="flex-1">
                  <label className="block text-sm font-medium text-olive-900 mb-1">CEP</label>
                  <input id="validate-zipcode" value={zipcode} onChange={e => setZipcode(e.target.value)}
                    placeholder="00000-000" className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
                </div>
                <div className="pt-6"><VMark field="contratante-cep" /></div>
              </div>
              <div className="flex items-start gap-3">
                <div className="flex-1">
                  <label className="block text-sm font-medium text-olive-900 mb-1">Estado</label>
                  <select id="validate-state" value={addrState} onChange={e => setAddrState(e.target.value)}
                    className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm">
                    {['AC','AL','AP','AM','BA','CE','DF','ES','GO','MA','MT','MS','MG','PA','PB','PR','PE','PI','RJ','RN','RS','RO','RR','SC','SP','SE','TO'].map(s => <option key={s} value={s}>{s}</option>)}
                  </select>
                </div>
                <div className="pt-6"><VMark field="contratante-estado" /></div>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-1">
                <label className="block text-sm font-medium text-olive-900 mb-1">Rua</label>
                <input id="validate-street" value={addrStreet} onChange={e => setAddrStreet(e.target.value)}
                  className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
              </div>
              <div className="pt-6"><VMark field="contratante-rua" /></div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="flex items-start gap-3">
                <div className="flex-1">
                  <label className="block text-sm font-medium text-olive-900 mb-1">Número</label>
                  <input id="validate-addr-number" value={addrNumber} onChange={e => setAddrNumber(e.target.value)}
                    className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
                </div>
                <div className="pt-6"><VMark field="contratante-numero" /></div>
              </div>
              <div className="flex items-start gap-3">
                <div className="flex-1">
                  <label className="block text-sm font-medium text-olive-900 mb-1">Bairro</label>
                  <input id="validate-neighborhood" value={addrNeighborhood} onChange={e => setAddrNeighborhood(e.target.value)}
                    className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
                </div>
                <div className="pt-6"><VMark field="contratante-bairro" /></div>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="flex items-start gap-3">
                <div className="flex-1">
                  <label className="block text-sm font-medium text-olive-900 mb-1">Cidade</label>
                  <input id="validate-city" value={addrCity} onChange={e => setAddrCity(e.target.value)}
                    className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
                </div>
                <div className="pt-6"><VMark field="contratante-cidade" /></div>
              </div>
              <div className="flex-1">
                <label className="block text-sm font-medium text-olive-900 mb-1">Complemento</label>
                <input id="validate-complement" value={addrComplement} onChange={e => setAddrComplement(e.target.value)}
                  placeholder="Opcional" className="w-full border border-earth-200 rounded-lg px-3 py-2.5 text-sm" />
              </div>
            </div>
            <div className="flex justify-end gap-3 pt-3 border-t border-earth-200">
              <button
                id="v-save-contratante"
                type="button"
                onClick={handleSaveContratante}
                disabled={!contratanteAllV}
                className="bg-olive-500 text-white px-5 py-2.5 rounded-lg text-sm font-semibold hover:bg-olive-700 disabled:opacity-50 disabled:cursor-not-allowed min-h-[44px]"
              >
                {gruposSalvos.has('contratante') ? <><Check size={16} weight="bold" className="inline -mt-0.5 mr-1" />Salvo</> : 'Salvar Contratante'}
              </button>
            </div>
          </div>
        )}

        {/* Error */}
        {error && <p className="text-sm text-terracotta-500">{error}</p>}

        {/* ── Final promote (only when all 3 groups saved) ─── */}
        {allSaved && (
          <div className="pt-3 border-t border-olive-200">
            <button
              id="v-submit-validacao"
              type="button"
              onClick={handleFinalValidate}
              className="w-full bg-olive-500 text-white px-6 py-3 rounded-xl text-sm font-semibold hover:bg-olive-700 min-h-[44px]"
            >
              <Check size={16} weight="bold" className="inline -mt-0.5 mr-1" />
              Validar Lead — promover para {stageLabels.validated}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
