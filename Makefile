start: build
	docker-compose down
	@echo "Docker images closed!"
	docker-compose up --build -d 
	@echo "Docker images built and started!"

stop:
	docker-compose down
	@echo "Docker images closed!"

build:
	@echo "Building app binary..."
	cd ./app tailwindcss -i ./assets/app.css -o ./assets/tw.css --minify
	cd ./app && env GOOS=linux CGO_ENABLED=0 go build -o hyppo .
	@echo "Done!"
