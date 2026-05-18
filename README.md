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

When should you use it:

- You want to use the SingleTable pattern for your application.
- You want a simple, stateless and portable database engine that only requires an S3 bucket.
- You are fine with higher latency per request (~50ms).

When to avoid it:

- You are building a write heavy application.
- You need low latency (<50ms).
- You are running on limited hardware / budget.


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

## Speed Notice

If you are looking for a super fast and optimized database, you are wrong here 👺

While dynamitedb scales virtually forever, it will have a base latency dictated by the S3 backend. 
Since S3 uses randomized data, the total overhead (S3 + DynamiteDB) will also always be higher compared to classic databases like Postgres. 

Since performance is not a priority, I decided to also sacrifice go performance in favor of usability (using reflection heavy operations). 

Currently most go code is very unoptimized since the engine was built in a few days.
If performance bottlenecks on the go layer, it is possible to drastically optimize many reflection based operations.
