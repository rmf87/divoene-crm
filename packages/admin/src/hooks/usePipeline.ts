// Híbrido: busca leads reais da API (site → Go service → Firestore emulator)
// e mantém localStorage para fallback e dados mock de desenvolvimento.
// Usa react-undo-redo para undo/redo de mutações (moveLead, addNote, createLead).
import { useState, useEffect, useCallback, useMemo } from 'react';
import { createUndoRedo } from './useUndoRedo';
import { useAuthContext, fetchWithAuth } from './useAuth';
import { pipelineReducer, isUndoableAction, type PipelineState, type PipelineAction } from './pipelineReducer';

export interface EventInfo {
  possibleDates: string[];
  eventType: string;
  desiredDurationHours: number;
  desiredDayOfWeek?: string;
  estimatedPeople: number;
}

export interface ContactPerson {
  name: string;
  whatsapp: string;
  role?: string;
}

export interface ContractingParty {
  taxID: string;
  address: {
    zipcode: string;
    street: string;
    number: string;
    neighborhood: string;
    city: string;
    state: string;
    complement?: string;
  };
}

export interface AddOnItem {
  id: string;
  name: string;
  quantity: number;
}

export const ADDON_CATALOG: Record<string, { id: string; name: string }[]> = {
  buffet_infantil: [
    { id: 'monitor', name: 'Monitor' }, { id: 'brinquedos', name: 'Brinquedos' },
    { id: 'mesa_doces', name: 'Mesa de doces' }, { id: 'decoracao_tematica', name: 'Decoração temática' },
  ],
  casamentos: [
    { id: 'dj_banda', name: 'DJ/Banda' }, { id: 'buffet_completo', name: 'Buffet completo' },
    { id: 'decoracao_floral', name: 'Decoração floral' }, { id: 'fotografo_extra', name: 'Fotógrafo extra' },
  ],
  corporativo: [
    { id: 'projetor', name: 'Projetor' }, { id: 'coffee_break', name: 'Coffee break' },
    { id: 'recepcionista', name: 'Recepcionista' }, { id: 'estacionamento_vip', name: 'Estacionamento VIP' },
  ],
  locacao_eventos: [
    { id: 'seguranca', name: 'Segurança' }, { id: 'limpeza_extra', name: 'Limpeza extra' },
    { id: 'gerador', name: 'Gerador' }, { id: 'iluminacao', name: 'Iluminação' },
  ],
  ensaio_fotografico: [
    { id: 'maquiador', name: 'Maquiador' }, { id: 'figurino', name: 'Figurino' },
    { id: 'assistente', name: 'Assistente' },
  ],
  passeios_escolares: [
    { id: 'monitor', name: 'Monitor' }, { id: 'lanche_incluso', name: 'Lanche incluso' },
    { id: 'transporte', name: 'Transporte' },
  ],
};

export interface Lead {
  id: string;
  name: string;
  whatsapp: string;
  product: string;
  desiredDate: string;
  source: string;
  stage: string;
  lastStageChange: string;
  assignedSeller: string;
  createdBy: string;
  notes: { text: string; createdBy: string; createdAt: string }[];
  stageHistory?: StageHistoryEntry[];
  visit?: { visitId?: string; guideId: string; date: string; timeSlot: string; confirmed: boolean; feedback?: { result: string; notes: string } };
  commission?: { amount: number; status: string };
  event?: EventInfo;
  contactPerson?: ContactPerson;
  addOns?: AddOnItem[];
  contractingParty?: ContractingParty;
  contractId?: string;
  paymentId?: string;
}

export interface StageHistoryEntry {
  stage: string;
  changedAt: string;
  changedBy: string;
}

const ACTIVE_STAGES = ['lead', 'validated', 'visit_scheduled', 'visit_done', 'contract', 'paid', 'booked'];
export const STAGES = [...ACTIVE_STAGES, 'completed', 'cancelled'];

// Every active stage can transition to cancelled. Completed only from booked.
export const VALID_TRANSITIONS: Record<string, string[]> = {
  lead: ['validated', 'cancelled'],
  validated: ['visit_scheduled', 'cancelled'],
  visit_scheduled: ['visit_done', 'cancelled'],
  visit_done: ['contract', 'cancelled'],
  contract: ['paid', 'cancelled'],
  paid: ['booked', 'cancelled'],
  booked: ['completed', 'cancelled'],
};

export function getNextStages(stage: string): string[] {
  return VALID_TRANSITIONS[stage] || [];
}

const API_URL = import.meta.env.VITE_LEAD_SERVICE_URL || '/api/leads';

function toLead(apiLead: Record<string, unknown>): Lead {
  const rawEvent = apiLead.event as Record<string, unknown> | undefined;
  const rawContact = apiLead.contact_person as Record<string, unknown> | undefined;
  const rawAddOns = apiLead.add_ons as Array<Record<string, unknown>> | undefined;
  return {
    id: (apiLead.id as string) || '',
    name: (apiLead.name as string) || '',
    whatsapp: (apiLead.whatsapp as string) || '',
    product: (apiLead.product as string) || '',
    desiredDate: (apiLead.desired_date as string) || '',
    source: (apiLead.source as string) || '',
    stage: (apiLead.stage as string) || 'lead',
    lastStageChange: new Date().toISOString(),
    assignedSeller: '',
    createdBy: 'site',
    notes: [],
    stageHistory: (apiLead.stage_history as StageHistoryEntry[]) || [],
    event: rawEvent ? {
      possibleDates: (rawEvent.possible_dates as string[]) || [],
      eventType: (rawEvent.event_type as string) || '',
      desiredDurationHours: (rawEvent.desired_duration_hours as number) || 0,
      desiredDayOfWeek: rawEvent.desired_day_of_week as string | undefined,
      estimatedPeople: (rawEvent.estimated_people as number) || 0,
    } : undefined,
    contactPerson: rawContact ? {
      name: (rawContact.name as string) || '',
      whatsapp: (rawContact.whatsapp as string) || '',
      role: rawContact.role as string | undefined,
    } : undefined,
    addOns: rawAddOns ? rawAddOns.map(a => ({
      id: (a.id as string) || '',
      name: (a.name as string) || '',
      quantity: (a.quantity as number) || 1,
    })) : undefined,
    visit: apiLead.visit ? {
      visitId: (apiLead.visit as Record<string, unknown>).id as string || '',
      guideId: (apiLead.visit as Record<string, unknown>).guide_id as string || '',
      date: (apiLead.visit as Record<string, unknown>).date as string || '',
      timeSlot: (apiLead.visit as Record<string, unknown>).time_slot as string || '',
      confirmed: (apiLead.visit as Record<string, unknown>).status === 'confirmed' || (apiLead.visit as Record<string, unknown>).status === 'done',
      feedback: (apiLead.visit as Record<string, unknown>).feedback as { result: string; notes: string } | undefined,
    } : undefined,
  };
}

const stageLabels: Record<string, string> = {
  lead: 'Lead', validated: 'Validado', visit_scheduled: 'Visita Agendada', visit_done: 'Visita Feita',
  contract: 'Contrato', paid: 'Pago', booked: 'Reservado', completed: 'Concluído', cancelled: 'Cancelado',
};

export { stageLabels };

// Thin client: the API is the single source of truth for leads. Local state is
// an in-memory optimistic mirror (react-undo-redo); nothing lead-related is
// persisted in localStorage (removed to avoid stale client-only lead IDs).

// ── react-undo-redo provider + hooks ──────────────────────────────────────────

const { UndoRedoProvider, usePresent, useUndoRedo } = createUndoRedo(
  pipelineReducer,
  { track: isUndoableAction },
);

const initialPipelineState: PipelineState = {
  leads: [],
  apiError: false,
  loading: true,
  toast: null,
};

export { UndoRedoProvider as PipelineUndoProvider };
export { initialPipelineState };

// ── Hook ──────────────────────────────────────────────────────────────────────

export function usePipeline() {
  const { user } = useAuthContext();
  const [state, dispatch] = usePresent();
  const [undo, redo] = useUndoRedo();
  const { leads, apiError, loading, toast } = state;

  // One-shot cleanup of the old localStorage key (pre-thin-client versions kept
  // a local leads copy that could drift from the API with client-only IDs).
  useEffect(() => {
    localStorage.removeItem('divoene_leads');
  }, []);

  // ── enrich leads with visits ────────────────────────────────────────────────
  const VISITS_URL = import.meta.env.VITE_VISIT_SERVICE_URL || '/api/visits';

  async function enrichWithVisits(leads: Lead[]): Promise<Lead[]> {
    try {
      const res = await fetchWithAuth(VISITS_URL);
      if (!res.ok) return leads;
      const visits: Array<{ id: string; lead_id: string; guide_id: string; guide_name: string; date: string; time_slot: string; status: string; feedback?: { result: string; notes: string } }> = await res.json();
      if (!Array.isArray(visits)) return leads;
      return leads.map(lead => {
        if (lead.visit) return lead;
        const v = visits.find(v => v.lead_id === lead.id);
        if (!v) return lead;
        return {
          ...lead,
          visit: {
            visitId: v.id,
            guideId: v.guide_id,
            date: v.date,
            timeSlot: v.time_slot,
            confirmed: v.status === 'confirmed' || v.status === 'done',
            feedback: v.feedback,
          },
        };
      });
    } catch { return leads; }
  }

  // ── API fetch ──────────────────────────────────────────────────────────────
  const fetchLeads = useCallback(() => {
    dispatch({ type: 'SET_LOADING', loading: true });
    fetchWithAuth(API_URL)
      .then(res => res.json())
      .then((data: Record<string, unknown>[]) => {
        if (Array.isArray(data)) {
          const mapped = data.map(toLead);
          enrichWithVisits(mapped).then(enriched => {
            dispatch({ type: 'SET_LEADS', leads: enriched });
            dispatch({ type: 'SET_API_ERROR', error: false });
          });
        }
      })
      .catch(() => dispatch({ type: 'SET_API_ERROR', error: true }))
      .finally(() => dispatch({ type: 'SET_LOADING', loading: false }));
  }, [dispatch]);

  useEffect(() => { fetchLeads(); }, [fetchLeads]);

  // ── allLeads (single source: API) ──────────────────────────────────────────
  const allLeads = leads;

  function getFilteredLeads(): Lead[] {
    if (!user) return [];
    if (user.roles.includes('manager')) return allLeads;
    if (user.roles.includes('seller')) return allLeads.filter(l => l.assignedSeller === user.uid);
    if (user.roles.includes('associate')) return allLeads.filter(l => l.createdBy === user.uid);
    return [];
  }

  function getLeadById(id: string): Lead | undefined {
    return allLeads.find(l => l.id === id);
  }

  function dismissToast() {
    dispatch({ type: 'SET_TOAST', toast: null });
  }

  // ── syncDispatch: optimistic local update + server sync ───────────────────
  // Stage moves must reach the server so the pipeline (and WhatsApp templates)
  // reflect reality. Other actions stay optimistic-only to avoid double-creates.
  function syncDispatch(action: any) {
    dispatch(action);
    if (action.type === 'MOVE_LEAD') {
      fetchWithAuth(`${API_URL}/${action.leadId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ stage: action.toStage }),
      }).then(res => {
        if (!res.ok) { fetchLeads(); dispatch({ type: 'SET_TOAST', toast: 'Falha ao salvar a mudança de estágio.' }); }
      }).catch(() => fetchLeads());
    }
  }

  // ── mutation wrappers (pre-validate, then dispatch) ────────────────────────
  function moveLead(leadId: string, toStage: string): boolean {
    const lead = allLeads.find(l => l.id === leadId);
    if (!lead) return false;
    const allowed = VALID_TRANSITIONS[lead.stage] || [];
    if (!allowed.includes(toStage)) {
      dispatch({ type: 'SET_TOAST', toast: `Transição inválida: ${stageLabels[lead.stage] || lead.stage} → ${stageLabels[toStage] || toStage}` });
      return false;
    }
    syncDispatch({ type: 'MOVE_LEAD', leadId, toStage });
    return true;
  }

  function addNote(leadId: string, text: string) {
    if (!text.trim()) return;
    dispatch({ type: 'ADD_NOTE', leadId, text, createdBy: user?.uid || '' });
    fetchWithAuth(`${API_URL}/${leadId}/notes`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text }),
    }).then(res => {
      if (!res.ok) { fetchLeads(); dispatch({ type: 'SET_TOAST', toast: 'Falha ao salvar a nota.' }); }
    }).catch(() => fetchLeads());
  }

  // ── undo / redo ────────────────────────────────────────────────────────────
  function handleUndo() {
    if (!undo.isPossible) return;
    undo();
  }

  function handleRedo() {
    if (!redo.isPossible) return;
    redo();
  }

  async function createLead(leadData: Omit<Lead, 'id' | 'lastStageChange' | 'notes' | 'createdBy'>): Promise<Lead> {
    // Persist first so the lead gets the server ID (LEAD-…). The chat and
    // stage mutations talk to the API by lead ID, so a client-side-only
    // "L…" ID makes them fail with "lead não encontrado".
    try {
      const res = await fetchWithAuth(API_URL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: leadData.name,
          whatsapp: leadData.whatsapp,
          product: leadData.product,
          desired_date: leadData.desiredDate,
          source: leadData.source,
        }),
      });
      if (res.ok) {
        const created = await res.json();
        if (created && created.id) {
          const lead: Lead = { ...toLead(created), createdBy: user?.uid || '', assignedSeller: '' };
          dispatch({ type: 'CREATE_LEAD', lead });
          return lead;
        }
      }
    } catch {
      // best-effort — fall through to local-only creation below
    }

    const newLead: Lead = {
      ...leadData,
      id: `L${Date.now()}`,
      lastStageChange: new Date().toISOString(),
      notes: [],
      createdBy: user?.uid || '',
    };
    dispatch({ type: 'CREATE_LEAD', lead: newLead });
    return newLead;
  }

  // ── persist validation to backend ─────────────────────────────────────────
  async function patchValidation(leadId: string, payload: { event?: EventInfo; contactPerson?: ContactPerson; addOns?: AddOnItem[] }) {
    try {
      await fetchWithAuth(`${API_URL}/${leadId}/validation`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
    } catch {
      // best-effort — local state already updated via reducer
    }
  }

  // ── visit wrappers ──────────────────────────────────────────────────────────
  function bookVisit(leadId: string, visit: { guideId: string; guideName: string; date: string; timeSlot: string; visitId: string }) {
    dispatch({ type: 'BOOK_VISIT', leadId, visit });
  }

  function confirmVisitLocal(leadId: string) {
    dispatch({ type: 'SET_VISIT_CONFIRMED', leadId });
  }

  function setVisitFeedback(leadId: string, result: string, notes: string) {
    dispatch({ type: 'SET_VISIT_FEEDBACK', leadId, result, notes });
  }

  // ── contract creation ───────────────────────────────────────────────────
  async function createContract(leadId: string, contractData: any) {
    const CONTRACTS_API = '/api/contracts';
    const res = await fetchWithAuth(CONTRACTS_API, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ lead_id: leadId, ...contractData }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: 'erro desconhecido' }));
      throw new Error(err.error || 'erro ao criar contrato');
    }
    const contract = await res.json();
    syncDispatch({ type: 'MOVE_LEAD', leadId, toStage: 'contract' });
    return contract;
  }

  // ── payment creation ────────────────────────────────────────────────────
  async function createPayment(leadId: string, contractId: string, paymentData: any) {
    const PAYMENTS_API = '/api/payments';
    const res = await fetchWithAuth(PAYMENTS_API, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ lead_id: leadId, contract_id: contractId, ...paymentData }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: 'erro desconhecido' }));
      throw new Error(err.error || 'erro ao criar pagamento');
    }
    return res.json();
  }

  // ── payment polling ─────────────────────────────────────────────────────
  function pollPayment(transactionID: string, onConfirmed: () => void): () => void {
    const interval = setInterval(async () => {
      try {
        const res = await fetchWithAuth('/api/payments/' + transactionID + '/status');
        if (res.ok) {
          const data = await res.json();
          if (data.status === 'confirmed') {
            clearInterval(interval);
            onConfirmed();
          }
        }
      } catch { /* retry on next interval */ }
    }, 10_000);
    return () => clearInterval(interval);
  }

  function reassignVisitLocal(leadId: string, visit: { guideId: string; date: string; timeSlot: string }) {
    dispatch({ type: 'SET_VISIT_REASSIGN', leadId, visit });
  }

  return {
    leads: getFilteredLeads(),
    allLeads,
    stages: STAGES,
    moveLead,
    addNote,
    getLeadById,
    apiError,
    loading,
    toast,
    dismissToast,
    // undo/redo
    undo: handleUndo,
    redo: handleRedo,
    canUndo: undo.isPossible,
    canRedo: redo.isPossible,
    // dispatch for direct use (synced for stage moves)
    dispatch: syncDispatch,
    createLead,
    patchValidation,
    // contract & payment
    createContract,
    createPayment,
    pollPayment,
    // visit actions
    bookVisit,
    confirmVisitLocal,
    setVisitFeedback,
    reassignVisitLocal,
  };
}
