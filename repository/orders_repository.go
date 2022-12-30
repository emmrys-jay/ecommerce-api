package repository

import (
	"context"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func OrderProductDirectly(
	collection *mongo.Collection, location *entity.Location, quantity int,
	userID, fullname, productID, paymentMethod string) (*mongo.InsertOneResult, string, error) {

	ctx := context.Background()

	productsCollection := db.GetCollection(collection.Database(), "products")

	product, err := FindOneProduct(productsCollection, productID)
	if err != nil {
		return nil, "", err
	}

	order := entity.Order{
		ID:               primitive.NewObjectIDFromTimestamp(time.Now()).Hex(),
		UserID:           userID,
		FullName:         fullname,
		DeliveryLocation: *location,
		Product:          *product,
		ProductQuantity:  quantity,
		IsDelivered:      false,
		CreatedAt:        time.Now(),
	}

	result, err := collection.InsertOne(ctx, order)
	if err != nil {
		return nil, "", err
	}

	// Update product quantity left and number of orders of that product
	_, err = UpdateProduct(productsCollection, productID, 0.00, -int64(quantity), int64(quantity))
	if err != nil {
		return nil, "", err
	}

	return result, product.Name, nil
}

func GetSingleOrder(collection *mongo.Collection, orderID string) (*entity.Order, error) {
	ctx := context.Background()
	var order entity.Order

	filter := bson.M{"_id": orderID}

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}

	err := result.Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func GetOrdersByUser(collection *mongo.Collection, userID string, limit, offset int) ([]entity.Order, int64, error) {
	ctx := context.Background()
	var orders = []entity.Order{}
	var order = entity.Order{}

	filter := bson.M{"user_id": userID}

	length, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, -1, err
	}

	options := options.Find()
	options.SetLimit(int64(limit))
	options.SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		return nil, -1, err
	}

	for cursor.Next(ctx) {
		err := cursor.Decode(&order)
		if err != nil {
			return nil, -1, err
		}
		orders = append(orders, order)
	}

	return orders, length, nil
}

func GetAllOrders(collection *mongo.Collection, limit, offset int) ([]entity.Order, int64, error) {
	ctx := context.Background()
	var order = entity.Order{}
	var orders = []entity.Order{}

	filter := bson.M{}

	length, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, -1, err
	}

	options := options.Find()
	options.SetLimit(int64(limit))
	options.SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		return nil, -1, err
	}

	for cursor.Next(ctx) {
		err := cursor.Decode(&order)
		if err != nil {
			return nil, -1, err
		}
		orders = append(orders, order)
	}

	return orders, length, nil
}

func DeliverOrder(collection *mongo.Collection, orderID string) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	filter := bson.M{"_id": orderID}

	order, err := GetSingleOrder(collection, orderID)
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()

	order.IsDelivered = true
	order.TimeDelivered = currentTime

	result, err := collection.ReplaceOne(ctx, filter, order)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ReceiveOrder(collection *mongo.Collection, username, orderID string) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	filter := bson.M{
		"$and": []bson.M{
			{
				"username": username,
			},
			{
				"_id": orderID,
			},
		},
	}

	order, err := GetSingleOrder(collection, orderID)
	if err != nil {
		return nil, err
	}

	order.IsReceived = true

	result, err := collection.ReplaceOne(ctx, filter, order)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func OrderAllCartItems(collection *mongo.Collection, userID, fullname, paymentMethod string, location entity.Location) (int, error) {

	cartCollection := db.GetCollection(collection.Database(), "cart")

	cartItems, _, err := GetUserCartItems(cartCollection, userID, 0, 0)
	if err != nil {
		return 0, err
	}

	for _, val := range cartItems {
		_, _, err := OrderProductDirectly(collection, &location, int(val.Quantity), userID, fullname, val.Product.ID, paymentMethod)
		if err != nil {
			return 0, err
		}
	}

	return len(cartItems), nil

}

func DeleteOrder(collection *mongo.Collection, id string) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{"_id": id}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DeleteAllOrdersWithUserID(collection *mongo.Collection, userID string) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{"user_id": userID}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DeleteAllOrders(collection *mongo.Collection) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}
