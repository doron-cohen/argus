interface Component {
  id: string;
  name: string;
  description: string;
  team: string;
  maintainers: string[];
}

const components: Component[] = [
  {
    id: "auth-service",
    name: "Authentication Service",
    description:
      "Handles user authentication, authorization, and session management.",
    team: "Security Team",
    maintainers: ["alice.smith", "bob.jones"],
  },
  {
    id: "user-management",
    name: "User Management Service",
    description: "Manages user profiles, roles, and permissions.",
    team: "Platform Team",
    maintainers: ["carol.wilson", "dave.brown"],
  },
  {
    id: "payment-service",
    name: "Payment Processing Service",
    description: "Handles payment processing and billing operations.",
    team: "Finance Team",
    maintainers: ["eve.davis", "frank.miller"],
  },
];

function renderComponents(): void {
  const tbody = document.getElementById("components-tbody");
  const countSpan = document.getElementById("component-count");

  if (!tbody || !countSpan) return;

  // Update component count
  countSpan.textContent = components.length.toString();

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
          comp.id
        }</div>
      </td>
      <td class="px-6 py-4">
        <div class="text-sm text-gray-900" data-testid="component-description">${
          comp.description
        }</div>
      </td>
      <td class="px-6 py-4 whitespace-nowrap">
        <div class="text-sm text-gray-500" data-testid="component-team">${
          comp.team
        }</div>
      </td>
      <td class="px-6 py-4 whitespace-nowrap">
        <div class="text-sm text-gray-500" data-testid="component-maintainers">${comp.maintainers.join(
          ", "
        )}</div>
      </td>
    </tr>
  `
    )
    .join("");
}

document.addEventListener("DOMContentLoaded", renderComponents);
