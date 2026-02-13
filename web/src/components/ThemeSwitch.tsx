import { useState, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Monitor } from "lucide-react";
import { type Theme } from "@/hooks/useTheme";

interface ThemeSwitchProps {
  theme: Theme;
  themes: Theme[];
  setTheme: (t: Theme) => void;
}

const THEME_COLORS: Record<Theme, string> = {
  matrix: "#00ff41",
  amber: "#ffb000",
  dracula: "#bd93f9",
  nord: "#88c0d0",
};

export function ThemeSwitch({ theme, themes, setTheme }: ThemeSwitchProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1.5 px-2 py-1 text-xs border border-terminal hover:border-[var(--border-bright)] transition-colors"
        title={t("theme.title")}
      >
        <Monitor size={12} />
        <span
          className="inline-block w-2 h-2 rounded-full"
          style={{ backgroundColor: THEME_COLORS[theme] }}
        />
        <span className="hidden sm:inline uppercase">{t(`theme.${theme}`)}</span>
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-1 z-50 border border-[var(--border-bright)] bg-[var(--bg)] min-w-[140px]">
          <div className="px-2 py-1 text-[10px] uppercase tracking-widest text-[var(--dim)] border-b border-terminal">
            {t("theme.title")}
          </div>
          {themes.map((th) => (
            <button
              key={th}
              onClick={() => {
                setTheme(th);
                setOpen(false);
              }}
              className={`w-full text-left px-2 py-1.5 text-xs flex items-center gap-2 hover:bg-[var(--selection-bg)] transition-colors ${
                th === theme ? "text-[var(--accent)]" : "text-[var(--fg)]"
              }`}
            >
              <span
                className="inline-block w-2.5 h-2.5 rounded-full border"
                style={{
                  backgroundColor: THEME_COLORS[th],
                  borderColor: THEME_COLORS[th],
                }}
              />
              <span>{t(`theme.${th}`)}</span>
              {th === theme && <span className="ml-auto">*</span>}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
