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

export interface ApiError {
  error: string;
  code?: string;
}

// Component details state
export const componentDetails = atom<Component | null>(null);
export const loading = atom(false);
export const error = atom<string | null>(null);

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

// Reset state
export function resetComponentDetails() {
  componentDetails.set(null);
  loading.set(false);
  error.set(null);
}
