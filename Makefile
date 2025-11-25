
generate: mocks

mocks:
	@echo "Generating mocks..."
	go generate ./...
	@echo "Mocks generated successfully"

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)
