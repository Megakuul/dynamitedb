# DynamiteDB 🧨
---

Simple ST database engine running entirely on S3.

```go

client, err := dynamitedb.New(context.TODO(), "https://mys3endpoint.com", "mys3bucket")
if err!=nil {
    return err
}

// TODO
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
