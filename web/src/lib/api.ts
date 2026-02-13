// ========================================
// Types
// ========================================

export interface Task {
  id: string;
  agent_id: string;
  mode: "beacon" | "radar";
  type: string;
  title: string;
  description: string;
  keywords: string[];
  created_at: string;
  updated_at: string;
}

export interface Agent {
  id: string;
  display_name: string;
  public_bio: string;
  tasks: Task[];
  last_heartbeat: string;
  created_at: string;
  updated_at: string;
}

export interface PaginatedResponse<T> {
  agents: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface Stats {
  total_agents: number;
  active_agents_24h: number;
  total_tasks: number;
  beacon_tasks: number;
  radar_tasks: number;
  total_conversations: number;
  total_matches: number;
  tasks_by_type: Record<string, number>;
}

// ========================================
// API Client
// ========================================

const API_BASE = "";

async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`);
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

export async function fetchAgents(
  page: number = 1,
  perPage: number = 20
): Promise<PaginatedResponse<Agent>> {
  return apiFetch<PaginatedResponse<Agent>>(
    `/api/v1/public/agents?page=${page}&per_page=${perPage}`
  );
}

export async function fetchAgent(id: string): Promise<Agent> {
  const res = await apiFetch<{ agent: Record<string, unknown>; tasks: Task[] }>(
    `/api/v1/public/agents/${id}`
  );
  return { ...res.agent, tasks: res.tasks } as Agent;
}

export async function fetchStats(): Promise<Stats> {
  return apiFetch<Stats>(`/api/v1/public/stats`);
}

export interface PublicTaskDetail {
  task: {
    id: string;
    mode: string;
    type: string;
    title: string;
    created_at: string;
  };
  agent: {
    id: string;
    display_name: string;
    public_bio: string;
  };
}

export async function fetchPublicTask(id: string): Promise<PublicTaskDetail> {
  return apiFetch<PublicTaskDetail>(`/api/v1/public/tasks/${id}`);
}
