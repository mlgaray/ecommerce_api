# DB Expert Agent Memory

## JSON Package Pattern
- The repository package uses `github.com/json-iterator/go` via a package-level variable in `product_repository.go`:
  ```go
  var json = jsoniter.ConfigCompatibleWithStandardLibrary
  ```
- Do NOT import `encoding/json` in any repository file -- use the `json` variable directly (it's available package-wide).
- This causes a compile error: `json already declared through import of package json`.

## Stored Procedure Migration Numbering
- Functions are in `database/migrations/functions/` with independent versioning.
- As of now, the latest function migration is `000008_create_function_update_order`.
- Table migrations are in `database/migrations/` (latest: `000028_add_delivery_zone_to_orders`).

## PostgreSQL Error Codes Used
- `P0002` (No data found) -> `RecordNotFoundError`
- `P0003` (Raise exception) -> used for custom validation errors
- `23505` (Unique violation) -> `DuplicateRecordError`
- `23503` (FK violation) -> `ValidationError`

## Order Model Delivery Zone Pattern
- Delivery zone data is stored as a snapshot on the orders table (`delivery_zone_id`, `delivery_zone_name`, `delivery_zone_price`).
- In Go, it's represented as `order.DeliveryMethod.DeliveryZones[0]` (a slice with one element on the `DeliveryMethod` struct).
- For list views (GetAll), only `delivery_zone_name` is fetched (lightweight).
- For detail views (GetByID), all three fields (`id`, `name`, `price`) are fetched.
