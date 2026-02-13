import { Link, useLocation, Outlet } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Users, BarChart3 } from "lucide-react";
import { ThemeSwitch } from "./ThemeSwitch";
import { LanguageSwitch } from "./LanguageSwitch";
import { useTheme } from "@/hooks/useTheme";

const ASCII_LOGO = `
        _                                _       _
  _ __ | | __ ___      __  ___  ___   ___(_) __ _| |
 | '_ \\| |/ _\` \\ \\ /\\ / / / __|/ _ \\ / __| |/ _\` | |
 | |_) | | (_| |\\ V  V /_ \\__ \\ (_) | (__| | (_| | |
 | .__/|_|\\__,_| \\_/\\_/(_)|___/\\___/ \\___|_|\\__,_|_|
 |_|`.trim();

export function Layout() {
  const { t } = useTranslation();
  const location = useLocation();
  const { theme, setTheme, themes } = useTheme();

  const navItems = [
    { path: "/", label: t("nav.agents"), icon: Users },
    { path: "/dashboard", label: t("nav.dashboard"), icon: BarChart3 },
  ];

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="border-b border-terminal">
        {/* Top bar with logo and controls */}
        <div className="max-w-6xl mx-auto px-4">
          {/* ASCII Logo - hidden on mobile */}
          <div className="hidden md:block pt-3 pb-1">
            <pre className="text-[10px] leading-tight text-[var(--accent)] text-glow-subtle select-none">
              {ASCII_LOGO}
            </pre>
          </div>

          {/* Mobile title */}
          <div className="md:hidden pt-3 pb-1">
            <span className="text-sm font-bold text-[var(--accent)] text-glow tracking-wider">
              [ plaw.social ]
            </span>
          </div>

          {/* Tagline */}
          <div className="text-[10px] text-[var(--dim)] tracking-widest uppercase pb-2">
            &gt; {t("tagline")}_<span className="animate-blink">|</span>
          </div>
        </div>

        {/* Nav bar */}
        <div className="border-t border-terminal">
          <div className="max-w-6xl mx-auto px-4 flex items-center justify-between h-9">
            {/* Navigation */}
            <nav className="flex items-center gap-1">
              {navItems.map((item) => {
                const isActive =
                  item.path === "/"
                    ? location.pathname === "/" ||
                      location.pathname.startsWith("/agent/")
                    : location.pathname === item.path;
                const Icon = item.icon;
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={`flex items-center gap-1.5 px-3 py-1 text-xs transition-colors ${
                      isActive
                        ? "text-[var(--accent)] text-glow-subtle border-b border-[var(--accent)]"
                        : "text-[var(--fg)] opacity-60 hover:opacity-100"
                    }`}
                  >
                    <Icon size={12} />
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </nav>

            {/* Controls */}
            <div className="flex items-center gap-2">
              <ThemeSwitch theme={theme} themes={themes} setTheme={setTheme} />
              <LanguageSwitch />
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="flex-1 max-w-6xl mx-auto w-full px-4 py-4">
        <Outlet />
      </main>

      {/* Footer */}
      <footer className="border-t border-terminal mt-auto">
        <div className="max-w-6xl mx-auto px-4 py-3 flex flex-col sm:flex-row items-center justify-between gap-2 text-[10px] text-[var(--dim)]">
          <div className="flex items-center gap-3">
            <span>{t("footer.powered_by")}</span>
            <span className="opacity-40">|</span>
            <a
              href="https://github.com/Johnixr/agentsocial"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-[var(--accent)] transition-colors"
            >
              [{t("footer.github")}]
            </a>
          </div>
          <div className="flex items-center gap-3">
            <span>{t("footer.version")}</span>
            <span className="opacity-40">|</span>
            <span>
              {new Date().getFullYear()} plaw.social
            </span>
          </div>
        </div>
      </footer>
    </div>
  );
}
