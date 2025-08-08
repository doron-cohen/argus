export default {
  components: {
    input: '../backend/api/openapi.yaml',
    output: {
      target: './src/api/services/components/client.ts',
      client: 'fetch',
      override: {
        mutator: {
          path: './src/api/fetcher.ts',
          name: 'apiFetch',
        },
      },
    },
  },
} as const;
