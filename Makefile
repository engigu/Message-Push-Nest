.PHONY: dev dev-backend dev-frontend

dev:
	@make -j2 dev-backend dev-frontend

dev-backend:
	go tool air

dev-frontend:
	cd web && npm run dev
