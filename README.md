# DevPlatform — Internal Developer SaaS Platform

A self-hosted Vercel/Netlify alternative for internal teams. Deploy apps from Git repositories onto Kubernetes with one click.

**Live:** [https://35.175.60.72.sslip.io](https://35.175.60.72.sslip.io)

## Architecture

```
                              ┌─────────────────────────────┐
                              │         Amazon EKS          │
                              │                             │
   ┌───────────┐   /api   ┌──────────┐   git/k8s   ┌──────────────┐
   │  Next.js  │─────────▶│  Go API  │────────────▶│  Deployments │
   │ Dashboard │◀─────────│ Server   │◀────────────│  (K8s)       │
   └───────────┘  HTTPS   └────┬─────┘             └──────────────┘
          NGINX Ingress ──────▶│
     cert-manager (Let's Encrypt)
                              │
                     ┌────────┴────────┐
                     │  RDS PostgreSQL │    │  Redis
                     │  (ElastiCache)  │    │  (ElastiCache)
                     └─────────────────┘    └──────────
```

## Repository Layout

```
.
├── dev-platform/
│   ├── apps/
│   │   ├── api/            # Go control-plane API server
│   │   └── web/            # Next.js dashboard UI
│   └── deployments/        # Kubernetes manifests (kustomize, local dev)
├── kubernetes/
│   ├── helm-chart/devplatform/  # Helm chart deployed to EKS
│   ├── argocd/                  # ArgoCD Application (GitOps)
│   ├── cert-manager/            # ClusterIssuer (Let's Encrypt)
│   └── hpa/                     # HorizontalPodAutoscaler manifests
├── infrastructure/         # Terraform: EKS + RDS + networking
└── .github/workflows/      # CI/CD pipeline
```

## CI/CD Pipeline (GitOps)

On every push to `main`, GitHub Actions:

1. **Test** — `go vet` + build for the API; `npm lint` + `next build` for the web.
2. **Build & Scan** — builds both images (with layer caching), scans them with Trivy (fails on HIGH/CRITICAL), and pushes to GHCR as `:latest` and `:<commit-sha>`.
3. **GitOps** — bumps the image tag in `kubernetes/helm-chart/devplatform/values.yaml` and commits it with `[skip ci]`.
4. **Deploy** — ArgoCD watches the repo and auto-syncs the new tag onto the cluster.

### Register the ArgoCD app once

```bash
kubectl apply -f kubernetes/argocd/application.yaml
```

ArgoCD needs read access to the repo — either keep it public or add credentials
via `argocd repo add https://github.com/2005mohit/Microservices-Platform-on-Kubernetes` with a PAT.

## Infrastructure (Terraform)

Terraform provisions the EKS cluster (v1.35), managed node group, RDS PostgreSQL,
Redis, VPC, and IAM roles. Secrets (`terraform.tfvars`, `deployments/.db.env`)
are gitignored — supply your own.

```bash
cd infrastructure
terraform init
terraform apply
```

Then install the platform components:

```bash
kubectl apply -f kubernetes/cert-manager/cluster-issuer.yaml
helm upgrade --install devplatform kubernetes/helm-chart/devplatform
kubectl apply -f kubernetes/hpa/
```

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

## Local Development

```bash
cd dev-platform/apps/api && make dev      # Go API
cd dev-platform/apps/web && npm run dev   # Next.js UI
```

## License

MIT
