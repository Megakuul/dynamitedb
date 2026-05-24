![banner](/banner.svg)
---

Simple ST database engine running entirely on S3.

```go
// define schemas via KeyField and DataField:
type OrderItem struct {
	OrderId     dynamitedb.KeyField          `pk:"order" json:"-"`
	ItemId      dynamitedb.KeyField          `sk:"item" json:"-"`
	Hidden      dynamitedb.DataField[bool]   `json:"hidden"`
	Name        dynamitedb.DataField[string] `json:"name"`
	Description dynamitedb.DataField[string] `json:"description"`
	Count       dynamitedb.DataField[int]    `json:"count"`
	Price       dynamitedb.DataField[int]    `json:"price"`
}

// do something with it:
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

	fmt.Println(item.Name.Value())  // CNC Machine
	fmt.Println(item.Count.Value()) // 2
	fmt.Println(item.Price.Value()) // 1_001_000
	return nil
}
```

## Schemas

DynamiteDB schemas are defined as basic go structs using `KeyField` and `DataField` interfaces.


Serialization is done transparently by tagging fields with `json:""` tags 💡


### KeyFields

KeyFields are only used to define the ST partition and sort key (for ST concept see [concept](#concept)).

Use the tag `pk` and `sk` to define their respective names. The sort key is optional.

```go
type OrderItem struct {
    OrderId dynamitedb.KeyField `pk:"order" json:"-"` // equivalent to a dynamodb pk ORDER#<id>
    ItemId  dynamitedb.KeyField `sk:"item" json:"-"`  // equivalent to a dynamodb sk ITEM#<id>
}
```

### DataFields

DataFields are used for mutable data. They are generic and only allow a certain set of types defined as `dataConstraint` in [data.go](/data.go):

- `string`
- `int`
- `float64`
- `bool`
- `time.Time`
- `time.Duration`
- `[]string`
- `map[string]string`

Since serialization is handled transparently by a json marshaller you can also use any other marshallable type on the schema.
However, all fields that are not DataFields become immutable after insertion and cannot be changed or filtered!

```go
type OrderItem struct {
    OrderId  dynamitedb.KeyField `pk:"order" json:"-"`
    ItemId   dynamitedb.KeyField `sk:"item" json:"-"`
    StaticId uuid.UUID           `json:"static_id"` // <- this is allowed but immutable and non-filterable
}
```

Nested datafields in structs are also allowed:

```go
type Description struct {
    Title   dynamitedb.DataField[string] `json:"title"`
    Tooltip dynamitedb.DataField[string] `json:"tooltip"`
    Text    dynamitedb.DataField[string] `json:"text"`
}

type OrderItem struct {
    OrderId     dynamitedb.KeyField            `pk:"order" json:"-"`
    ItemId      dynamitedb.KeyField            `sk:"item" json:"-"`
    Description Description                    `json:"description"` // <- this is allowed
    Invalid     []dynamitedb.DataField[string] `json:"invalid"`     // <- this is NOT allowed
}
```


## Operators

For data manipulation and retrieval there are two different types of "operators": `filter` and `update`.

> [!TIP]
> DynamiteDB treats inserts (`Create`/`Put`) and updates (`Update`) exactly the same. 
> Under the hood inserts just default initialize the item first and then apply the update operators.


Operators are designed to write readable queries, updates and inserts. 

If you need more complex checks you can use `CustomFilter` and `CustomUpdate`:

```go
dynamitedb.Get(context.TODO(), bucket, &OrderItem{
    Price: dynamitedb.CustomFilter(func(price int) bool {
        return value%2 == 0
    }), // check if price is even
})
```
*(I'm open to suggestions for introducing new operators for common use cases.)*



All operators are always designed to be read from left to right:

```go 
dynamitedb.Get(context.TODO(), bucket, &OrderItem{
    OrderId:    dynamitedb.Key("1"),
    ItemId:     dynamitedb.Key("3"),
    Price:      dynamitedb.GreaterThan(1337),  // database price is greater than 1337
    Expiration: dynamitedb.Before(time.Now()), // expiration is before now
})
```


> [!WARNING]  
> Using filter operators on update calls like `Create`, `Put`, `Update`, etc will panic.
> Same thing for update operators on filter calls like `Get`, `Query`, etc. 

```go
dynamitedb.Get(context.TODO(), bucket, &OrderItem{
    OrderId: dynamitedb.Key("1"),
    ItemId:  dynamitedb.Key("3"),
    Hidden:  dynamitedb.Set(false), // <- this is incorrect usage on Get() therefore it will panic
})
```

> [!NOTE]  
> All operators contain comments describing whether they are for filtering or updating. 


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


In addition there are integration tests for all operations in the `test/integration/` directory.


## Speed Notice

While properly designed DynamiteDB schemas scale virtually forever, it will have a base latency dictated by the S3 backend. 
Since S3 uses randomized data, the total overhead (S3 + DynamiteDB) will also always be higher compared to classic databases like Postgres. 

> [!NOTE]  
> Since performance is not a priority, I decided to also sacrifice go performance in favor of usability (using reflection heavy operations).
