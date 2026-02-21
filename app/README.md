# CoreSend Frontend

React-based frontend for CoreSend temporary email service.

## Tech Stack

- **React 19** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **TanStack Query** - Data fetching and caching
- **React Router** - Client-side routing
- **Tailwind CSS 4** - Styling
- **Orval** - API client generation

## Quick Start

```bash
# Install dependencies (using bun)
bun install

# Start development server
bun run dev
```

Frontend runs on `http://localhost:5173`

## Available Scripts

| Command                | Description                      |
| ---------------------- | -------------------------------- |
| `bun run dev`          | Start development server         |
| `bun run build`        | Build for production             |
| `bun run preview`      | Preview production build         |
| `bun run lint`         | Run ESLint                       |
| `bun run generate-api` | Generate API client from swagger |
| `bun run format:all`   | Format all files with Prettier   |

## Project Structure

```
app/
├── src/
│   ├── api/           # Generated API client
│   │   └── generated.ts
│   ├── components/    # Reusable UI components
│   ├── hooks/         # Custom React hooks
│   ├── lib/           # Utilities and helpers
│   ├── pages/         # Page components
│   ├── assets/        # Static assets
│   ├── App.tsx        # Root component
│   └── main.tsx       # Entry point
├── orval.config.ts    # API generation config
├── package.json
└── vite.config.ts
```

## API Client Generation

The API client is auto-generated from the backend's swagger spec:

```bash
bun run generate-api
```

This reads `../backend/docs/swagger.yaml` and generates:

- Type definitions
- React Query hooks
- Fetch client

### Configuration

See `orval.config.ts`:

```typescript
export default defineConfig({
    coresend: {
        input: '../backend/docs/swagger.yaml',
        output: {
            target: './src/api/generated.ts',
            client: 'react-query',
            httpClient: 'fetch',
        },
    },
});
```

## Key Dependencies

### Authentication

- `@noble/ed25519` - Ed25519 cryptography
- `@noble/hashes` - SHA256 hashing
- `@scure/bip39` - Mnemonic generation (for key derivation)

### UI Components

- `@radix-ui/*` - Accessible UI primitives
- `lucide-react` - Icons
- `sonner` - Toast notifications
- `motion` - Animations

### Data Fetching

- `@tanstack/react-query` - Server state management

## Environment Variables

Create `.env` for local development:

```env
VITE_API_URL=http://localhost:8080
```

## Building for Production

```bash
bun run build
```

Output is written to `dist/` directory.

## Development Workflow

1. Start backend (see ../backend/README.md)
2. Generate API client if swagger changed: `bun run generate-api`
3. Start dev server: `bun run dev`
4. Make changes
5. Run lint: `bun run lint`
6. Build: `bun run build`

Run formatting:

```bash
bun run format:all
```
