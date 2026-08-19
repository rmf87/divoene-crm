// Pure reducer for pipeline state. The API is the single source of truth;
// local state is a thin optimistic mirror (undo/redo in-memory). Undoable
// actions (MOVE_LEAD, ADD_NOTE, CREATE_LEAD, BOOK_VISIT) are tracked by
// react-undo-redo; non-undoable actions update state without history.
import type { Lead, EventInfo, ContactPerson, AddOnItem } from './usePipeline';
import { VALID_TRANSITIONS, stageLabels } from './usePipeline';

export interface PipelineState {
  leads: Lead[];
  apiError: boolean;
  loading: boolean;
  toast: string | null;
}

export type PipelineAction =
  | { type: 'SET_LEADS'; leads: Lead[] }
  | { type: 'SET_API_ERROR'; error: boolean }
  | { type: 'SET_LOADING'; loading: boolean }
  | { type: 'SET_TOAST'; toast: string | null }
  | { type: 'MOVE_LEAD'; leadId: string; toStage: string }
  | { type: 'ADD_NOTE'; leadId: string; text: string; createdBy: string }
  | { type: 'CREATE_LEAD'; lead: Lead }
  | { type: 'UPDATE_EVENT'; leadId: string; event: EventInfo; contactPerson: ContactPerson; addOns: AddOnItem[] }
  | { type: 'BOOK_VISIT'; leadId: string; visit: { guideId: string; guideName: string; date: string; timeSlot: string; visitId: string } }
  | { type: 'SET_VISIT_CONFIRMED'; leadId: string }
  | { type: 'SET_VISIT_FEEDBACK'; leadId: string; result: string; notes: string }
  | { type: 'UPDATE_CONTRACTING_PARTY'; leadId: string; contractingParty: import('./usePipeline').ContractingParty }
  | { type: 'SET_VISIT_GUIDE'; leadId: string; guideId: string }
  | { type: 'SET_VISIT_REASSIGN'; leadId: string; visit: { guideId: string; date: string; timeSlot: string } };

function mapLead(leads: Lead[], leadId: string, fn: (l: Lead) => Lead): Lead[] {
  return leads.map(l => (l.id === leadId ? fn(l) : l));
}

export function pipelineReducer(state: PipelineState, action: PipelineAction): PipelineState {
  switch (action.type) {
    case 'SET_LEADS':
      return { ...state, leads: action.leads };

    case 'SET_API_ERROR':
      return { ...state, apiError: action.error };

    case 'SET_LOADING':
      return { ...state, loading: action.loading };

    case 'SET_TOAST':
      return { ...state, toast: action.toast };

    case 'MOVE_LEAD': {
      const lead = state.leads.find(l => l.id === action.leadId);
      if (!lead) return state;

      const allowed = VALID_TRANSITIONS[lead.stage] || [];
      if (!allowed.includes(action.toStage)) {
        return {
          ...state,
          toast: `Transição inválida: ${stageLabels[lead.stage] || lead.stage} → ${stageLabels[action.toStage] || action.toStage}`,
        };
      }

      return {
        ...state,
        leads: mapLead(state.leads, action.leadId, l => ({
          ...l,
          stage: action.toStage,
          lastStageChange: new Date().toISOString(),
        })),
        toast: `${lead.name} → ${stageLabels[action.toStage] || action.toStage}`,
      };
    }

    case 'ADD_NOTE': {
      const exists = state.leads.some(l => l.id === action.leadId);
      if (!exists) return state;
      return {
        ...state,
        leads: mapLead(state.leads, action.leadId, l => ({
          ...l,
          notes: [...(l.notes || []), { text: action.text, createdBy: action.createdBy, createdAt: new Date().toISOString() }],
        })),
      };
    }

    case 'CREATE_LEAD':
      return {
        ...state,
        leads: [action.lead, ...state.leads],
        toast: `Lead ${action.lead.name} criado`,
      };

    case 'UPDATE_EVENT': {
      const exists = state.leads.some(l => l.id === action.leadId);
      if (!exists) return state;
      return {
        ...state,
        leads: mapLead(state.leads, action.leadId, l => ({
          ...l,
          event: action.event,
          contactPerson: action.contactPerson,
          addOns: action.addOns,
        })),
        toast: `Dados do evento salvos`,
      };
    }

    case 'BOOK_VISIT': {
      const lead = state.leads.find(l => l.id === action.leadId);
      if (!lead) return state;
      return {
        ...state,
        leads: mapLead(state.leads, action.leadId, l => ({
          ...l,
          stage: 'visit_scheduled',
          lastStageChange: new Date().toISOString(),
          visit: { guideId: action.visit.guideId, date: action.visit.date, timeSlot: action.visit.timeSlot, confirmed: false, visitId: action.visit.visitId },
        })),
        toast: `${lead.name} → ${stageLabels['visit_scheduled']}`,
      };
    }

    case 'SET_VISIT_CONFIRMED': {
      return {
        ...state,
        leads: mapLead(state.leads, action.leadId, l =>
          l.visit ? { ...l, visit: { ...l.visit, confirmed: true } } : l
        ),
      };
    }

    case 'SET_VISIT_FEEDBACK': {
      return {
        ...state,
        leads: mapLead(state.leads, action.leadId, l =>
          l.visit ? { ...l, visit: { ...l.visit, feedback: { result: action.result, notes: action.notes } } } : l
        ),
      };
    }

    case 'UPDATE_CONTRACTING_PARTY': {
      return {
        ...state,
        leads: mapLead(state.leads, action.leadId, l => ({ ...l, contractingParty: action.contractingParty })),
      };
    }

    case 'SET_VISIT_GUIDE': {
      return {
        ...state,
        leads: mapLead(state.leads, action.leadId, l =>
          l.visit ? { ...l, visit: { ...l.visit, guideId: action.guideId } } : l
        ),
      };
    }

    case 'SET_VISIT_REASSIGN': {
      return {
        ...state,
        leads: mapLead(state.leads, action.leadId, l =>
          l.visit ? { ...l, visit: { ...l.visit, guideId: action.visit.guideId, date: action.visit.date, timeSlot: action.visit.timeSlot } } : l
        ),
        toast: 'Visita realocada',
      };
    }

    default:
      return state;
  }
}

// Only track mutations the user can reverse
export function isUndoableAction(action: PipelineAction): boolean {
  return action.type === 'MOVE_LEAD' || action.type === 'ADD_NOTE' || action.type === 'CREATE_LEAD' || action.type === 'BOOK_VISIT';
}
