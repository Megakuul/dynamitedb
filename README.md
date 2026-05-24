![banner](/banner.svg)
---

Simple ST database engine running entirely on S3.

```go
// define schemas via KeyField and DataField (see supported types below).
type OrderItem struct {
	OrderId     dynamitedb.KeyField          `pk:"order" json:"-"`
	ItemId      dynamitedb.KeyField          `sk:"item" json:"-"`
	Hidden      dynamitedb.DataField[bool]   `json:"hidden"`
	Name        dynamitedb.DataField[string] `json:"name"`
	Description dynamitedb.DataField[string] `json:"description"`
	Count       dynamitedb.DataField[int]    `json:"count"`
	Price       dynamitedb.DataField[int]    `json:"price"`
}

func example() error {
    // create a bucket client
	bucket, err := dynamitedb.New(context.TODO(), "http://127.0.0.1:3900", "test",
		dynamitedb.WithCredentials("access_key", "secret_key"),
		dynamitedb.WithRegion("garage"),
	)
	if err != nil {
		return err
	}

	err = dynamitedb.Create(context.TODO(), bucket, &OrderItem{
		OrderId:     dynamitedb.Key("1"), // order 1
		ItemId:      dynamitedb.Key("3"), // item 3 on order 1
		Name:        dynamitedb.Set("CNC Machine"),
		Description: dynamitedb.Set("The flagship of our store"),
		Count:       dynamitedb.Set(1),
		Price:       dynamitedb.Set(1_000_000),
		Hidden:      dynamitedb.Set(true),
	})
	if err != nil {
		if errors.Is(err, dynamitedb.ErrAlreadyExists) {
			// do something special
		}
		return err
	}

	err = dynamitedb.Update(context.TODO(), bucket, &OrderItem{
		OrderId: dynamitedb.Key("1"),
		ItemId:  dynamitedb.Key("3"),
		Count:   dynamitedb.Mul(2),
		Price:   dynamitedb.Inc(1_000),
		Hidden:  dynamitedb.Toggle(),
	})
	if err != nil {
		return err
	}

	item, err := dynamitedb.Get(context.TODO(), bucket, &OrderItem{
		OrderId: dynamitedb.Key("1"),  // order 1
		ItemId:  dynamitedb.Key("3"),  // item 3 on order 1
		Hidden:  dynamitedb.Eq(false), // must be active
	})
	if err != nil {
		if errors.Is(err, dynamitedb.ErrNotFound) {
			// do something special
		}
		return err
	}

	fmt.Println(item.Name.Value())
	fmt.Println(item.Count.Value())
	fmt.Println(item.Price.Value())
	return nil
}
```


## Concept

DynamiteDB is based on a simplified version of the SingleTable data schema often used in DynamoDB.

Data is organized in objects which are identified by:

- **Partition Key (PK)**: Defines root object type and value e.g. "user/187".
- **Sort Key (SK)**: Defines child object type and value of root e.g. "order/69".

*Notice that this effectively just supports one to many relationships. Other relations must be modelled by denormalizing data!*

> [!TIP]  
> Unlike DynamoDB, DynamiteDB supports lexically sorted partition keys (you can just omit the SK for the base model).
 
> [!WARNING]  
> DynamiteDB only supports lexical ASCENDING sorting!


While DynamiteDB is schemaless, the coding pattern effectively enforces a client side data model.

## Testing

DynamiteDB provides unit tests for important internal reflection functions (like update, filter and serialization).


In addition there are e2e tests for all operations in the `test/` directory.

## Speed Notice

If you are looking for a super fast and optimized database, you are wrong here 👺

While properly designed DynamiteDB schemas scale virtually forever, it will have a base latency dictated by the S3 backend. 
Since S3 uses randomized data, the total overhead (S3 + DynamiteDB) will also always be higher compared to classic databases like Postgres. 

> [!NOTE]  
> Since performance is not a priority, I decided to also sacrifice go performance in favor of usability (using reflection heavy operations).> 
> Currently most go code is very unoptimized since the engine was built in a few days.
