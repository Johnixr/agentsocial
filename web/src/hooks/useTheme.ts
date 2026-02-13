import { useState, useEffect, useCallback } from "react";

export type Theme = "matrix" | "amber" | "dracula" | "nord";

const THEMES: Theme[] = ["matrix", "amber", "dracula", "nord"];
const STORAGE_KEY = "agentsocial-theme";
const DEFAULT_THEME: Theme = "matrix";

function getStoredTheme(): Theme {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored && THEMES.includes(stored as Theme)) {
      return stored as Theme;
    }
  } catch {
    // localStorage not available
  }
  return DEFAULT_THEME;
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(getStoredTheme);

  const applyTheme = useCallback((t: Theme) => {
    document.documentElement.setAttribute("data-theme", t);
  }, []);

  useEffect(() => {
    applyTheme(theme);
  }, [theme, applyTheme]);

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t);
    try {
      localStorage.setItem(STORAGE_KEY, t);
    } catch {
      // localStorage not available
    }
  }, []);

  const cycleTheme = useCallback(() => {
    setThemeState((current) => {
      const idx = THEMES.indexOf(current);
      const next = THEMES[(idx + 1) % THEMES.length];
      try {
        localStorage.setItem(STORAGE_KEY, next);
      } catch {
        // noop
      }
      return next;
    });
  }, []);

  return { theme, setTheme, cycleTheme, themes: THEMES };
}
