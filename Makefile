start: build_app
	docker-compose down
	@echo "Docker images closed!"
	docker-compose up --build -d 
	@echo "Docker images built and started!"

stop:
	docker-compose down
	@echo "Docker images closed!"

build: build_app
	@echo "Docker images built!"

build_app:
	@echo "Building broker binary..."
	env GOOS=linux CGO_ENABLED=0 go build -o hyppo ./
	@echo "Done!"
