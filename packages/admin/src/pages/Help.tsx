import { useState } from 'react';
import {
  CaretDown,
  CheckCircle,
  Funnel,
  GearSix,
  ListChecks,
  MapPin,
  Tag,
  UsersThree,
  XCircle,
} from '@phosphor-icons/react';

interface HelpStep {
  title: string;
  body: string;
}

interface HelpSection {
  id: string;
  title: string;
  steps: HelpStep[];
}

interface ProfileHelp {
  id: string;
  label: string;
  icon: typeof UsersThree;
  color: string;
  intro: string;
  sections: HelpSection[];
}

// ── Funnel overview (visible to everyone) ──────────────────────────────

const FUNNEL_STAGES = [
  'Lead',
  'Validado',
  'Visita Agendada',
  'Visita Feita',
  'Contrato',
  'Pago',
  'Reservado',
];

const FUNNEL_TIPS: HelpStep[] = [
  {
    title: 'Mover um lead',
    body: 'No Pipeline, arraste o card para a próxima coluna ou use os botões "→" dentro do card. Cada estágio aceita transições válidas; transições inválidas são bloqueadas com aviso.',
  },
  {
    title: 'Desfazer / refazer',
    body: 'Use Ctrl+Z (desfazer) e Ctrl+Shift+Z (refazer) ou os botões no topo do Pipeline.',
  },
  {
    title: 'WhatsApp integrado',
    body: 'Todo lead tem atalho para o WhatsApp. Na página do lead, o painel de conversa permite enviar mensagens por template (primeiro contato, follow-up, cobrança e confirmação de visita).',
  },
  {
    title: 'Lead cancelado',
    body: 'Qualquer estágio pode cancelar o lead. O lead cancelado vai para a faixa "Finalizados" e não exige mais ações.',
  },
];

// ── Profiles ─────────────────────────────────────────────────────────────

const PROFILES: ProfileHelp[] = [
  {
    id: 'associate',
    label: 'Representante',
    icon: UsersThree,
    color: 'text-earth-700 bg-earth-100',
    intro: 'Você cadastra leads interessados nos produtos da Chácara e acompanha suas comissões.',
    sections: [
      {
        id: 'associate-create',
        title: 'Cadastrar um lead',
        steps: [
          { title: 'Abrir Meus Leads', body: 'Use o menu "Meus Leads".' },
          { title: 'Criar lead', body: 'Clique em "+ Novo Lead" e preencha nome, WhatsApp (com DDD), produto de interesse, data desejada e origem.' },
          { title: 'Confirmar', body: 'Clique em "Cadastrar Lead". Você verá a confirmação no topo da página.' },
          { title: 'E agora?', body: 'Um vendedor assume o lead: ele valida os dados, agenda a visita e conduz até a venda. Não é preciso fazer mais nada.' },
        ],
      },
      {
        id: 'associate-commission',
        title: 'Acompanhar comissão',
        steps: [
          { title: 'Valor a receber', body: 'Em "Meus Leads", o campo "Comissão a receber" soma as comissões pendentes dos leads criados por você.' },
          { title: 'Status', body: 'Cada card mostra o estágio do lead e o status da comissão (pendente / paga).' },
        ],
      },
    ],
  },
  {
    id: 'seller',
    label: 'Vendedor',
    icon: Tag,
    color: 'text-olive-700 bg-olive-100',
    intro: 'Você conduz o lead por todo o funil: validação, visita, contrato e pagamento.',
    sections: [
      {
        id: 'seller-pipeline',
        title: 'Usar o Pipeline',
        steps: [
          { title: 'Visão geral', body: 'O menu "Pipeline" mostra as colunas por estágio, do "Lead" ao "Reservado", com contagem em cada coluna.' },
          { title: 'Mover entre estágios', body: 'Arraste o card para a próxima coluna ou use os botões "→" no card. O sistema valida as transições permitidas.' },
          { title: 'Abrir um lead', body: 'Clique no card para abrir a página do lead com as ações do estágio atual e o WhatsApp ao lado.' },
        ],
      },
      {
        id: 'seller-validate',
        title: 'Validar um lead',
        steps: [
          { title: 'Abrir o lead', body: 'Na fase "Lead", abra o card para ver o formulário de validação.' },
          { title: 'Preencher as abas', body: 'Complete as abas Evento, Contato, Adicionais e Contratante.' },
          { title: 'Confirmar cada campo', body: 'Marque o círculo de confirmação ao lado de cada campo preenchido.' },
          { title: 'Salvar cada grupo', body: 'Clique em "Salvar" em cada aba. Um grupo só conta quando todos os campos estão confirmados.' },
          { title: 'Validar', body: 'Com os 4 grupos salvos, o botão "Validar Lead" aparece no final — clique para promover para "Validado".' },
        ],
      },
      {
        id: 'seller-schedule',
        title: 'Agendar visita técnica',
        steps: [
          { title: 'Abrir o agendamento', body: 'Na fase "Validado", clique em "Agendar Visita".' },
          { title: 'Escolher data e horário', body: 'Navegue por semana (setas ou seletor de data) e escolha o guia e o horário disponível. O número ocupado/capacidade aparece em cada slot.' },
          { title: 'Confirmar', body: 'Confirme a reserva. O guia verá a visita em "Minhas Visitas".' },
        ],
      },
      {
        id: 'seller-contract',
        title: 'Enviar contrato',
        steps: [
          { title: 'Abrir envio', body: 'Na fase "Visita Feita", clique em "Enviar Contrato".' },
          { title: 'Revisar valores', body: 'Confira o produto, os adicionais e o total, e escolha as condições de pagamento.' },
          { title: 'Enviar', body: 'Clique em "Gerar e Enviar Contrato". O cliente recebe para assinar via Clicksign; o lead avança para "Contrato".' },
        ],
      },
      {
        id: 'seller-pix',
        title: 'Gerar cobrança PIX',
        steps: [
          { title: 'Abrir cobrança', body: 'Na fase "Contrato", escolha "Gerar Cobrança PIX".' },
          { title: 'Tipo de cobrança', body: 'Selecione PIX à vista ou PIX parcelado.' },
          { title: 'Enviar ao cliente', body: 'Envie o PIX copia-e-cola ou o link de pagamento pelo WhatsApp.' },
          { title: 'Confirmação', body: 'O pagamento é detectado automaticamente e o lead avança para "Pago".' },
        ],
      },
    ],
  },
  {
    id: 'guide',
    label: 'Guia',
    icon: MapPin,
    color: 'text-green-800 bg-green-100',
    intro: 'Você realiza as visitas técnicas e informa o resultado para o vendedor.',
    sections: [
      {
        id: 'guide-visits',
        title: 'Minhas Visitas',
        steps: [
          { title: 'Abrir a página', body: 'O menu "Minhas Visitas" separa as visitas de Hoje e as Próximas.' },
          { title: 'Confirmar presença', body: 'Ao chegar no local, clique em "Confirmar presença".' },
          { title: 'Registrar resultado', body: 'Após a confirmação, escolha o resultado: Gostou, Não gostou ou Talvez.' },
          { title: 'Enviar feedback', body: 'Adicione notas se quiser e clique em "Enviar feedback". O vendedor recebe e dá sequência à venda.' },
        ],
      },
      {
        id: 'guide-availability',
        title: 'Disponibilidade',
        steps: [
          { title: 'Abrir a página', body: 'Use o menu "Disponibilidade".' },
          { title: 'Marcar horários', body: 'Clique nos horários da grade semanal em que você pode realizar visitas.' },
          { title: 'Copiar / limpar dia', body: 'Use "Copiar" ao lado de um dia para repetir a agenda nos demais, ou "Limpar" para zerar aquele dia.' },
          { title: 'Bloquear períodos', body: 'Em "Datas Indisponíveis", registre períodos em que não atende (ex.: férias).' },
          { title: 'Limite por horário', body: 'Ajuste o máximo de leads por horário e clique em "Salvar".' },
        ],
      },
    ],
  },
  {
    id: 'manager',
    label: 'Administrador',
    icon: GearSix,
    color: 'text-purple-700 bg-purple-100 dark:bg-purple-300/20 dark:text-purple-300',
    intro: 'Você gerencia tudo: pipeline, visitas, relatórios, usuários e backup.',
    sections: [
      {
        id: 'manager-visits',
        title: 'Gerenciar Visitas',
        steps: [
          { title: 'Abrir a página', body: 'O menu "Gerenciar Visitas" lista todas as visitas de todos os guias.' },
          { title: 'Filtrar por guia', body: 'Use o seletor para ver apenas as visitas de um guia.' },
          { title: 'Realocar', body: 'Clique em "Realocar" para trocar guia, data ou horário da visita.' },
        ],
      },
      {
        id: 'manager-reports',
        title: 'Relatórios',
        steps: [
          { title: 'Métricas principais', body: 'Total de leads, taxa de conversão, receita e leads em andamento.' },
          { title: 'Distribuição', body: 'Veja quantos leads estão em cada estágio do funil.' },
        ],
      },
      {
        id: 'manager-config-keys',
        title: 'Configurações — Chaves API',
        steps: [
          { title: 'Abrir aba', body: 'Em "Configurações", aba "Chaves API".' },
          { title: 'Editar', body: 'Configure a Clicksign (contratos) e a OpenPix/Woovi (PIX).' },
          { title: 'Segurança', body: 'Valores aparecem mascarados; digite o novo valor para substituir.' },
        ],
      },
      {
        id: 'manager-config-users',
        title: 'Configurações — Usuários',
        steps: [
          { title: 'Criar usuário', body: 'Em "Configurações", aba "Usuários", informe e-mail, nome e senha.' },
          { title: 'Atribuir perfis', body: 'Marque um ou mais perfis: Representante, Vendedor, Guia e/ou Administrador. Um usuário pode ter vários perfis.' },
          { title: 'Ativar/desativar', body: 'Contas inativas não conseguem fazer login.' },
        ],
      },
      {
        id: 'manager-config-backup',
        title: 'Configurações — Backup',
        steps: [
          { title: 'Abrir aba', body: 'Em "Configurações", aba "Backup".' },
          { title: 'Criar snapshot', body: '"Novo Backup" gera um snapshot local do banco (VACUUM INTO).' },
          { title: 'Enviar arquivos', body: '"Enviar Backup" importa arquivos .sqlite3/.db ou um .zip.' },
          { title: 'Baixar', body: 'Selecione um ou vários backups e clique em "Baixar selecionados".' },
          { title: 'Restaurar', body: '"Carregar como snapshot" substitui o banco atual e reinicia o servidor automaticamente. Use com cuidado — a operação não pode ser desfeita.' },
          { title: 'Histórico', body: 'A seção "Histórico de backups e cargas" registra criações, uploads e restaurações.' },
        ],
      },
    ],
  },
];

const ACCENT_CHIP: Record<string, string> = {
  lead: 'bg-earth-100 text-olive-700',
  validated: 'bg-olive-100 text-olive-700',
  visit_scheduled: 'bg-amber-100 text-amber-800',
  visit_done: 'bg-blue-100 text-blue-700',
  contract: 'bg-terracotta-50 text-terracotta-600',
  paid: 'bg-green-100 text-green-800',
  booked: 'bg-purple-100 text-purple-700 dark:bg-purple-300/20 dark:text-purple-300',
};

function FunnelOverview() {
  return (
    <div id="help-funnel" className="bg-white rounded-2xl border border-earth-200 p-5 mb-6">
      <h2 className="font-semibold text-olive-900 flex items-center gap-2 mb-3">
        <Funnel size={18} weight="fill" className="text-olive-600" />
        Como o funil funciona
      </h2>
      <p className="text-sm text-olive-700 mb-4">
        Todo lead passa pelas etapas abaixo, do cadastro até o evento. Em qualquer etapa o lead pode ser cancelado.
      </p>
      <div className="flex flex-wrap items-center gap-2 mb-5">
        {FUNNEL_STAGES.map((stage, i) => (
          <span key={stage} className="flex items-center gap-2">
            <span className={`px-2.5 py-1 rounded-full text-xs font-semibold ${ACCENT_CHIP[Object.keys(ACCENT_CHIP)[i]]}`}>{stage}</span>
            {i < FUNNEL_STAGES.length - 1 && <span className="text-olive-400">→</span>}
          </span>
        ))}
        <span className="flex items-center gap-2">
          <span className="text-olive-400">→</span>
          <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-olive-500 text-white">Concluído</span>
          <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-red-50 text-red-600">Cancelado</span>
        </span>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {FUNNEL_TIPS.map(tip => (
          <div key={tip.title} className="bg-earth-50 rounded-lg p-3">
            <p className="text-sm font-semibold text-olive-900 flex items-center gap-1.5">
              <CheckCircle size={15} weight="fill" className="text-olive-500" />
              {tip.title}
            </p>
            <p className="text-xs text-olive-600 mt-1">{tip.body}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function Help() {
  const [activeProfile, setActiveProfile] = useState<string>('all');
  const [openSection, setOpenSection] = useState<string | null>('associate-create');

  const profiles = activeProfile === 'all' ? PROFILES : PROFILES.filter(p => p.id === activeProfile);

  return (
    <div id="page-help" className="max-w-4xl mx-auto p-4 md:p-6">
      <h1 id="help-heading" className="font-serif text-2xl italic text-olive-900 mb-1">
        Ajuda
      </h1>
      <p id="help-subtitle" className="text-sm text-olive-600 mb-6">
        Guia de uso do sistema por perfil. Escolha seu perfil para ver os processos principais.
      </p>

      <FunnelOverview />

      {/* Profile selector */}
      <div id="help-profiles" className="flex flex-wrap gap-2 mb-6">
        <button
          onClick={() => setActiveProfile('all')}
          className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
            activeProfile === 'all'
              ? 'bg-olive-600 text-white'
              : 'bg-earth-100 text-olive-700 hover:bg-earth-200'
          }`}
        >
          Todos
        </button>
        {PROFILES.map(p => {
          const Icon = p.icon;
          const active = activeProfile === p.id;
          return (
            <button
              key={p.id}
              id={`help-profile-${p.id}`}
              onClick={() => setActiveProfile(p.id)}
              className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors flex items-center gap-1.5 ${
                active
                  ? 'bg-olive-600 text-white'
                  : 'bg-earth-100 text-olive-700 hover:bg-earth-200'
              }`}
            >
              <Icon size={15} weight="fill" />
              {p.label}
            </button>
          );
        })}
      </div>

      {/* Accordion per profile */}
      {profiles.map(profile => {
        const Icon = profile.icon;
        return (
          <section key={profile.id} id={`help-section-${profile.id}`} className="mb-8">
            <div className="flex items-center gap-2 mb-2">
              <span className={`w-8 h-8 rounded-lg flex items-center justify-center ${profile.color}`}>
                <Icon size={18} weight="fill" />
              </span>
              <h2 className="font-serif text-lg italic text-olive-900">{profile.label}</h2>
            </div>
            <p className="text-sm text-olive-600 mb-3">{profile.intro}</p>

            <div className="space-y-2">
              {profile.sections.map(section => {
                const open = openSection === section.id;
                return (
                  <div key={section.id} className="bg-white border border-earth-200 rounded-lg overflow-hidden">
                    <button
                      id={`help-accordion-${section.id}`}
                      onClick={() => setOpenSection(open ? null : section.id)}
                      className="w-full px-4 py-3 text-sm font-medium text-olive-900 flex items-center justify-between hover:bg-earth-50 transition-colors"
                      aria-expanded={open}
                    >
                      <span className="flex items-center gap-2">
                        <ListChecks size={16} weight="bold" className="text-olive-500" />
                        {section.title}
                      </span>
                      <CaretDown size={16} weight="bold" className={`text-olive-400 transition-transform ${open ? 'rotate-180' : ''}`} />
                    </button>
                    {open && (
                      <ol className="px-4 pb-4 space-y-3 border-t border-earth-100 pt-3">
                        {section.steps.map((step, i) => (
                          <li key={step.title} className="flex gap-3">
                            <span className="flex-shrink-0 w-6 h-6 rounded-full bg-olive-100 text-olive-700 text-xs font-bold flex items-center justify-center">
                              {i + 1}
                            </span>
                            <div className="min-w-0">
                              <p className="text-sm font-semibold text-olive-900">{step.title}</p>
                              <p className="text-xs text-olive-600 mt-0.5">{step.body}</p>
                            </div>
                          </li>
                        ))}
                      </ol>
                    )}
                  </div>
                );
              })}
            </div>
          </section>
        );
      })}

      <div id="help-footer" className="text-xs text-olive-500 border-t border-earth-200 pt-4">
        <XCircle size={13} weight="bold" className="inline -mt-0.5 mr-1 text-terracotta-500" />
        Precisa de mais ajuda? Peça ao administrador.
      </div>
    </div>
  );
}
