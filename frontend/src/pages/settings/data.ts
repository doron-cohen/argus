import {
  setSyncSources,
  setSourceStatus,
  setSettingsLoading,
  setStatusLoading,
  setSettingsError,
  setStatusError,
} from "./store";
import {
  getSyncSources,
  getSyncSourceStatus,
} from "../../api/services/sync/client";

export async function loadSyncSources(): Promise<void> {
  try {
    setSettingsLoading(true);
    setSettingsError(null);

    const response = await getSyncSources();
    const statusCode =
      typeof response.status === "number" ? response.status : 200;
    const sourcesData = response.data;

    if (statusCode < 200 || statusCode >= 300) {
      const message =
        sourcesData &&
        typeof sourcesData === "object" &&
        "message" in sourcesData &&
        typeof sourcesData.message === "string"
          ? sourcesData.message
          : `HTTP ${statusCode}`;
      throw new Error(message);
    }

    if (!Array.isArray(sourcesData)) {
      throw new Error("Invalid API response: expected array of sync sources");
    }

    setSyncSources(sourcesData);

    // Load status for each source
    for (const source of sourcesData) {
      if (source.id !== undefined) {
        await loadSourceStatus(source.id);
      }
    }
  } catch (err) {
    const errorMessage =
      err instanceof Error ? err.message : "Failed to fetch sync sources";
    setSettingsError(errorMessage);
    console.error("Error fetching sync sources:", err);
  } finally {
    setSettingsLoading(false);
  }
}

export async function loadSourceStatus(
  sourceId: number | undefined,
): Promise<void> {
  if (sourceId === undefined) {
    return;
  }
  try {
    setStatusLoading(sourceId, true);
    setStatusError(sourceId, null);

    const response = await getSyncSourceStatus(sourceId);
    const statusCode =
      typeof response.status === "number" ? response.status : 200;
    const statusData = response.data;

    if (statusCode < 200 || statusCode >= 300) {
      const message =
        statusData &&
        typeof statusData === "object" &&
        "message" in statusData &&
        typeof statusData.message === "string"
          ? statusData.message
          : `HTTP ${statusCode}`;
      throw new Error(message);
    }

    setSourceStatus(sourceId, statusData);
  } catch (err) {
    const errorMessage =
      err instanceof Error ? err.message : "Failed to fetch source status";
    setStatusError(sourceId, errorMessage);
    console.error(`Error fetching status for source ${sourceId}:`, err);
  } finally {
    setStatusLoading(sourceId, false);
  }
}
