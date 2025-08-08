export default {
  components: {
    // Use backend spec as the single source of truth
    input: "../backend/api/openapi.yaml",
    output: {
      target: "./src/api/services/components/client.ts",
      client: "fetch",
      override: {
        baseUrl: "/api/catalog/v1",
      },
    },
  },
} as const;
