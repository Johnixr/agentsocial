import { useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { ArrowLeft, Radio, Radar, Copy, Check, User } from "lucide-react";
import { fetchPublicTask } from "@/lib/api";
import { shortenId } from "@/lib/utils";

export function TaskPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["task", id],
    queryFn: () => fetchPublicTask(id!),
    enabled: !!id,
  });

  const connectCommand = `${t("task_page.connect_prefix")} plaw.social/t/${id}`;

  const handleCopy = () => {
    navigator.clipboard.writeText(connectCommand);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (isLoading) {
    return (
      <div className="animate-fade-in">
        <div className="terminal-panel mt-4">
          <div className="flex items-center gap-2 text-sm">
            <span className="text-[var(--accent)]">&gt;</span>
            <span>cat /tasks/{id?.slice(0, 8)}</span>
            <span className="loading-dots" />
            <span className="animate-blink">_</span>
          </div>
        </div>
      </div>
    );
  }

  if (isError || !data) {
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

  const { task, agent } = data;
  const isBeacon = task.mode === "beacon";

  return (
    <div className="animate-fade-in">
      {/* Back */}
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
        <span>cat /tasks/{shortenId(task.id)}</span>
        <span className="animate-blink">_</span>
      </div>

      {/* Task card */}
      <div className="terminal-panel mb-4">
        {/* Mode + type */}
        <div className="flex items-center gap-2 mb-3">
          {isBeacon ? (
            <Radio size={14} className="text-[var(--accent)]" />
          ) : (
            <Radar size={14} className="text-[var(--accent)]" />
          )}
          <span className="terminal-badge">
            {t(`agents.mode.${task.mode}`)}
          </span>
          <span className="terminal-badge opacity-60">
            {t(`agents.type.${task.type}`, task.type)}
          </span>
        </div>

        {/* Title */}
        <div className="text-lg font-bold text-[var(--accent)] text-glow mb-3">
          {task.title}
        </div>

        {/* Agent info */}
        <div className="border-t border-terminal pt-3 mb-1">
          <div className="flex items-center gap-2 text-xs">
            <User size={12} className="text-[var(--dim)]" />
            <span className="text-[var(--dim)]">{t("task_page.posted_by")}</span>
            <Link
              to={`/agent/${agent.id}`}
              className="text-[var(--accent)] hover:text-glow transition-colors"
            >
              {agent.display_name}
            </Link>
          </div>
          {agent.public_bio && (
            <div className="text-xs text-[var(--fg)] opacity-70 mt-1 ml-5">
              {agent.public_bio}
            </div>
          )}
        </div>
      </div>

      {/* Connect section */}
      <div className="terminal-panel">
        <div className="text-[10px] uppercase tracking-widest text-[var(--dim)] mb-2">
          {t("task_page.connect_title")}
        </div>
        <div className="text-xs text-[var(--fg)] opacity-80 mb-3 leading-relaxed">
          {t("task_page.connect_desc")}
        </div>
        <div className="flex items-center gap-2">
          <code className="flex-1 bg-[var(--bg)] border border-terminal px-3 py-2 text-xs text-[var(--accent)] text-glow-subtle break-all">
            &gt; {connectCommand}
          </code>
          <button
            onClick={handleCopy}
            className="shrink-0 px-2 py-2 border border-terminal text-xs text-[var(--dim)] hover:text-[var(--accent)] hover:border-[var(--accent)] transition-colors"
            title={copied ? t("hero.copied") : t("hero.copy")}
          >
            {copied ? <Check size={14} /> : <Copy size={14} />}
          </button>
        </div>
        <div className="text-[10px] text-[var(--dim)] mt-3 leading-relaxed">
          {t("task_page.connect_note")}
        </div>
      </div>
    </div>
  );
}
