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
	collection *mongo.Collection, location entity.Location, quantity int,
	username, userID, fullname, productID, paymentMethod string) (*mongo.InsertOneResult, string, error) {

	ctx := context.Background()

	productsCollection := db.GetCollection(collection.Database(), "products")

	product, err := FindOneProduct(productsCollection, productID)
	if err != nil {
		return nil, "", err
	}

	order := entity.Order{
		ID:               primitive.NewObjectIDFromTimestamp(time.Now()).String()[10:34],
		Username:         username,
		FullName:         fullname,
		DeliveryLocation: location,
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

	usersCollection := db.GetCollection(collection.Database(), "users")

	// Add order to the orders in the user collection
	_, err = AddOrderToUser(usersCollection, userID, order)
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

func GetOrdersWithUsername(collection *mongo.Collection, username string, limit, offset int) ([]entity.Order, int64, error) {
	ctx := context.Background()
	var orders = []entity.Order{}
	var order = entity.Order{}

	filter := bson.M{"username": username}

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

func DeliverOrder(collection *mongo.Collection, orderID, username string) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	filter := bson.M{"$and": []bson.M{
		{
			"_id": orderID,
		},
		{
			"username": username,
		},
	}}

	order, err := GetSingleOrder(collection, orderID)
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()

	order.IsDelivered = true
	order.TimeDelivered = currentTime

	_, err = collection.ReplaceOne(ctx, filter, order)
	if err != nil {
		return nil, err
	}

	usersCollection := db.GetCollection(collection.Database(), "users")

	user, err := GetUser(usersCollection, username)
	if err != nil {
		return nil, err
	}

	for i := range user.Orders {
		if user.Orders[i].ID == orderID {
			user.Orders[i].IsDelivered = true
			user.Orders[i].TimeDelivered = currentTime
			break
		}
	}

	filter = bson.M{"username": username}

	result2, err := usersCollection.ReplaceOne(ctx, filter, user)
	if err != nil {
		return nil, err
	}

	return result2, nil
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

	_, err = collection.ReplaceOne(ctx, filter, order)
	if err != nil {
		return nil, err
	}

	usersCollection := db.GetCollection(collection.Database(), "users")

	user, err := GetUser(usersCollection, username)
	if err != nil {
		return nil, err
	}

	for i := range user.Orders {
		if user.Orders[i].ID == orderID {
			user.Orders[i].IsReceived = true
			break
		}
	}

	filter = bson.M{"username": username}

	result2, err := usersCollection.ReplaceOne(ctx, filter, user)
	if err != nil {
		return nil, err
	}

	return result2, nil

}

func OrderAllCartItems(collection *mongo.Collection,
	username, userID, fullname, paymentMethod string,
	location entity.Location) (int, string, error) {

	cartCollection := db.GetCollection(collection.Database(), "cart")

	cartItems, _, err := GetUserCartItems(cartCollection, username, 0, 0)
	if err != nil {
		return 0, "", err
	}

	productNamesString := ""
	for _, val := range cartItems {
		_, name, err := OrderProductDirectly(collection, location, int(val.Quantity), username, userID, fullname, val.Product.ID, paymentMethod)
		if err != nil {
			return 0, "", err
		}
		productNamesString = productNamesString + " " + name
	}

	return len(cartItems), productNamesString, nil

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

func DeleteAllOrdersWithUsername(collection *mongo.Collection, username string) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{"username": username}

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
