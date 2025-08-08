export default {
  components: {
    input: './src/api/services/components/openapi.yaml',
    output: {
      target: './src/api/services/components/client.ts',
      client: 'fetch',
      override: {
        baseUrl: '/api/catalog/v1',
      },
    },
  },
} as const;
