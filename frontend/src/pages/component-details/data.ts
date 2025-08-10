import {
  setComponentDetails,
  setLoading,
  setError,
  setLatestReports,
  setReportsLoading,
  setReportsError,
  componentDetails,
  error,
  type ComponentReportsResponse,
} from "./store";
import {
  getComponentById,
  getComponentReports,
} from "../../api/services/components/client";

export async function loadComponentDetails(componentId: string): Promise<void> {
  try {
    setLoading(true);
    setError(null);

    const { status, data } = await getComponentById(componentId);
    if (status === 404) {
      throw new Error(`Component not found: ${componentId}`);
    }
    if (status < 200 || status >= 300) {
      const message =
        data && typeof data === "object" && (data as any).error
          ? (data as any).error
          : `HTTP ${status}`;
      throw new Error(message);
    }
    setComponentDetails(data as any);
  } catch (err) {
    const errorMessage =
      err instanceof Error ? err.message : "Failed to fetch component details";
    setError(errorMessage);
    console.error("Error fetching component details:", err);
  } finally {
    setLoading(false);
  }
}

export async function loadComponentReports(componentId: string): Promise<void> {
  try {
    setReportsLoading(true);
    setReportsError(null);

    const { status, data } = await getComponentReports(componentId, {
      latest_per_check: true,
    });

    // Check if component changed while we were fetching (race condition protection)
    const currentComponent = componentDetails.get();
    if (
      !currentComponent ||
      (currentComponent.id !== componentId &&
        currentComponent.name !== componentId)
    ) {
      return; // Component changed, discard this response
    }

    if (status === 404) {
      // Component not found, but we already loaded component details, so this might be an empty state
      setLatestReports([]);
      return;
    }
    if (status < 200 || status >= 300) {
      const message =
        data && typeof data === "object" && (data as any).error
          ? (data as any).error
          : `HTTP ${status}`;
      throw new Error(message);
    }

    const reportsResponse: ComponentReportsResponse = data as any;

    // Double-check component ID before setting reports (additional race condition protection)
    const finalComponent = componentDetails.get();
    if (
      finalComponent &&
      (finalComponent.id === componentId || finalComponent.name === componentId)
    ) {
      setLatestReports(reportsResponse.reports);
    }
  } catch (err) {
    // Only set error if we're still on the same component
    const currentComponent = componentDetails.get();
    if (
      currentComponent &&
      (currentComponent.id === componentId ||
        currentComponent.name === componentId)
    ) {
      const errorMessage =
        err instanceof Error
          ? err.message
          : "Failed to fetch component reports";
      setReportsError(errorMessage);
      console.error("Error fetching component reports:", err);
    }
  } finally {
    setReportsLoading(false);
  }
}
