# Computer Network - Project

This repository contains the code for a chat application developed as part of a computer networks course project. The application consists of a backend server written in Go and a frontend application built with Next.js.

Tech Stack:

- **Backend**: Go (Gin, Gorilla WebSocket, MongoDB)
- **Frontend**: Next.js, React, TypeScript, Tailwind CSS

## Backend

To run the backend server, navigate to the `backend` directory and execute the following command:

```bash
cd backend && go run ./cmd/server
```

To generate wire dependencies, use:

```bash
cd backend/cmd/server && wire .
```

## Frontend

To install dependencies for the frontend application, navigate to the `frontend` directory and run:

```bash
cd frontend && pnpm install
```

To run the frontend application, navigate to the `frontend` directory and execute:

```bash
cd frontend && pnpm run dev
```
