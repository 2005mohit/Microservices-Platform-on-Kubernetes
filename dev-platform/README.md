# DevPlatform — Internal Developer SaaS Platform

A self-hosted Vercel/Netlify alternative for internal teams. Deploy apps from Git repositories onto Kubernetes with one click.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌──────────────┐
│  Next.js UI     │────▶│  Go API Server   │────▶│  Kubernetes  │
│  (Dashboard)    │◀────│  (Control Plane) │◀────│  (Deployments)│
└─────────────────┘     └───────┬──────────┘     └──────────────┘
                                │
                     ┌──────────┴──────────┐
                     │  PostgreSQL / Redis  │
                     └─────────────────────┘
```

## Quick Start

### Prerequisites
- Go 1.22+
- Node.js 18+
- Docker & Docker Compose
- Kubernetes cluster (kind/minikube for local)
- kubectl

### 1. Start Infrastructure
```bash
cd apps/api
docker-compose up -d
```

### 2. Start Backend
```bash
cd apps/api
make dev
```

### 3. Start Frontend
```bash
cd apps/web
npm install
npm run dev
```

### 4. Open Dashboard
Navigate to [http://localhost:3000](http://localhost:3000)

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/projects | List all projects |
| POST | /api/v1/projects | Create project |
| GET | /api/v1/projects/:id | Get project details |
| DELETE | /api/v1/projects/:id | Delete project |
| POST | /api/v1/projects/:id/deploy | Trigger deployment |
| GET | /api/v1/deployments/:id | Get deployment details |
| GET | /api/v1/projects/:id/deployments | List deployments |
| WS | /ws/deployments/:id | Stream deployment logs |

## Deploy to Kubernetes

```bash
kubectl apply -k deployments/
```

## License
MIT
