import { atom } from "nanostores";
import type { SyncSource, SyncStatus } from "../../api/services/sync/client";

export type { SyncSource, SyncStatus } from "../../api/services/sync/client";

// Store for sync sources
export const syncSources = atom<SyncSource[]>([]);

// Store for sync source statuses (keyed by source ID)
export const sourceStatuses = atom<Record<number, SyncStatus>>({});

// Loading states
export const settingsLoading = atom(false);
export const statusesLoading = atom<Record<number, boolean>>({});

// Error states
export const settingsError = atom<string | null>(null);
export const statusesError = atom<Record<number, string | null>>({});

// Actions
export function setSyncSources(sources: SyncSource[]): void {
  syncSources.set(sources);
}

export function setSourceStatus(sourceId: number, status: SyncStatus): void {
  const currentStatuses = sourceStatuses.get();
  sourceStatuses.set({
    ...currentStatuses,
    [sourceId]: status,
  });
}

export function setSettingsLoading(loading: boolean): void {
  settingsLoading.set(loading);
}

export function setStatusLoading(sourceId: number, loading: boolean): void {
  const currentLoading = statusesLoading.get();
  statusesLoading.set({
    ...currentLoading,
    [sourceId]: loading,
  });
}

export function setSettingsError(error: string | null): void {
  settingsError.set(error);
}

export function setStatusError(sourceId: number, error: string | null): void {
  const currentErrors = statusesError.get();
  statusesError.set({
    ...currentErrors,
    [sourceId]: error,
  });
}

export function resetSettings(): void {
  syncSources.set([]);
  sourceStatuses.set({});
  settingsLoading.set(false);
  statusesLoading.set({});
  settingsError.set(null);
  statusesError.set({});
}
