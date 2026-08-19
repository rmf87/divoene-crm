import { usePipeline, stageLabels } from '../hooks/usePipeline';

export default function Reports() {
  const { allLeads } = usePipeline();

  const stages = ['lead', 'validated', 'visit_scheduled', 'visit_done', 'contract', 'paid', 'booked'];
  const stageCounts = stages.map(s => ({ stage: s, count: allLeads.filter(l => l.stage === s).length }));

  const totalRevenue = allLeads
    .filter(l => l.stage === 'paid' || l.stage === 'booked')
    .reduce((sum, l) => sum + (l.commission ? l.commission.amount * 2 : 0), 0);

  const conversionRate = allLeads.length > 0
    ? Math.round((allLeads.filter(l => ['paid', 'booked', 'completed'].includes(l.stage)).length / allLeads.length) * 100)
    : 0;

  const inProgress = allLeads.filter(l => ['validated', 'visit_scheduled', 'visit_done', 'contract'].includes(l.stage)).length;

  const formatCurrency = (cents: number) =>
    new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL', maximumFractionDigits: 0 }).format(cents / 100);

  return (
    <div id="page-reports" className="max-w-6xl mx-auto p-4 md:p-6">
      <h1 id="reports-heading" className="font-serif text-2xl italic text-olive-900 mb-6">Relatórios</h1>

      <div id="reports-summary" className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <div id="report-stat-total" className="bg-white rounded-2xl p-4 border border-earth-200 text-center">
          <p className="text-3xl font-bold text-olive-900">{allLeads.length}</p>
          <p className="text-sm text-olive-700">Total de leads</p>
        </div>
        <div id="report-stat-conversion" className="bg-white rounded-2xl p-4 border border-earth-200 text-center">
          <p className="text-3xl font-bold text-olive-900">{conversionRate}%</p>
          <p className="text-sm text-olive-700">Conversão</p>
        </div>
        <div id="report-stat-revenue" className="bg-white rounded-2xl p-4 border border-earth-200 text-center">
          <p className="text-3xl font-bold text-terracotta-500">{formatCurrency(totalRevenue)}</p>
          <p className="text-sm text-olive-700">Receita</p>
        </div>
        <div id="report-stat-inprogress" className="bg-white rounded-2xl p-4 border border-earth-200 text-center">
          <p className="text-3xl font-bold text-olive-900">{inProgress}</p>
          <p className="text-sm text-olive-700">Em andamento</p>
        </div>
      </div>

      <div id="reports-pipeline" className="bg-white rounded-2xl p-6 border border-earth-200">
        <h2 className="font-semibold text-olive-900 mb-4">Leads por estágio</h2>
        <div className="space-y-3">
          {stageCounts.map(({ stage, count }) => {
            const pct = allLeads.length > 0 ? (count / allLeads.length) * 100 : 0;
            return (
              <div key={stage} id={`report-stage-${stage}`}>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-olive-700">{stageLabels[stage] || stage}</span>
                  <span className="text-olive-900 font-semibold">{count}</span>
                </div>
                <div className="h-2 bg-earth-100 rounded-full overflow-hidden">
                  <div className="h-full bg-olive-500 rounded-full transition-all" style={{ width: `${pct}%` }} />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
