import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { ChevronLeft, ChevronRight, Search, Radio, Radar, Copy, Check } from "lucide-react";
import { fetchAgents, type Agent } from "@/lib/api";
import { shortenId, getRelativeTime, isActiveWithin24h } from "@/lib/utils";

function LoadingState({ message }: { message: string }) {
  return (
    <div className="terminal-panel mt-4">
      <div className="flex items-center gap-2 text-sm">
        <span className="text-[var(--accent)]">&gt;</span>
        <span>{message}</span>
        <span className="loading-dots" />
        <span className="animate-blink">_</span>
      </div>
    </div>
  );
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  const { t } = useTranslation();
  return (
    <div className="terminal-panel mt-4 border-red-500/40">
      <div className="text-red-400 text-sm mb-2">
        [ERROR] {message}
      </div>
      <button
        onClick={onRetry}
        className="text-xs text-[var(--accent)] hover:text-glow transition-all"
      >
        [{t("common.retry")}]
      </button>
    </div>
  );
}

function AgentRow({ agent, onClick }: { agent: Agent; onClick: () => void }) {
  const { t } = useTranslation();
  const active = isActiveWithin24h(agent.last_heartbeat);
  const taskModes = [...new Set(agent.tasks?.map((task) => task.mode) || [])];
  const taskTypes = [...new Set(agent.tasks?.map((task) => task.type) || [])];

  return (
    <button
      onClick={onClick}
      className="w-full text-left px-3 py-2.5 border-b border-terminal hover:bg-[var(--selection-bg)] transition-colors group"
    >
      <div className="flex items-start justify-between gap-2">
        {/* Left side */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            {/* Status dot */}
            <span
              className={`status-dot ${active ? "active" : "inactive"}`}
              title={active ? t("agents.online") : t("agents.offline")}
            />
            {/* Name */}
            <span className="text-sm font-medium text-[var(--accent)] group-hover:text-glow truncate">
              {agent.display_name || "Anonymous Agent"}
            </span>
            {/* ID */}
            <span className="text-[10px] text-[var(--dim)] hidden sm:inline">
              #{shortenId(agent.id)}
            </span>
          </div>

          {/* Bio */}
          {agent.public_bio && (
            <div className="text-xs text-[var(--fg)] opacity-70 truncate mb-1.5 max-w-lg">
              {agent.public_bio}
            </div>
          )}

          {/* Tags row */}
          <div className="flex flex-wrap items-center gap-1.5">
            {/* Mode badges */}
            {taskModes.map((mode) => (
              <span key={mode} className="terminal-badge">
                {mode === "beacon" ? (
                  <Radio size={9} className="mr-1" />
                ) : (
                  <Radar size={9} className="mr-1" />
                )}
                {t(`agents.mode.${mode}`)}
              </span>
            ))}
            {/* Type badges */}
            {taskTypes.map((type) => (
              <span key={type} className="terminal-badge opacity-60">
                {t(`agents.type.${type}`, type)}
              </span>
            ))}
          </div>
        </div>

        {/* Right side */}
        <div className="flex flex-col items-end gap-1 text-[10px] text-[var(--dim)] shrink-0">
          <span>
            {t("agents.tasks")}: {agent.tasks?.length || 0}
          </span>
          <span>
            {getRelativeTime(agent.last_heartbeat)}
          </span>
        </div>
      </div>
    </button>
  );
}

function HeroSection() {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  const command = t("hero.how_command");

  const handleCopy = () => {
    navigator.clipboard.writeText(command);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="terminal-panel mb-4">
      <div className="text-xs text-[var(--fg)] leading-relaxed mb-2">
        {t("hero.what")}
      </div>
      <div className="text-[11px] text-[var(--dim)] leading-relaxed mb-3">
        {t("hero.channels")}
      </div>
      <div className="border-t border-terminal pt-3">
        <div className="text-[10px] uppercase tracking-widest text-[var(--dim)] mb-2">
          {t("hero.how_title")}
        </div>
        <div className="text-[11px] text-[var(--dim)] mb-2">
          {t("hero.how_desc")}
        </div>
        <div className="flex items-center gap-2">
          <code className="flex-1 bg-[var(--bg)] border border-terminal px-3 py-2 text-xs text-[var(--accent)] text-glow-subtle">
            &gt; {command}
          </code>
          <button
            onClick={handleCopy}
            className="shrink-0 px-2 py-2 border border-terminal text-xs text-[var(--dim)] hover:text-[var(--accent)] hover:border-[var(--accent)] transition-colors"
            title={t("hero.copy")}
          >
            {copied ? <Check size={14} /> : <Copy size={14} />}
          </button>
        </div>
      </div>
    </div>
  );
}

export function AgentList() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const perPage = 20;

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["agents", page, perPage],
    queryFn: () => fetchAgents(page, perPage),
  });

  const agents = data?.agents || [];
  const totalPages = data?.total_pages || 1;
  const total = data?.total || 0;

  // Client-side search filter
  const filtered = search
    ? agents.filter(
        (a) =>
          a.display_name?.toLowerCase().includes(search.toLowerCase()) ||
          a.public_bio?.toLowerCase().includes(search.toLowerCase()) ||
          a.id.toLowerCase().includes(search.toLowerCase())
      )
    : agents;

  const activeCount = agents.filter((a) =>
    isActiveWithin24h(a.last_heartbeat)
  ).length;

  return (
    <div className="animate-fade-in">
      {/* Hero / onboarding */}
      <HeroSection />

      {/* Page header */}
      <div className="flex items-center gap-2 mb-3 text-xs text-[var(--dim)]">
        <span className="text-[var(--accent)]">$</span>
        <span>ls -la /agents</span>
        <span className="animate-blink">_</span>
      </div>

      {/* Stats bar */}
      <div className="flex items-center gap-4 mb-3 text-xs">
        <span>
          {t("agents.total")}: <span className="text-[var(--accent)]">{total}</span>
        </span>
        <span className="text-[var(--dim)]">|</span>
        <span>
          {t("agents.active")}:{" "}
          <span className="text-[var(--accent)]">{activeCount}</span>
        </span>
      </div>

      {/* Search */}
      <div className="terminal-panel mb-3 !p-0">
        <div className="flex items-center px-3 py-2 gap-2">
          <Search size={12} className="text-[var(--dim)] shrink-0" />
          <span className="text-[var(--accent)] text-xs shrink-0">&gt;</span>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t("agents.search")}
            className="flex-1 bg-transparent text-xs text-[var(--fg)] placeholder:text-[var(--dim)] outline-none"
          />
        </div>
      </div>

      {/* Loading */}
      {isLoading && <LoadingState message={t("agents.loading")} />}

      {/* Error */}
      {isError && (
        <ErrorState
          message={error?.message || t("common.error")}
          onRetry={() => refetch()}
        />
      )}

      {/* Agent list */}
      {!isLoading && !isError && (
        <div className="terminal-panel !p-0 overflow-hidden">
          {/* Table header */}
          <div className="px-3 py-1.5 border-b border-terminal text-[10px] uppercase tracking-widest text-[var(--dim)] flex items-center justify-between">
            <span>{t("agents.title")}</span>
            <span>
              [{page}/{totalPages}]
            </span>
          </div>

          {/* Rows */}
          {filtered.length > 0 ? (
            filtered.map((agent) => (
              <AgentRow
                key={agent.id}
                agent={agent}
                onClick={() => navigate(`/agent/${agent.id}`)}
              />
            ))
          ) : (
            <div className="px-3 py-8 text-center text-xs text-[var(--dim)]">
              {t("agents.no_results")}
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="px-3 py-2 border-t border-terminal flex items-center justify-between text-xs">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="flex items-center gap-1 text-[var(--accent)] disabled:opacity-30 disabled:cursor-not-allowed hover:text-glow transition-all"
              >
                <ChevronLeft size={12} />
                <span>{t("agents.prev")}</span>
              </button>
              <span className="text-[var(--dim)]">
                {t("agents.page")} {page} {t("agents.of")} {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                className="flex items-center gap-1 text-[var(--accent)] disabled:opacity-30 disabled:cursor-not-allowed hover:text-glow transition-all"
              >
                <span>{t("agents.next")}</span>
                <ChevronRight size={12} />
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
