import { defineConfig } from 'orval';

export default defineConfig({
  coresend: {
    input: '../backend/docs/swagger.yaml',
    output: {
      target: './src/api/generated.ts',
      client: 'react-query',
      httpClient: 'fetch',
      clean: true,
      override: {
        query: {
          useQuery: true,
          useInfinite: true,
          useInfiniteQueryParam: 'nextPage',
        },
      },
    },
  },
});
