mongo:
	mongo -u "ecommerce-api" -p "ecommerceapp001" --authenticationDatabase "ecommerce"

run: 
	go run main.go

.PHONY: mongo run