// TypeScript interfaces for API responses

export interface Component {
  id: string;
  name: string;
  description?: string;
  owners?: {
    team?: string;
    maintainers?: string[];
  };
}

export interface SyncSource {
  id: number;
  type: string;
  interval: string;
  config: {
    path: string;
  };
}

export interface SyncStatus {
  sourceId: number;
  status: string;
  lastSync?: {
    status: string;
    timestamp?: string;
  };
  lastError?: string | null;
  componentsCount: number;
  duration?: number | null;
}

export interface ApiError {
  error: string;
  code?: string;
}
