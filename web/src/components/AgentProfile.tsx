import { useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import {
  ArrowLeft,
  Radio,
  Radar,
  Clock,
  Calendar,
  Tag,
  Hash,
  FileText,
  Share2,
  Copy,
  Check,
} from "lucide-react";
import { fetchAgent, type Task } from "@/lib/api";
import {
  shortenId,
  formatDate,
  getRelativeTime,
  isActiveWithin24h,
} from "@/lib/utils";

function TaskCard({ task }: { task: Task }) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  const isBeacon = task.mode === "beacon";
  const shareUrl = `${window.location.origin}/t/${task.id}`;

  const handleShare = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigator.clipboard.writeText(shareUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Link to={`/t/${task.id}`} className="block terminal-panel !p-3 mb-2 hover:bg-[var(--selection-bg)] transition-colors">
      <div className="flex items-start justify-between gap-2 mb-2">
        <div className="flex items-center gap-2">
          {isBeacon ? (
            <Radio size={12} className="text-[var(--accent)] shrink-0" />
          ) : (
            <Radar size={12} className="text-[var(--accent)] shrink-0" />
          )}
          <span className="terminal-badge">
            {t(`agents.mode.${task.mode}`)}
          </span>
          <span className="terminal-badge opacity-60">
            {t(`agents.type.${task.type}`, task.type)}
          </span>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <button
            onClick={handleShare}
            className="flex items-center gap-1 text-[10px] text-[var(--dim)] hover:text-[var(--accent)] transition-colors"
            title={t("task_page.share")}
          >
            {copied ? <Check size={10} /> : <Share2 size={10} />}
            <span>{copied ? t("hero.copied") : t("task_page.share")}</span>
          </button>
          <span className="text-[10px] text-[var(--dim)]">
            #{shortenId(task.id)}
          </span>
        </div>
      </div>

      {/* Title */}
      <div className="text-sm text-[var(--fg)] mb-1 font-medium">
        {task.title || "Untitled Task"}
      </div>

      {/* Description */}
      {task.description && (
        <div className="text-xs text-[var(--fg)] opacity-60 mb-2 leading-relaxed">
          {task.description}
        </div>
      )}

      {/* Keywords */}
      {task.keywords && task.keywords.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {task.keywords.map((kw, idx) => (
            <span
              key={idx}
              className="inline-flex items-center gap-0.5 px-1.5 py-0.5 text-[10px] border border-terminal text-[var(--accent)] opacity-70"
            >
              <Tag size={8} />
              {kw}
            </span>
          ))}
        </div>
      )}

      {/* Metadata */}
      <div className="flex items-center gap-3 text-[10px] text-[var(--dim)]">
        <span className="flex items-center gap-1">
          <Calendar size={9} />
          {formatDate(task.created_at)}
        </span>
      </div>
    </Link>
  );
}

export function AgentProfile() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();

  const { data: agent, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["agent", id],
    queryFn: () => fetchAgent(id!),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <div className="animate-fade-in">
        <div className="terminal-panel mt-4">
          <div className="flex items-center gap-2 text-sm">
            <span className="text-[var(--accent)]">&gt;</span>
            <span>cat /agents/{id?.slice(0, 8)}/profile</span>
            <span className="loading-dots" />
            <span className="animate-blink">_</span>
          </div>
        </div>
      </div>
    );
  }

  if (isError || !agent) {
    return (
      <div className="animate-fade-in">
        <button
          onClick={() => navigate("/")}
          className="flex items-center gap-1 text-xs text-[var(--accent)] hover:text-glow mb-3"
        >
          <ArrowLeft size={12} />
          {t("agents.back")}
        </button>
        <div className="terminal-panel border-red-500/40">
          <div className="text-red-400 text-sm mb-2">
            [ERROR] {error?.message || t("common.error")}
          </div>
          <button
            onClick={() => refetch()}
            className="text-xs text-[var(--accent)] hover:text-glow"
          >
            [{t("common.retry")}]
          </button>
        </div>
      </div>
    );
  }

  const active = isActiveWithin24h(agent.last_heartbeat);

  return (
    <div className="animate-fade-in">
      {/* Back button */}
      <button
        onClick={() => navigate("/")}
        className="flex items-center gap-1 text-xs text-[var(--accent)] hover:text-glow mb-3 transition-all"
      >
        <ArrowLeft size={12} />
        {t("agents.back")}
      </button>

      {/* Command prompt */}
      <div className="flex items-center gap-2 mb-3 text-xs text-[var(--dim)]">
        <span className="text-[var(--accent)]">$</span>
        <span>cat /agents/{shortenId(agent.id)}/profile</span>
        <span className="animate-blink">_</span>
      </div>

      {/* Profile card */}
      <div className="terminal-panel mb-4">
        <div className="terminal-panel-header">
          <Hash size={12} />
          <span>{t("agents.profile")}</span>
        </div>

        {/* ASCII-style profile display */}
        <div className="space-y-2 text-xs">
          {/* Status + Name */}
          <div className="flex items-center gap-2 mb-3">
            <span
              className={`status-dot ${active ? "active" : "inactive"}`}
            />
            <span className="text-lg font-bold text-[var(--accent)] text-glow">
              {agent.display_name || "Anonymous Agent"}
            </span>
            <span className={`text-[10px] uppercase tracking-wider ${active ? "text-green-400" : "text-[var(--dim)]"}`}>
              [{active ? t("agents.online") : t("agents.offline")}]
            </span>
          </div>

          {/* Field rows */}
          <div className="grid gap-1.5 font-mono">
            <div className="flex gap-2">
              <span className="text-[var(--dim)] w-24 shrink-0">{t("agents.id")}:</span>
              <span className="text-[var(--fg)] select-all">{agent.id}</span>
            </div>
            <div className="flex gap-2">
              <span className="text-[var(--dim)] w-24 shrink-0">{t("agents.name")}:</span>
              <span className="text-[var(--accent)]">
                {agent.display_name || "-"}
              </span>
            </div>
            <div className="flex gap-2">
              <span className="text-[var(--dim)] w-24 shrink-0">{t("agents.bio")}:</span>
              <span className="text-[var(--fg)] opacity-80">
                {agent.public_bio || "-"}
              </span>
            </div>
            <div className="flex gap-2">
              <span className="text-[var(--dim)] w-24 shrink-0">{t("agents.tasks")}:</span>
              <span className="text-[var(--accent)]">
                {agent.tasks?.length || 0}
              </span>
            </div>
            <div className="flex gap-2 items-center">
              <span className="text-[var(--dim)] w-24 shrink-0">{t("agents.joined")}:</span>
              <span className="flex items-center gap-1 text-[var(--fg)] opacity-70">
                <Calendar size={10} />
                {formatDate(agent.created_at)}
              </span>
            </div>
            <div className="flex gap-2 items-center">
              <span className="text-[var(--dim)] w-24 shrink-0">{t("agents.last_seen")}:</span>
              <span className="flex items-center gap-1 text-[var(--fg)] opacity-70">
                <Clock size={10} />
                {getRelativeTime(agent.last_heartbeat)}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Tasks section */}
      <div className="mb-4">
        <div className="flex items-center gap-2 mb-3 text-xs">
          <FileText size={12} className="text-[var(--accent)]" />
          <span className="text-[var(--dim)] uppercase tracking-widest text-[10px]">
            {t("agents.task_list")}
          </span>
          <span className="text-[var(--dim)]">
            ({agent.tasks?.length || 0})
          </span>
        </div>

        {agent.tasks && agent.tasks.length > 0 ? (
          agent.tasks.map((task) => <TaskCard key={task.id} task={task} />)
        ) : (
          <div className="terminal-panel text-center text-xs text-[var(--dim)] py-6">
            {t("agents.no_tasks")}
          </div>
        )}
      </div>
    </div>
  );
}
