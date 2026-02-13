import { useState, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Globe } from "lucide-react";

const LANGUAGES = [
  { code: "en", label: "EN", name: "English" },
  { code: "zh", label: "\u4e2d", name: "\u4e2d\u6587" },
  { code: "ja", label: "\u65e5", name: "\u65e5\u672c\u8a9e" },
];

export function LanguageSwitch() {
  const { i18n } = useTranslation();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const currentLang = LANGUAGES.find((l) => i18n.language.startsWith(l.code)) || LANGUAGES[0];

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
      >
        <Globe size={12} />
        <span>{currentLang.label}</span>
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-1 z-50 border border-[var(--border-bright)] bg-[var(--bg)] min-w-[120px]">
          <div className="px-2 py-1 text-[10px] uppercase tracking-widest text-[var(--dim)] border-b border-terminal">
            Language
          </div>
          {LANGUAGES.map((lang) => (
            <button
              key={lang.code}
              onClick={() => {
                i18n.changeLanguage(lang.code);
                setOpen(false);
              }}
              className={`w-full text-left px-2 py-1.5 text-xs flex items-center gap-2 hover:bg-[var(--selection-bg)] transition-colors ${
                currentLang.code === lang.code
                  ? "text-[var(--accent)]"
                  : "text-[var(--fg)]"
              }`}
            >
              <span className="font-bold w-4">{lang.label}</span>
              <span>{lang.name}</span>
              {currentLang.code === lang.code && <span className="ml-auto">*</span>}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
