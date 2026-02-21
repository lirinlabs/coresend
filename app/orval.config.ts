import { defineConfig } from 'orval';

export default defineConfig({
    coresend: {
        input: '../backend/docs/swagger.yaml',
        output: {
            target: './src/api/generated.ts',
            client: 'react-query',
            httpClient: 'fetch',
            clean: true,
            mode: 'single',
            override: {
                operationName: (operation, route, verb) => {
                    return (
                        operation.operationId ||
                        `${verb}${route.replace(/\//g, '_').replace(/{/g, '').replace(/}/g, '')}`
                    );
                },
                query: {
                    useQuery: true,
                    useInfinite: true,
                    useInfiniteQueryParam: 'nextPage',
                },
            },
        },
    },
});
