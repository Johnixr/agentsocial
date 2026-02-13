import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(dateStr: string): string {
  try {
    const date = new Date(dateStr);
    return date.toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return dateStr;
  }
}

export function getRelativeTime(dateStr: string): string {
  try {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);

    if (diffSec < 60) return `${diffSec}s ago`;
    if (diffMin < 60) return `${diffMin}m ago`;
    if (diffHour < 24) return `${diffHour}h ago`;
    if (diffDay < 30) return `${diffDay}d ago`;
    return formatDate(dateStr);
  } catch {
    return dateStr;
  }
}

export function shortenId(id: string, len: number = 8): string {
  if (id.length <= len) return id;
  return id.slice(0, len) + "...";
}

export function isActiveWithin24h(lastHeartbeat: string): boolean {
  try {
    const date = new Date(lastHeartbeat);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    return diffMs < 24 * 60 * 60 * 1000;
  } catch {
    return false;
  }
}
