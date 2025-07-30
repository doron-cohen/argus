export default function () {
  return {
    components: [
      {
        id: "auth-service",
        name: "Authentication Service",
        description:
          "Handles user authentication and authorization for the platform",
        owners: {
          team: "Platform Team",
          maintainers: ["alice@company.com", "bob@company.com"],
        },
      },
      {
        id: "user-service",
        name: "User Management Service",
        description: "Manages user profiles, preferences, and account settings",
        owners: {
          team: "Platform Team",
          maintainers: ["charlie@company.com"],
        },
      },
      {
        id: "payment-service",
        name: "Payment Processing Service",
        description: "Handles payment processing and billing operations",
        owners: {
          team: "Finance Team",
          maintainers: ["diana@company.com", "eve@company.com"],
        },
      },
    ],
    loading: false,
    error: null,
    searchQuery: "",

    async init() {
      // Simulate loading delay for dummy data
      this.loading = true;
      await new Promise((resolve) => setTimeout(resolve, 500));
      this.loading = false;
    },

    get filteredComponents() {
      if (!this.searchQuery.trim()) {
        return this.components;
      }

      const query = this.searchQuery.toLowerCase();
      return this.components.filter(
        (component) =>
          component.name?.toLowerCase().includes(query) ||
          component.id?.toLowerCase().includes(query) ||
          component.description?.toLowerCase().includes(query) ||
          component.owners?.team?.toLowerCase().includes(query)
      );
    },

    formatDate(dateString) {
      if (!dateString) return "N/A";
      return new Date(dateString).toLocaleDateString();
    },

    truncateText(text, maxLength = 100) {
      if (!text) return "";
      return text.length > maxLength
        ? text.substring(0, maxLength - 3) + "..."
        : text;
    },
  };
}
