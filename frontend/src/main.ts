interface Component {
  id: string;
  name: string;
  description: string;
  owners: {
    maintainers: string[];
    team: string;
  };
}

interface ApiError {
  error: string;
  code?: string;
}

let components: Component[] = [];
let isLoading = true;
let error: string | null = null;

async function fetchComponents(): Promise<void> {
  try {
    isLoading = true;
    error = null;

    const response = await fetch("/api/catalog/v1/components");

    if (!response.ok) {
      const errorData: ApiError = await response.json();
      throw new Error(
        errorData.error || `HTTP ${response.status}: ${response.statusText}`
      );
    }

    components = await response.json();
  } catch (err) {
    error = err instanceof Error ? err.message : "Failed to fetch components";
    console.error("Error fetching components:", err);
  } finally {
    isLoading = false;
    renderComponents();
  }
}

function renderComponents(): void {
  const tbody = document.getElementById("components-tbody");
  const countSpan = document.getElementById("component-count");

  if (!tbody || !countSpan) return;

  // Update component count
  countSpan.textContent = components.length.toString();

  if (isLoading) {
    tbody.innerHTML = `
      <tr>
        <td colspan="5" class="px-6 py-4 text-center">
          <div class="text-sm text-gray-500" data-testid="loading-message">Loading components...</div>
        </td>
      </tr>
    `;
    return;
  }

  if (error) {
    tbody.innerHTML = `
      <tr>
        <td colspan="5" class="px-6 py-4 text-center">
          <div class="text-sm text-red-500" data-testid="error-message">Error: ${error}</div>
        </td>
      </tr>
    `;
    return;
  }

  if (components.length === 0) {
    tbody.innerHTML = `
      <tr>
        <td colspan="5" class="px-6 py-4 text-center">
          <div class="text-sm text-gray-500" data-testid="no-components-message">No components found</div>
        </td>
      </tr>
    `;
    return;
  }

  // Render table rows
  tbody.innerHTML = components
    .map(
      (comp) => `
    <tr class="hover:bg-gray-50" data-testid="component-row">
      <td class="px-6 py-4 whitespace-nowrap">
        <div class="text-sm font-medium text-gray-900" data-testid="component-name">${
          comp.name
        }</div>
      </td>
      <td class="px-6 py-4 whitespace-nowrap">
        <div class="text-sm text-gray-500" data-testid="component-id">${
          comp.id || comp.name
        }</div>
      </td>
      <td class="px-6 py-4">
        <div class="text-sm text-gray-900" data-testid="component-description">${
          comp.description || ""
        }</div>
      </td>
      <td class="px-6 py-4 whitespace-nowrap">
        <div class="text-sm text-gray-500" data-testid="component-team">${
          comp.owners?.team || ""
        }</div>
      </td>
      <td class="px-6 py-4 whitespace-nowrap">
        <div class="text-sm text-gray-500" data-testid="component-maintainers">${
          comp.owners?.maintainers?.join(", ") || ""
        }</div>
      </td>
    </tr>
  `
    )
    .join("");
}

document.addEventListener("DOMContentLoaded", () => {
  // Show loading state immediately
  renderComponents();
  // Then fetch components
  fetchComponents();
});
