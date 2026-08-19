// ManageVisits — Admin view: all visits across all guides with reassignment modal.
import { useState, useEffect } from 'react';
import { usePipeline } from '../hooks/usePipeline';
import { fetchWithAuth } from '../hooks/useAuth';
import VisitScheduler from '../components/VisitScheduler';

interface GuideInfo {
  id: string;
  name: string;
}

export default function ManageVisits() {
  const { allLeads, reassignVisitLocal } = usePipeline();
  const [guides, setGuides] = useState<GuideInfo[]>([]);
  const [filterGuideId, setFilterGuideId] = useState('');
  const [reassignLead, setReassignLead] = useState<any>(null);

  useEffect(() => {
    fetchWithAuth('/api/guides')
      .then(r => r.ok ? r.json() : [])
      .then((data: GuideInfo[]) => { if (Array.isArray(data)) setGuides(data); })
      .catch(() => {});
  }, []);

  const visitLeads = allLeads.filter(l => l.visit);
  const filtered = filterGuideId
    ? visitLeads.filter(l => l.visit?.guideId === filterGuideId)
    : visitLeads;

  function handleReassign(leadId: string, visit: { guideId: string; date: string; timeSlot: string }) {
    reassignVisitLocal(leadId, visit);
    setReassignLead(null);
  }

  const stageLabels: Record<string, string> = {
    visit_scheduled: 'Agendada',
    visit_done: 'Realizada',
  };

  const FEEDBACK_LABELS: Record<string, string> = { liked: 'Gostou', disliked: 'Não gostou', maybe: 'Talvez' };
  const FEEDBACK_COLORS: Record<string, string> = {
    liked: 'bg-olive-100 text-olive-700',
    disliked: 'bg-red-50 text-red-600',
    maybe: 'bg-amber-100 text-amber-800',
  };

  return (
    <div id="page-manage-visits" className="max-w-6xl mx-auto p-4 md:p-6">
      <h1 id="manage-visits-heading" className="font-serif text-2xl italic text-olive-900 mb-2">Gerenciar Visitas</h1>
      <p className="text-sm text-olive-500 mb-6">Visualize e gerencie todas as visitas agendadas.</p>

      {/* Filter */}
      <div className="mb-4 flex items-center gap-3">
        <label htmlFor="manage-visits-filter-guide" className="text-xs text-olive-500">Filtrar por guia:</label>
        <select id="manage-visits-filter-guide" value={filterGuideId} onChange={e => setFilterGuideId(e.target.value)}
          className="border border-earth-200 rounded-lg px-3 py-2 text-sm bg-white">
          <option value="">Todos os guias</option>
          {guides.map(g => (
            <option key={g.id} value={g.id}>{g.name || g.id}</option>
          ))}
        </select>
      </div>

      {filtered.length === 0 ? (
        <p id="manage-visits-empty" className="text-olive-500">Nenhuma visita encontrada.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm" id="manage-visits-table">
            <thead>
              <tr className="border-b border-earth-200 text-left">
                <th className="p-3 text-olive-500 text-xs font-semibold">Lead</th>
                <th className="p-3 text-olive-500 text-xs font-semibold">Produto</th>
                <th className="p-3 text-olive-500 text-xs font-semibold">Data</th>
                <th className="p-3 text-olive-500 text-xs font-semibold">Horário</th>
                <th className="p-3 text-olive-500 text-xs font-semibold">Guia</th>
                <th className="p-3 text-olive-500 text-xs font-semibold">Status</th>
                <th className="p-3 text-olive-500 text-xs font-semibold">Ações</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(lead => (
                <tr key={lead.id} id={`manage-visit-row-${lead.id}`} className="border-b border-earth-100 hover:bg-earth-50">
                  <td className="p-3 text-olive-900 font-medium">{lead.name}</td>
                  <td className="p-3 text-olive-700">{lead.product}</td>
                  <td className="p-3 text-olive-700">{lead.visit?.date}</td>
                  <td className="p-3 text-olive-700">{lead.visit?.timeSlot}</td>
                  <td className="p-3 text-olive-700">
                    {guides.find(g => g.id === lead.visit?.guideId)?.name || lead.visit?.guideId}
                  </td>
                  <td className="p-3">
                    <div className="flex items-center gap-1.5">
                      <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${
                        lead.visit?.confirmed ? 'bg-olive-100 text-olive-700' : 'bg-amber-100 text-amber-700'
                      }`}>
                        {stageLabels[lead.stage] || lead.stage}
                      </span>
                      {lead.visit?.feedback && (
                        <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${FEEDBACK_COLORS[lead.visit.feedback.result] || 'bg-earth-200 text-earth-700'}`}>
                          {FEEDBACK_LABELS[lead.visit.feedback.result] || lead.visit.feedback.result}
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="p-3">
                    <button
                      id={`manage-visit-reassign-${lead.id}`}
                      onClick={() => setReassignLead(lead)}
                      className="bg-olive-500 text-white px-3 py-1.5 rounded-lg text-xs font-semibold hover:bg-olive-700 min-h-[32px]"
                    >
                      Realocar
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Reassign modal */}
      {reassignLead && (
        <VisitScheduler
          lead={reassignLead}
          onClose={() => setReassignLead(null)}
          onBook={(leadId, visit) => handleReassign(leadId, visit)}
          reassignMode={{
            visitId: reassignLead.visit?.visitId || '',
            currentDate: reassignLead.visit?.date || '',
            currentTimeSlot: reassignLead.visit?.timeSlot || '',
            currentGuideId: reassignLead.visit?.guideId || '',
          }}
        />
      )}
    </div>
  );
}
