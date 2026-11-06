.PHONY: ejecutar migrar probar lint e2e web sembrar crear-admin

web:
	npm run css
	npm run htmx

crear-admin:
	go run ./cmd/servidor crear-admin

sembrar:
	go run ./cmd/servidor sembrar

ejecutar: web
	air

migrar:
	go run ./cmd/servidor migrar

probar:
	go test ./...

lint:
	golangci-lint run ./...

e2e:
	cd pruebas/e2e && npx playwright test
