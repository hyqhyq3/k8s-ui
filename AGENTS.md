# k8s-ui Project Guide

## Project Overview

This is a Kubernetes web management interface project built with **Golang** (backend) and **React** (frontend).

The project is currently in early scaffolding stage with minimal setup.

## Technology Stack

- **Backend**: Golang
- **Frontend**: React
- **Containerization**: Docker
- **Deployment**: Kubernetes

## Project Structure

```
.
├── Dockerfile          # Docker build configuration (currently empty)
├── Makefile            # Build automation commands
├── README.md           # Project description
├── AGENTS.md           # This file - AI agent guidance
└── frontend/           # Frontend React application directory (currently empty)
```

## Build Commands

The project uses `make` for build automation. Available commands:

```bash
# Build Docker image
make build

# Push Docker image (requires $(image) variable)
make push

# Full deployment pipeline: build, push, and deploy to Kubernetes
make deploy
```

**Note**: The `deploy` command depends on Kubernetes manifests located at `deploy/manifests`, which do not exist yet.

## Current Status

This project is in initial scaffolding phase:

- [ ] Backend Golang code not yet implemented
- [ ] Frontend React application not yet initialized
- [ ] Dockerfile is empty
- [ ] Kubernetes deployment manifests not created
- [ ] No testing framework set up

## Development Guidelines

### Code Organization (Planned)

Based on the technology stack, the recommended structure would be:

```
.
├── backend/            # Golang backend API
│   ├── cmd/           # Application entry points
│   ├── internal/      # Internal packages
│   ├── pkg/           # Public packages
│   └── go.mod         # Go module definition
├── frontend/          # React frontend
│   ├── src/           # Source code
│   ├── public/        # Static assets
│   └── package.json   # Node dependencies
├── deploy/            # Kubernetes manifests
├── Dockerfile         # Multi-stage build for both backend and frontend
└── Makefile           # Build automation
```

### Next Steps for Development

1. **Initialize Go module** for backend:
   ```bash
   go mod init github.com/yourusername/k8s-ui
   ```

2. **Initialize React application** in frontend directory:
   ```bash
   cd frontend && npx create-react-app . --template typescript
   # or
   cd frontend && npm create vite@latest . -- --template react-ts
   ```

3. **Create Dockerfile** with multi-stage build for both backend and frontend

4. **Create Kubernetes manifests** in `deploy/manifests/` directory

5. **Set up testing frameworks**:
   - Go testing for backend
   - Jest/React Testing Library for frontend

## Security Considerations

When implementing this project, consider:

- Kubernetes RBAC configuration for API access
- Secure handling of kubeconfig credentials
- Authentication and authorization for web UI
- HTTPS/TLS for all communications
- Container security best practices (non-root user, minimal base images)

## Language

Project documentation and comments are primarily in **Chinese (中文)**.
