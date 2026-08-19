import { useState, useEffect } from 'react';
import { Check, Copy, X } from '@phosphor-icons/react';
import { useAuthContext, fetchWithAuth } from '../hooks/useAuth';

const DAYS = ['mon','tue','wed','thu','fri','sat','sun'];
const DAY_LABELS: Record<string, string> = {
  mon: 'Seg', tue: 'Ter', wed: 'Qua', thu: 'Qui', fri: 'Sex', sat: 'Sáb', sun: 'Dom',
};
const HOURS = ['08:00','09:00','10:00','11:00','12:00','13:00','14:00','15:00','16:00','17:00'];

interface GuideAvailability {
  guide_id: string;
  weekly_schedule: Record<string, string[]>;
  unavailable_dates: { from: string; to: string }[];
  max_per_slot: number;
}

function defaultSchedule(): Record<string, string[]> {
  const s: Record<string, string[]> = {};
  DAYS.forEach(d => { s[d] = []; });
  return s;
}

interface GuideInfo {
  id: string;
  name: string;
}

export default function Availability() {
  const { user } = useAuthContext();
  const [guides, setGuides] = useState<GuideInfo[]>([]);
  const [selectedGuideId, setSelectedGuideId] = useState(user?.uid || 'dev-guide-uid');
  const [av, setAv] = useState<GuideAvailability>({ guide_id: selectedGuideId, weekly_schedule: defaultSchedule(), unavailable_dates: [], max_per_slot: 3 });
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [newFrom, setNewFrom] = useState('');
  const [newTo, setNewTo] = useState('');

  const isManager = user?.roles.includes('manager') ?? false;

  // Fetch guide list for managers
  useEffect(() => {
    if (!isManager) return;
    fetchWithAuth('/api/guides')
      .then(r => r.ok ? r.json() : [])
      .then((data: GuideInfo[]) => {
        if (Array.isArray(data)) setGuides(data);
      })
      .catch(() => {});
  }, [isManager]);

  // Load selected guide's availability
  useEffect(() => {
    setLoading(true);
    setError('');
    fetchWithAuth(`/api/guides/${selectedGuideId}/availability`)
      .then(r => r.ok ? r.json() : null)
      .then(data => {
        if (data) {
          setAv({
            guide_id: data.guide_id || selectedGuideId,
            weekly_schedule: data.weekly_schedule || defaultSchedule(),
            unavailable_dates: data.unavailable_dates || [],
            max_per_slot: data.max_per_slot || 3,
          });
        } else {
          setAv({ guide_id: selectedGuideId, weekly_schedule: defaultSchedule(), unavailable_dates: [], max_per_slot: 3 });
        }
      })
      .catch(() => setError('Erro ao carregar disponibilidade'))
      .finally(() => setLoading(false));
  }, [selectedGuideId]);

  function toggleSlot(day: string, hour: string) {
    setAv(prev => {
      const slots = [...(prev.weekly_schedule[day] || [])];
      const idx = slots.indexOf(hour);
      if (idx >= 0) slots.splice(idx, 1);
      else slots.push(hour);
      slots.sort();
      return { ...prev, weekly_schedule: { ...prev.weekly_schedule, [day]: slots } };
    });
    setSaved(false);
  }

  function copyDay(fromDay: string) {
    setAv(prev => {
      const from = prev.weekly_schedule[fromDay] || [];
      const updated: Record<string, string[]> = {};
      DAYS.forEach(d => { updated[d] = d === fromDay ? from : [...from]; });
      return { ...prev, weekly_schedule: updated };
    });
    setSaved(false);
  }

  function clearDay(day: string) {
    setAv(prev => ({ ...prev, weekly_schedule: { ...prev.weekly_schedule, [day]: [] } }));
    setSaved(false);
  }

  function addUnavailable() {
    if (!newFrom || !newTo) return;
    if (newTo < newFrom) return;
    setAv(prev => ({
      ...prev,
      unavailable_dates: [...prev.unavailable_dates, { from: newFrom, to: newTo }],
    }));
    setNewFrom('');
    setNewTo('');
    setSaved(false);
  }

  function removeUnavailable(idx: number) {
    setAv(prev => ({
      ...prev,
      unavailable_dates: prev.unavailable_dates.filter((_, i) => i !== idx),
    }));
    setSaved(false);
  }

  async function handleSave() {
    setError('');
    try {
      const res = await fetchWithAuth(`/api/guides/${selectedGuideId}/availability`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          weekly_schedule: av.weekly_schedule,
          unavailable_dates: av.unavailable_dates,
          max_per_slot: av.max_per_slot,
        }),
      });
      if (!res.ok) throw new Error('Erro ao salvar');
    } catch {
      setError('Erro ao salvar disponibilidade');
      return;
    }
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  }

  function isToggled(day: string, hour: string): boolean {
    return (av.weekly_schedule[day] || []).includes(hour);
  }

  function slotCount(day: string): number {
    return (av.weekly_schedule[day] || []).length;
  }

  if (loading) return <div className="p-6 text-olive-500 text-sm">Carregando...</div>;

  return (
    <div id="page-availability" className="max-w-5xl mx-auto p-4 md:p-6">
      <h1 id="availability-heading" className="font-serif text-2xl italic text-olive-900 mb-2">Disponibilidade</h1>
      <p className="text-sm text-olive-500 mb-4">Defina seus horários disponíveis para visitas e datas de indisponibilidade.</p>
      {error && <p className="text-sm text-terracotta-500 mb-4 bg-red-50 rounded-lg px-3 py-2">{error}</p>}

      {/* Manager guide selector */}
      {isManager && (
        <div className="mb-4 flex items-center gap-3">
          <label htmlFor="guide-select" className="text-xs text-olive-500">Guia:</label>
          <select id="guide-select" value={selectedGuideId} onChange={e => setSelectedGuideId(e.target.value)}
            className="border border-earth-200 rounded-lg px-3 py-2 text-sm text-olive-700 bg-white min-w-[200px]">
            {guides.map(g => (
              <option key={g.id} value={g.id}>{g.name || g.id}</option>
            ))}
          </select>
        </div>
      )}

      {/* Max per slot */}
      <div className="mb-4 flex items-center gap-3">
        <label className="text-xs text-olive-500">Máximo de leads por horário:</label>
        <input id="availability-max" type="number" min={1} max={10} value={av.max_per_slot}
          onChange={e => { setAv(prev => ({ ...prev, max_per_slot: Number(e.target.value) })); setSaved(false); }}
          className="w-16 border border-earth-200 rounded-lg px-2 py-1 text-sm text-center" />
      </div>

      {/* Weekly schedule grid */}
      <div className="bg-white rounded-2xl border border-earth-200 overflow-x-auto mb-6">
        <table className="w-full text-sm" id="availability-grid">
          <thead>
            <tr className="border-b border-earth-200">
              <th className="p-3 text-left text-olive-500 text-xs font-semibold w-16" />
              {DAYS.map(day => (
                <th key={day} className="p-3 text-center text-olive-700 text-xs font-semibold">
                  <div>{DAY_LABELS[day]}</div>
                  <div className="text-olive-400 font-normal">{slotCount(day)}h</div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {HOURS.map(hour => (
              <tr key={hour} className="border-b border-earth-100 last:border-0">
                <td className="p-2 text-olive-400 text-xs text-center">{hour}</td>
                {DAYS.map(day => {
                  const on = isToggled(day, hour);
                  return (
                    <td key={day} className="p-1">
                      <button
                        id={`slot-${day}-${hour.replace(':','')}`}
                        onClick={() => toggleSlot(day, hour)}
                        className={`w-full py-2 rounded text-xs font-medium transition-colors min-h-[36px] ${
                          on ? 'bg-olive-500 text-white hover:bg-olive-700' : 'bg-earth-100 text-olive-400 hover:bg-earth-200'
                        }`}
                        title={`${DAY_LABELS[day]} ${hour}`}
                      >
                        {on ? <Check weight="bold" size={14} /> : null}
                      </button>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Day actions */}
      <div className="flex flex-wrap gap-2 mb-6">
        {DAYS.map(day => (
          <div key={day} className="flex gap-1">
            <button onClick={() => copyDay(day)}
              className="px-2 py-1 text-xs rounded bg-earth-100 text-olive-700 hover:bg-earth-200 min-h-[32px] flex items-center gap-1"
              title={`Copiar ${DAY_LABELS[day]} para todos`}>
              <Copy size={14} weight="bold" />
              {DAY_LABELS[day]}
            </button>
            <button onClick={() => clearDay(day)}
              className="px-2 py-1 text-xs rounded bg-earth-100 text-terracotta-500 hover:bg-earth-200 min-h-[32px] flex items-center gap-1"
              title={`Limpar ${DAY_LABELS[day]}`}>
              <X size={14} weight="bold" />
            </button>
          </div>
        ))}
      </div>

      {/* Unavailable date ranges */}
      <div className="bg-white rounded-2xl border border-earth-200 p-4 mb-6">
        <h3 className="text-sm font-semibold text-olive-900 mb-3">Datas Indisponíveis</h3>
        {av.unavailable_dates.length > 0 && (
          <div className="space-y-1 mb-3">
            {av.unavailable_dates.map((d, i) => (
              <div key={i} className="flex items-center gap-2 text-sm text-olive-700">
                <span className="bg-red-50 text-terracotta-600 px-2 py-0.5 rounded text-xs">{d.from} → {d.to}</span>
                <button onClick={() => removeUnavailable(i)} className="text-terracotta-400 hover:text-terracotta-600 text-xs p-1" aria-label="Remover período">
                  <X size={14} weight="bold" />
                </button>
              </div>
            ))}
          </div>
        )}
        <div className="flex items-center gap-2 flex-wrap">
          <input id="unavailable-from" type="date" value={newFrom} onChange={e => setNewFrom(e.target.value)}
            className="border border-earth-200 rounded-lg px-3 py-2 text-sm" />
          <span className="text-olive-400 text-sm">até</span>
          <input id="unavailable-to" type="date" value={newTo} onChange={e => setNewTo(e.target.value)}
            className="border border-earth-200 rounded-lg px-3 py-2 text-sm" />
          <button id="unavailable-add" onClick={addUnavailable}
            className="bg-terracotta-500 text-white px-3 py-2 rounded-lg text-sm hover:bg-terracotta-600 min-h-[40px]">
            Bloquear período
          </button>
        </div>
      </div>

      {/* Save */}
      <div className="flex items-center gap-3">
        <button id="availability-save" onClick={handleSave}
          className="bg-olive-500 text-white px-6 py-2.5 rounded-lg text-sm font-semibold hover:bg-olive-700 transition-colors min-h-[44px]">
          Salvar
        </button>
        {saved && <span className="text-sm text-olive-500 flex items-center gap-1"><Check size={16} weight="bold" /> Salvo com sucesso!</span>}
      </div>
    </div>
  );
}
