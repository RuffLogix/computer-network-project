.PHONY: backend frontend run

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

run:
	make -j2 backend frontend
