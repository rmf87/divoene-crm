// Shared visit scheduling modal — week view: guides × 7-day grid.
// Fetches real data from GET /api/visits/slots?week=YYYY-WNN (single request).
// Chevrons cycle ±1 week. Calendar picker jumps to a week.
// Supports new bookings (POST) and reassignment (PATCH) via reassignMode prop.
import { useState, useEffect } from 'react';
import { CaretLeft, CaretRight, X } from '@phosphor-icons/react';
import { fetchWithAuth } from '../hooks/useAuth';

interface LeadSlim {
  id: string;
  name: string;
  whatsapp: string;
  product: string;
  stage?: string;
}

interface GuideSlot {
  guide_id: string; guide_name?: string; date: string; time_slot: string; booked: number; max_slots: number;
}

export interface VisitData { guideId: string; guideName: string; date: string; timeSlot: string; visitId: string; }

interface ReassignMode {
  visitId: string;
  currentDate: string;
  currentTimeSlot: string;
  currentGuideId: string;
}

interface Props {
  lead: LeadSlim;
  onBook?: (leadId: string, visit: VisitData) => void;
  onClose: () => void;
  reassignMode?: ReassignMode;
}

const DAYS = ['Dom','Seg','Ter','Qua','Qui','Sex','Sáb'];

// ── ISO week helpers ────────────────────────────────────────────

function getISOWeek(d: Date): { year: number; week: number } {
  const tmp = new Date(d);
  tmp.setHours(0, 0, 0, 0);
  tmp.setDate(tmp.getDate() + 3 - ((tmp.getDay() + 6) % 7));
  const jan1 = new Date(tmp.getFullYear(), 0, 4);
  const week = 1 + Math.round(((tmp.getTime() - jan1.getTime()) / 86400000 - 3 + ((jan1.getDay() + 6) % 7)) / 7);
  return { year: tmp.getFullYear(), week };
}

function isoWeekDates(year: number, week: number): string[] {
  const jan4 = new Date(year, 0, 4, 12, 0, 0);
  const monday = new Date(jan4);
  monday.setDate(jan4.getDate() - ((jan4.getDay() + 6) % 7) + (week - 1) * 7);
  return Array.from({length: 7}, (_, i) => {
    const d = new Date(monday);
    d.setDate(monday.getDate() + i);
    return d.toISOString().split('T')[0];
  });
}

function formatFromTo(dates: string[]) {
  const [from, to] = [dates[0], dates[6]];
  const fmt = (d: string) => {
    const dt = new Date(d + 'T12:00:00');
    return `${dt.getDate().toString().padStart(2,'0')}/${(dt.getMonth()+1).toString().padStart(2,'0')}`;
  };
  return `${fmt(from)} — ${fmt(to)}`;
}

function fmtDayHeader(d: string) {
  const dt = new Date(d + 'T12:00:00');
  return `${DAYS[dt.getDay()]} ${dt.getDate().toString().padStart(2,'0')}/${(dt.getMonth()+1).toString().padStart(2,'0')}`;
}

type CellSlots = Record<string, { timeSlot: string; booked: number; maxSlots: number; available: boolean }>;

export default function VisitScheduler({ lead, onBook, onClose, reassignMode }: Props) {
  const isReassign = !!reassignMode;
  const initDate = reassignMode?.currentDate || new Date().toISOString().split('T')[0];
  const { year: initY, week: initW } = getISOWeek(new Date(initDate + 'T12:00:00'));

  const [year, setYear] = useState(initY);
  const [week, setWeek] = useState(initW);
  const [slots, setSlots] = useState<GuideSlot[]>([]);
  const [selectedSlot, setSelectedSlot] = useState<{ date: string; guideId: string; guideName: string; timeSlot: string } | null>(null);
  const [bookingError, setBookingError] = useState('');
  const [booking, setBooking] = useState(false);

  const dates = isoWeekDates(year, week);
  const weekStr = `${year}-W${week.toString().padStart(2, '0')}`;

  function changeWeek(offset: number) {
    let newWeek = week + offset;
    let newYear = year;
    if (newWeek < 1) { newYear--; newWeek = 52; }
    if (newWeek > 52) { newYear++; newWeek = 1; }
    setYear(newYear);
    setWeek(newWeek);
    setSelectedSlot(null);
  }

  function jumpToWeek(dateStr: string) {
    const { year: y, week: w } = getISOWeek(new Date(dateStr + 'T12:00:00'));
    setYear(y);
    setWeek(w);
    setSelectedSlot(null);
  }

  useEffect(() => {
    let cancelled = false;
    fetchWithAuth(`/api/visits/slots?week=${encodeURIComponent(weekStr)}`)
      .then(r => r.ok ? r.json() as Promise<GuideSlot[]> : Promise.resolve([]))
      .then(data => { if (!cancelled) setSlots(Array.isArray(data) ? data : []); })
      .catch(() => { if (!cancelled) setSlots([]); });
    return () => { cancelled = true; };
  }, [weekStr]);

  const guideSet = new Map<string, string>();
  for (const s of slots) {
    if (!guideSet.has(s.guide_id)) {
      guideSet.set(s.guide_id, s.guide_name || s.guide_id);
    }
  }
  const guideIds = Array.from(guideSet.keys());

  function cellSlots(guideId: string, date: string): CellSlots {
    const cells: CellSlots = {};
    slots.filter(s => s.guide_id === guideId && s.date === date).forEach(s => {
      cells[s.time_slot] = { timeSlot: s.time_slot, booked: s.booked, maxSlots: s.max_slots, available: s.booked < s.max_slots };
    });
    return cells;
  }

  async function confirmBooking() {
    if (!selectedSlot) return;
    const { date, guideId, guideName, timeSlot } = selectedSlot;
    setBookingError('');
    setBooking(true);
    try {
      const res = await fetchWithAuth('/api/visits', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ lead_id: lead.id, lead_name: lead.name, guide_id: guideId, guide_name: guideName, date, time_slot: timeSlot, product: lead.product, whatsapp: lead.whatsapp }),
      });
      if (!res.ok) throw new Error('Falha ao criar visita');
      const visit = await res.json();
      onBook?.(lead.id, { guideId, guideName, date, timeSlot, visitId: visit.id });
    } catch (e: any) {
      setBookingError(e.message || 'Erro ao agendar');
      setBooking(false);
    }
  }

  async function confirmReassign() {
    if (!selectedSlot || !reassignMode) return;
    const { date, guideId, guideName, timeSlot } = selectedSlot;
    setBookingError('');
    setBooking(true);
    try {
      const res = await fetchWithAuth(`/api/visits/${reassignMode.visitId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guide_id: guideId, guide_name: guideName, date, time_slot: timeSlot }),
      });
      if (!res.ok) throw new Error('Falha ao realocar visita');
      onClose();
      // Dispatch via onBook-style callback for reassign
      onBook?.(lead.id, { guideId, guideName, date, timeSlot, visitId: reassignMode.visitId });
    } catch (e: any) {
      setBookingError(e.message || 'Erro ao realocar');
      setBooking(false);
    }
  }

  const title = isReassign ? `Realocar Visita — ${lead.name}` : `Agendar Visita — ${lead.name}`;
  const actionLabel = isReassign ? 'Realocar' : 'Confirmar';
  const currentSlotKey = isReassign ? `${reassignMode!.currentGuideId}-${reassignMode!.currentDate}-${reassignMode!.currentTimeSlot}` : null;

  return (
    <div id="visit-modal-overlay" className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4"
      onClick={onClose}>
      <div id="visit-modal" className="bg-white rounded-2xl shadow-2xl w-full max-w-6xl max-h-[90vh] overflow-y-auto"
        onClick={e => e.stopPropagation()}>
        <div className="sticky top-0 bg-white border-b border-earth-200 px-6 py-4 flex items-center justify-between rounded-t-2xl">
          <h2 className="font-serif text-xl italic text-olive-900">{title}</h2>
          <button id="visit-cancel" onClick={onClose} className="text-olive-400 hover:text-olive-700 p-2" aria-label="Fechar">
            <X size={20} weight="bold" />
          </button>
        </div>
        <div className="p-6 space-y-4">
          <div className="flex items-center justify-center gap-4">
            <button id="visit-nav-prev" onClick={() => changeWeek(-1)} className="px-3 py-1.5 text-sm rounded bg-earth-100 text-olive-700 hover:bg-earth-200 min-h-[40px] flex items-center" aria-label="Semana anterior">
              <CaretLeft size={16} weight="bold" />
            </button>
            <span id="visit-date-display" className="text-lg font-semibold text-olive-900">
              Semana {weekStr} · {formatFromTo(dates)}
            </span>
            <button id="visit-nav-next" onClick={() => changeWeek(1)} className="px-3 py-1.5 text-sm rounded bg-earth-100 text-olive-700 hover:bg-earth-200 min-h-[40px] flex items-center" aria-label="Próxima semana">
              <CaretRight size={16} weight="bold" />
            </button>
          </div>
          <div className="flex justify-center">
            <input id="visit-date-picker" type="date" value={dates[0]} onChange={e => jumpToWeek(e.target.value)}
              className="border border-earth-200 rounded-lg px-3 py-2 text-sm text-olive-700" />
          </div>

          {isReassign && (
            <p className="text-xs text-olive-500 text-center">
              Atual: {reassignMode!.currentDate} às {reassignMode!.currentTimeSlot} com {guideSet.get(reassignMode!.currentGuideId) || reassignMode!.currentGuideId}
            </p>
          )}

          {guideIds.length === 0 ? (
            <p id="visit-no-guides" className="text-sm text-olive-500 text-center py-4">Nenhum guia disponível nesta semana.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm border-collapse">
                <thead>
                  <tr>
                    <th className="p-2 text-left text-olive-400 font-medium w-28 sticky left-0 bg-white">Guia</th>
                    {dates.map(d => (
                      <th key={d} className={`p-2 text-center text-olive-400 font-medium min-w-[100px] ${d === new Date().toISOString().split('T')[0] ? 'bg-olive-50' : ''}`}>
                        {fmtDayHeader(d)}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {guideIds.map(gid => (
                    <tr key={gid} className="border-t border-earth-100">
                      <td className="p-2 text-olive-700 font-semibold truncate max-w-[112px] sticky left-0 bg-white" title={guideSet.get(gid)}>
                        {guideSet.get(gid)}
                      </td>
                      {dates.map(date => {
                        const cells = cellSlots(gid, date);
                        const entries = Object.values(cells).sort((a, b) => a.timeSlot.localeCompare(b.timeSlot));
                        return (
                          <td key={date} className={`p-1 align-top ${date === new Date().toISOString().split('T')[0] ? 'bg-olive-50' : ''}`}>
                            {entries.length === 0 ? (
                              <div className="text-olive-300 text-xs text-center py-3">—</div>
                            ) : (
                              <div className="space-y-1">
                                {entries.map(s => {
                                  const isSelected = selectedSlot?.date === date && selectedSlot?.guideId === gid && selectedSlot?.timeSlot === s.timeSlot;
                                  const isCurrent = isReassign && gid === reassignMode!.currentGuideId && date === reassignMode!.currentDate && s.timeSlot === reassignMode!.currentTimeSlot;
                                  return (
                                    <button
                                      key={s.timeSlot}
                                      id={`visit-slot-${gid}-${date}-${s.timeSlot.replace(':','')}`}
                                      onClick={() => s.available && setSelectedSlot({ date, guideId: gid, guideName: guideSet.get(gid) || gid, timeSlot: s.timeSlot })}
                                      disabled={!s.available}
                                      className={`w-full rounded-lg px-2 py-1.5 text-center transition-colors text-xs ${
                                        isSelected ? 'bg-olive-500 text-white font-semibold ring-2 ring-olive-700' :
                                        isCurrent ? 'bg-amber-100 text-amber-800 border-2 border-amber-400 font-medium' :
                                        s.available ? 'bg-olive-50 text-olive-600 hover:bg-olive-100 border border-olive-200' :
                                        'bg-earth-100 text-olive-300 cursor-not-allowed'
                                      }`}
                                      title={`${s.timeSlot} — ${s.booked}/${s.maxSlots}${isCurrent ? ' (atual)' : ''}`}
                                    >
                                      <div>{s.timeSlot.slice(0,2)}h</div>
                                      <div className="opacity-60">{s.booked}/{s.maxSlots}</div>
                                    </button>
                                  );
                                })}
                              </div>
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
          {bookingError && <p id="visit-booking-error" className="text-sm text-terracotta-500 text-center">{bookingError}</p>}
          <div className="flex justify-end gap-3 pt-2 border-t border-earth-200">
            <button id="visit-cancel-booking" onClick={onClose} className="px-5 py-2.5 rounded-lg text-sm text-olive-500 hover:bg-earth-100 min-h-[44px]">Cancelar</button>
            {selectedSlot && (
              <button id="visit-confirm-booking" onClick={isReassign ? confirmReassign : confirmBooking} disabled={booking}
                className="bg-olive-500 text-white px-6 py-2.5 rounded-lg text-sm font-semibold hover:bg-olive-700 disabled:opacity-50 min-h-[44px]">
                {booking ? (isReassign ? 'Realocando...' : 'Agendando...') : `${actionLabel} — ${selectedSlot.guideName} ${fmtDayHeader(selectedSlot.date)} às ${selectedSlot.timeSlot}`}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
