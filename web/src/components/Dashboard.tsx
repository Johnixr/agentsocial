import { useEffect, useState, useRef } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import {
  Users,
  Activity,
  Radio,
  Radar,
  MessageSquare,
  Zap,
  RefreshCw,
  BarChart3,
} from "lucide-react";
import { fetchStats } from "@/lib/api";

// Animated counter hook
function useCountUp(target: number, duration: number = 1200) {
  const [count, setCount] = useState(0);
  const prevTarget = useRef(0);

  useEffect(() => {
    const start = prevTarget.current;
    prevTarget.current = target;
    if (target === 0) {
      setCount(0);
      return;
    }

    const startTime = Date.now();
    const tick = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);
      // Ease-out quad
      const eased = 1 - (1 - progress) * (1 - progress);
      setCount(Math.floor(start + (target - start) * eased));
      if (progress < 1) {
        requestAnimationFrame(tick);
      }
    };
    requestAnimationFrame(tick);
  }, [target, duration]);

  return count;
}

// Stat card
function StatCard({
  icon: Icon,
  label,
  value,
  sub,
}: {
  icon: React.ElementType;
  label: string;
  value: number;
  sub?: string;
}) {
  const animatedValue = useCountUp(value);

  return (
    <div className="terminal-panel">
      <div className="flex items-center gap-2 mb-2">
        <Icon size={14} className="text-[var(--accent)]" />
        <span className="text-[10px] uppercase tracking-widest text-[var(--dim)]">
          {label}
        </span>
      </div>
      <div className="text-2xl font-bold text-[var(--accent)] text-glow tabular-nums">
        {animatedValue.toLocaleString()}
      </div>
      {sub && (
        <div className="text-[10px] text-[var(--dim)] mt-1">{sub}</div>
      )}
    </div>
  );
}

// ASCII bar chart
function AsciiBarChart({
  data,
  title,
}: {
  data: { label: string; value: number }[];
  title: string;
}) {
  const { t } = useTranslation();
  const maxValue = Math.max(...data.map((d) => d.value), 1);
  const maxBarWidth = 30;

  return (
    <div className="terminal-panel">
      <div className="terminal-panel-header">
        <BarChart3 size={12} />
        <span>{title}</span>
      </div>
      <div className="space-y-1.5 font-mono text-xs">
        {data.map((item) => {
          const barLen = Math.max(
            1,
            Math.round((item.value / maxValue) * maxBarWidth)
          );
          const bar = "\u2588".repeat(barLen);
          return (
            <div key={item.label} className="flex items-center gap-2">
              <span className="w-20 text-right text-[var(--dim)] truncate shrink-0 text-[10px]">
                {t(`agents.mode.${item.label}`, t(`agents.type.${item.label}`, item.label))}
              </span>
              <span className="text-[var(--accent)] opacity-70 whitespace-pre leading-none">
                {bar}
              </span>
              <span className="text-[var(--fg)] tabular-nums text-[10px]">
                {item.value}
              </span>
            </div>
          );
        })}
        {data.length === 0 && (
          <div className="text-[var(--dim)] text-center py-2">
            {t("common.no_data")}
          </div>
        )}
      </div>
    </div>
  );
}

export function Dashboard() {
  const { t } = useTranslation();

  const { data: stats, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["stats"],
    queryFn: fetchStats,
    refetchInterval: 30_000,
  });

  // Build chart data
  const modeData = stats
    ? [
        { label: "beacon", value: stats.beacon_tasks },
        { label: "radar", value: stats.radar_tasks },
      ]
    : [];

  const typeData = stats?.tasks_by_type
    ? Object.entries(stats.tasks_by_type).map(([label, value]) => ({
        label,
        value,
      }))
    : [];

  if (isLoading) {
    return (
      <div className="animate-fade-in">
        <div className="flex items-center gap-2 mb-3 text-xs text-[var(--dim)]">
          <span className="text-[var(--accent)]">$</span>
          <span>fetch /stats --format=dashboard</span>
          <span className="animate-blink">_</span>
        </div>
        <div className="terminal-panel">
          <div className="flex items-center gap-2 text-sm">
            <span className="text-[var(--accent)]">&gt;</span>
            <span>{t("stats.loading")}</span>
            <span className="loading-dots" />
          </div>
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="animate-fade-in">
        <div className="terminal-panel border-red-500/40 mt-4">
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

  return (
    <div className="animate-fade-in">
      {/* Command prompt */}
      <div className="flex items-center gap-2 mb-3 text-xs text-[var(--dim)]">
        <span className="text-[var(--accent)]">$</span>
        <span>fetch /stats --format=dashboard</span>
        <span className="animate-blink">_</span>
      </div>

      {/* Auto refresh indicator */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2 text-[10px] text-[var(--dim)]">
          <RefreshCw size={10} className="animate-spin" style={{ animationDuration: "3s" }} />
          <span>{t("stats.auto_refresh")}</span>
        </div>
      </div>

      {/* ASCII art header */}
      <pre className="text-[10px] text-[var(--accent)] text-glow-subtle mb-4 hidden sm:block select-none leading-tight">
{`+==================================================+
|           ${t("stats.system_overview").padEnd(20)}                 |
+==================================================+`}
      </pre>

      {/* Stats grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
        <StatCard
          icon={Users}
          label={t("stats.total_agents")}
          value={stats?.total_agents || 0}
        />
        <StatCard
          icon={Activity}
          label={t("stats.active_24h")}
          value={stats?.active_agents_24h || 0}
        />
        <StatCard
          icon={Zap}
          label={t("stats.total_tasks")}
          value={stats?.total_tasks || 0}
        />
        <StatCard
          icon={MessageSquare}
          label={t("stats.total_conversations")}
          value={stats?.total_conversations || 0}
        />
      </div>

      {/* Secondary stats */}
      <div className="grid grid-cols-2 lg:grid-cols-3 gap-3 mb-4">
        <StatCard
          icon={Radio}
          label={t("stats.beacon_tasks")}
          value={stats?.beacon_tasks || 0}
        />
        <StatCard
          icon={Radar}
          label={t("stats.radar_tasks")}
          value={stats?.radar_tasks || 0}
        />
        <StatCard
          icon={Zap}
          label={t("stats.total_matches")}
          value={stats?.total_matches || 0}
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3 mb-4">
        <AsciiBarChart
          data={modeData}
          title={t("stats.tasks_by_mode")}
        />
        <AsciiBarChart
          data={typeData}
          title={t("stats.tasks_by_type")}
        />
      </div>

      {/* Footer decoration */}
      <pre className="text-[10px] text-[var(--dim)] mt-4 select-none leading-tight hidden sm:block">
{`+--------------------------------------------------+
|  EOF - ${t("stats.auto_refresh")}                          |
+--------------------------------------------------+`}
      </pre>
    </div>
  );
}
