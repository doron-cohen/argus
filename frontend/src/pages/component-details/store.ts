import { atom } from "nanostores";

export interface Component {
  id: string;
  name: string;
  description: string;
  owners: {
    maintainers: string[];
    team: string;
  };
}

export interface CheckReport {
  id: string;
  check_slug: string;
  status:
    | "pass"
    | "fail"
    | "disabled"
    | "skipped"
    | "unknown"
    | "error"
    | "completed";
  timestamp: string;
}

export interface ComponentReportsResponse {
  reports: CheckReport[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    has_more: boolean;
  };
}

export interface ApiError {
  error: string;
  code?: string;
}

// Component details state (co-located with the page)
export const componentDetails = atom<Component | null>(null);
export const loading = atom(false);
export const error = atom<string | null>(null);

// Reports state
export const latestReports = atom<CheckReport[]>([]);
export const reportsLoading = atom(false);
export const reportsError = atom<string | null>(null);

// Actions
export function setComponentDetails(component: Component | null) {
  componentDetails.set(component);
}

export function setLoading(isLoading: boolean) {
  loading.set(isLoading);
}

export function setError(errorMessage: string | null) {
  error.set(errorMessage);
}

export function setLatestReports(reports: CheckReport[]) {
  latestReports.set(reports);
}

export function setReportsLoading(isLoading: boolean) {
  reportsLoading.set(isLoading);
}

export function setReportsError(errorMessage: string | null) {
  reportsError.set(errorMessage);
}

// Reset state
export function resetComponentDetails() {
  componentDetails.set(null);
  loading.set(false);
  error.set(null);
}

export function resetReports() {
  latestReports.set([]);
  reportsLoading.set(false);
  reportsError.set(null);
}
