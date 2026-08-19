// ProcessRouter — dispatches to the correct stage-specific process component.
import type { Lead, EventInfo, ContactPerson, AddOnItem } from '../../hooks/usePipeline';
import LeadProcess from './LeadProcess';
import ValidatedProcess from './ValidatedProcess';
import VisitScheduledProcess from './VisitScheduledProcess';
import VisitDoneProcess from './VisitDoneProcess';
import ContractProcess from './ContractProcess';
import PaidProcess from './PaidProcess';
import BookedProcess from './BookedProcess';
import TerminalProcess from './TerminalProcess';

interface Props {
  lead: Lead;
  dispatch: (action: any) => void;
  patchValidation: (leadId: string, payload: { event?: EventInfo; contactPerson?: ContactPerson; addOns?: AddOnItem[] }) => void;
  createContract: (leadId: string, data: any) => Promise<any>;
  createPayment: (leadId: string, contractId: string, data: any) => Promise<any>;
  pollPayment: (txID: string, onConfirmed: () => void) => () => void;
}

export default function ProcessRouter({ lead, dispatch, patchValidation, createContract, createPayment, pollPayment }: Props) {
  switch (lead.stage) {
    case 'lead':
      return <LeadProcess lead={lead} dispatch={dispatch} patchValidation={patchValidation} />;
    case 'validated':
      return <ValidatedProcess lead={lead} dispatch={dispatch} />;
    case 'visit_scheduled':
      return <VisitScheduledProcess lead={lead} dispatch={dispatch} />;
    case 'visit_done':
      return <VisitDoneProcess lead={lead} dispatch={dispatch} />;
    case 'contract':
      return <ContractProcess lead={lead} dispatch={dispatch} createContract={createContract} />;
    case 'paid':
      return <PaidProcess lead={lead} dispatch={dispatch} createPayment={createPayment} pollPayment={pollPayment} />;
    case 'booked':
      return <BookedProcess lead={lead} dispatch={dispatch} />;
    default:
      return <TerminalProcess lead={lead} />;
  }
}
