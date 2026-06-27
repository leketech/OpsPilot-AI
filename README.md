# OpsPilot-AI
This is an autonomous DevOps agent that investigates production incidents, retrieves historical knowledge, analyzes logs and metrics using Qwen, recommends remediation actions, and safely automates operational workflows with human approval.

## Stack

- Frontend: Next.js 15, TypeScript, Tailwind CSS
- Backend: Go, Fiber
- Database: PostgreSQL with pgvector
- Cache: Redis
- AI: Qwen API
- Containers: Docker
- Deployment: Alibaba Cloud ACK
- CI/CD: GitHub Actions

## Local Development

Start PostgreSQL and Redis:

```bash
docker compose up -d postgres redis
```

Run the backend:

```bash
cd backend
go run ./cmd/server
```

Run the frontend:

```bash
cd frontend
npm install
npm run dev
```

The backend health endpoint is available at `http://localhost:8080/health`.
