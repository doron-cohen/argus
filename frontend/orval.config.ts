export default {
  components: {
    input: "../backend/api/openapi.yaml",
    output: {
      target: "./src/api/services/components/client.ts",
      client: "fetch",
      baseUrl: "/api/catalog/v1",
      override: {
        mutator: {
          path: "./src/api/fetcher.ts",
          name: "apiFetch",
        },
      },
    },
  },
  sync: {
    input: "../backend/sync/api/openapi.yaml",
    output: {
      target: "./src/api/services/sync/client.ts",
      client: "fetch",
      baseUrl: "/api/sync/v1",
      override: {
        mutator: {
          path: "./src/api/fetcher.ts",
          name: "apiFetch",
        },
      },
    },
  },
} as const;
