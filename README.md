![banner](/banner.svg)
---

Simple ST database engine running entirely on S3.

```bash
go get github.com/megakuul/dynamitedb
```

```go
// define schemas via KeyField and DataField:
type OrderItem struct {
	OrderId     dynamitedb.KeyField          `pk:"order" cbor:"-"`
	ItemId      dynamitedb.KeyField          `sk:"item" cbor:"-"`
    ETag        dynamitedb.ETagField         `etag:"true" cbor:"-"`
	Hidden      dynamitedb.DataField[bool]   `cbor:"hidden,omitempty"`
	Name        dynamitedb.DataField[string] `cbor:"name,omitempty"`
	Description dynamitedb.DataField[string] `cbor:"description,omitempty"`
	Count       dynamitedb.DataField[int]    `cbor:"count,omitempty"`
	Price       dynamitedb.DataField[int]    `cbor:"price,omitempty"`
}

// do something with it:
func example() error {
	// create a bucket client to e.g. RustFS
	bucket, err := dynamitedb.New(context.TODO(), "http://127.0.0.1:9000", "test",
		dynamitedb.WithCredentials("rustfsadmin", "rustfsadmin"),
		dynamitedb.WithRegion("us-east-1"),
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
		Count:   dynamitedb.Multiply(2),
		Price:   dynamitedb.Increment(1_000),
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


Serialization is done transparently by tagging fields with `cbor:""` tags 💡

> [!TIP]
> It is recommended to always use `omitempty`. 
> Otherwise fields that are explicitly unset on deserialization and cannot be checked with e.g. `dynamitedb.Eq("")`.


### KeyFields

KeyFields are only used to define the ST partition and sort key (for ST concept see [concept](#concept)).

Use the tag `pk` and `sk` to define their respective names. The sort key is optional.

```go
type OrderItem struct {
    OrderId dynamitedb.KeyField `pk:"order" cbor:"-"` // equivalent to a dynamodb pk ORDER#<id>
    ItemId  dynamitedb.KeyField `sk:"item" cbor:"-"`  // equivalent to a dynamodb sk ITEM#<id>
}
```

### ETagField

ETagField is automatically populated if you retrieve data and can be set on updates to ensure the object didn't change in the meantime.

```go
deliveryNote, err := dynamitedb.Get(ctx, bucket, &DeliveryNote{
    DeliveryNoteID: dynamitedb.Key("019ffc9b-6bdb-7983-a78f-4b98f55ea7ef"),
})
if err != nil {
    return err
}

time.Sleep(time.Hour)

err = dynamitedb.Delete(ctx, bucket, &DeliveryNote{
    DeliveryNoteID: dynamitedb.Key(deliveryNote.DeliveryNoteID.Value()),
    ETag:           deliveryNote.ETag,
})
if err != nil {
    if errors.Is(err, ErrConcurrencyConflict) {
        // this happens if someone changed the document since you called Get.
    }
    return err
}
```

### DataFields

DataFields are used for mutable data. Because serialization is handled transparently by the `cbor` marshaller you can also just add other fields to your object, however, those fields will become static and cannot be changed/filtered after creation!

```go
type OrderItem struct {
    OrderId  dynamitedb.KeyField `pk:"order" cbor:"-"`
    ItemId   dynamitedb.KeyField `sk:"item" cbor:"-"`
    StaticId uuid.UUID           `cbor:"static_id,omitempty"` // <- this is allowed but immutable and non-filterable
}
```

Nested datafields in structs are also allowed:

```go
type Description struct {
    Title   dynamitedb.DataField[string] `cbor:"title,omitempty"`
    Tooltip dynamitedb.DataField[string] `cbor:"tooltip,omitempty"`
    Text    dynamitedb.DataField[string] `cbor:"text,omitempty"`
}

type OrderItem struct {
    OrderId     dynamitedb.KeyField            `pk:"order" cbor:"-"`
    ItemId      dynamitedb.KeyField            `sk:"item" cbor:"-"`
    Description Description                    `cbor:"description,omitempty"` // <- this is allowed
    Invalid     []dynamitedb.DataField[string] `cbor:"invalid,omitempty"`     // <- this is NOT allowed
    Valid       dynamitedb.DataField[[]Custom] `cbor:"valid,omitempty"`
}
```


## Operators

For data retrieval and manipulation there are two different types of "operators": `filter` and `update`.

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


## Errors

DynamiteDB intentionally hides the underlying S3 errors so they are not relied upon and not part of the public API.


Instead, it provides its own set of sentinel errors in [errors](/errors.go) which can be used to check for certain conditions of the request.

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


## Backend Support

DynamiteDB requires a proper S3 compatible backend that supports `Conditional Locking`.

Some operational backend inspiration:

- [RustFS](https://github.com/rustfs/rustfs) (simple s3 with cool webui)
- [Alarik](https://github.com/achtungsoftware/alarik) (simple s3 with cool webui (alpha))
- [SeaweedFS](https://github.com/seaweedfs/seaweedfs) (battletested and not too complex)
- [Ceph Rados Gateway](https://docs.ceph.com/en/reef/radosgw/) (battletested for massive workloads)

> [!WARNING]
> The `Garage` s3 project cannot be used since they do not support conditional locking! 

## Speed Notice

While properly designed DynamiteDB schemas scale virtually forever, it will have a base latency dictated by the S3 backend. 
Since S3 uses randomized data, the total overhead (S3 + DynamiteDB) will also always be higher compared to classic databases like Postgres. 

> [!NOTE]  
> Since performance is not a priority, I decided to also sacrifice go performance in favor of usability (using reflection heavy operations).
