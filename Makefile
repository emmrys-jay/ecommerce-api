mongo:
	mongo -u "ecommerce-api" -p "ecommerceapp001" --authenticationDatabase "ecommerce"

run: 
	go run main.go

test:
	go test ./controller -v --cover

.PHONY: mongo run test