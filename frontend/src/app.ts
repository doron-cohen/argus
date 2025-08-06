import Navigo from "navigo";
import {
  setComponentDetails,
  setLoading,
  setError,
  resetComponentDetails,
  setLatestReports,
  setReportsLoading,
  setReportsError,
  resetReports,
  type ComponentReportsResponse,
} from "./stores/app-store";

// Import and register web components
import "./components/component-list";
import { ComponentDetails } from "./components/component-details";

// Ensure ComponentDetails is registered
if (!customElements.get("component-details")) {
  customElements.define("component-details", ComponentDetails);
}

const router = new Navigo("/");

// Initialize routes
router.on("/", () => {
  showComponentsPage();
});

router.on("/components", () => {
  showComponentsPage();
});

router.on("/components/:id", (match) => {
  if (match && match.data && match.data.id) {
    showComponentDetail(match.data.id);
  }
});

function showComponentsPage() {
  const app = document.getElementById("app");
  if (app) {
    app.innerHTML = `
      <div class="container mx-auto px-4 py-8">
        <div class="mb-8">
          <h1 class="text-3xl font-bold text-gray-900 mb-2" data-testid="page-title">Component Catalog</h1>
          <p class="text-gray-600" data-testid="page-description">Browse and search components in the Argus catalog</p>
        </div>
        
        <div data-testid="components-container">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg leading-6 font-medium text-gray-900" data-testid="components-header">
              Components
            </h3>
          </div>
          
          <component-list></component-list>
        </div>
      </div>
    `;
  }
}

async function showComponentDetail(componentId: string) {
  const app = document.getElementById("app");
  if (app) {
    app.innerHTML = `
      <div class="container mx-auto px-4 py-8">
        <div class="mb-8">
          <h1 class="text-3xl font-bold text-gray-900 mb-2" data-testid="page-title">Component Details</h1>
          <p class="text-gray-600" data-testid="page-description">View detailed information about the component</p>
        </div>
        
        <component-details></component-details>
      </div>
    `;
  }

  // Reset previous state
  resetComponentDetails();
  resetReports();

  // Load component details
  await loadComponentDetails(componentId);
}

async function loadComponentDetails(componentId: string): Promise<void> {
  try {
    setLoading(true);
    setError(null);

    const response = await fetch(
      `/api/catalog/v1/components/${encodeURIComponent(componentId)}`
    );

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Component not found: ${componentId}`);
      }
      const errorData = await response.json();
      throw new Error(
        errorData.error || `HTTP ${response.status}: ${response.statusText}`
      );
    }

    const component = await response.json();
    setComponentDetails(component);

    // Load reports after component details are loaded
    await loadComponentReports(componentId);
  } catch (err) {
    const errorMessage =
      err instanceof Error ? err.message : "Failed to fetch component details";
    setError(errorMessage);
    console.error("Error fetching component details:", err);
  } finally {
    setLoading(false);
  }
}

async function loadComponentReports(componentId: string): Promise<void> {
  try {
    setReportsLoading(true);
    setReportsError(null);

    const response = await fetch(
      `/api/catalog/v1/components/${encodeURIComponent(
        componentId
      )}/reports?latest_per_check=true`
    );

    if (!response.ok) {
      if (response.status === 404) {
        // Component not found, but we already loaded component details, so this might be an empty state
        setLatestReports([]);
        return;
      }
      const errorData = await response.json();
      throw new Error(
        errorData.error || `HTTP ${response.status}: ${response.statusText}`
      );
    }

    const reportsResponse: ComponentReportsResponse = await response.json();
    setLatestReports(reportsResponse.reports);
  } catch (err) {
    const errorMessage =
      err instanceof Error ? err.message : "Failed to fetch component reports";
    setReportsError(errorMessage);
    console.error("Error fetching component reports:", err);
  } finally {
    setReportsLoading(false);
  }
}

// Start the router
router.resolve();
