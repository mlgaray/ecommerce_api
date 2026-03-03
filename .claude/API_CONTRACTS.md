# API Contracts Standard

Estandar de contratos HTTP para mantener consistencia entre endpoints.
Este documento define el patron oficial de request/response y naming.

## Objetivo

- Evitar respuestas inconsistentes entre endpoints.
- Reducir acoplamiento entre capa HTTP y modelos de dominio.
- Facilitar evoluciones del API sin romper clientes.

## Principios

- Nunca devolver modelos de dominio directamente.
- Cada endpoint devuelve un Response DTO explicito.
- Listados incluyen metadata de paginacion.
- Responses con nombres predecibles y consistentes.
- Auth siempre devuelve DTOs tipados (no map generico).

## Estructura de paquetes

Los contratos estan organizados en dos sub-paquetes dentro de `contracts/`:

```
internal/infraestructure/adapters/http/contracts/
├── requests/          # DTOs de entrada (HTTP → Domain)
│   ├── auth.go        # SignInRequest, SignUpRequest
│   ├── category.go    # CategoryCreate/Update/FiltersRequest
│   ├── product.go     # ProductCreate/Update/FiltersRequest
│   ├── order.go       # CreateOrderRequest, OrderFiltersRequest
│   └── shop.go        # ShopUpdateRequest
└── responses/         # DTOs de salida (Domain → HTTP)
    ├── auth.go        # AuthResponse, SignInResponse
    ├── category.go    # CategoryResponse, ListCategoriesResponse
    ├── product.go     # ProductResponse, ListProductsResponse
    ├── order.go       # OrderResponse, ListOrdersResponse
    ├── shop.go        # ShopResponse + 18 helpers anidados
    └── store.go       # StoreResponse (type aliases de Shop)
```

## Patron de Requests

### Conversion: `ToModel()`

Cada request DTO implementa `ToModel()` para convertir a modelo de dominio:

```go
func (r *CategoryRequest) ToModel() *models.Category
func (r *ProductRequest) ToModel() *models.Product
func (r *OrderRequest) ToModel() *models.Order
func (r *ShopRequest) ToModel() *models.Shop
func (r *SignUpRequest) ToUserModel() *models.User
func (r *SignUpRequest) ToShopModel() *models.Shop
```

### Validacion HTTP

Cada request con datos de entrada implementa `Validate() error`:
- Valida campos requeridos, formato, rangos.
- Retorna `*httpErrors.BadRequestError` con mensaje descriptivo.
- NO valida reglas de negocio (eso lo hace `Model.Validate()`).

### Multipart Form Parsing

Categories, products y shops usan multipart para subir imagenes:

```go
NewCategoryCreateRequest(r *http.Request, shopID int) (*CategoryCreateRequest, error)
NewCategoryUpdateRequest(r *http.Request) (*CategoryUpdateRequest, error)
NewProductCreateRequest(r *http.Request, shopID int) (*ProductCreateRequest, error)
NewProductUpdateRequest(r *http.Request) (*ProductUpdateRequest, error)
NewShopUpdateRequest(r *http.Request) (*ShopUpdateRequest, error)
```

- Parsean JSON del campo `data` del form.
- Extraen archivos de imagen del form.
- Products usa campos indexados: `images[0]`, `images[1]`, etc.
- Validacion de imagenes: MIME type (jpeg/jpg/png), tamaño maximo 3MB.

### Conversion a buffers

```go
func (r *CategoryCreateRequest) ToImageBuffer() ([]byte, error)     // imagen unica
func (r *ProductCreateRequest) ToImageBuffers() ([][]byte, error)   // multiples imagenes
func (r *ShopUpdateRequest) ToLogoBuffer() ([]byte, error)          // logo opcional
func (r *ShopUpdateRequest) ToCoverBuffer() ([]byte, error)         // cover opcional
```

### Filters (Query Params)

Cada recurso listable tiene un request de filtros:

```go
NewCategoryFiltersRequest(queryParams map[string][]string) (*CategoryFiltersRequest, error)
NewProductFiltersRequest(queryParams map[string][]string) (*ProductFiltersRequest, error)
NewOrderFiltersRequest(queryParams map[string][]string) (*OrderFiltersRequest, error)
```

- Parsean query params con conversion de tipos.
- Decodifican cursor para paginacion.
- Degradacion graceful: cursors invalidos se tratan como primera pagina.

## Patron de Responses

### Conversion: `FromModel()`

Cada response DTO provee funciones de conversion desde modelos de dominio:

```go
CategoryResponseFromModel(c *models.Category) *CategoryResponse
ProductResponseFromModel(p *models.Product) *ProductResponse
OrderResponseFromModel(order *models.Order) *OrderResponse
ShopResponseFromModel(shop *models.Shop) *ShopResponse
StoreResponseFromModel(store *models.Store) *StoreResponse
```

- Todas son null-safe (retornan `nil` si el input es `nil`).
- Conversion batch: `XResponsesFromModels(items []*models.X) []*XResponse`.

### Constructores de lista

```go
NewListCategoriesResponse(categories, nextCursor, hasMore, totalCount) *ListCategoriesResponse
NewListProductsResponse(products, nextCursor, hasMore, totalCount) *ListProductsResponse
NewListOrdersResponse(orders, nextCursor, hasMore, totalCount) *ListOrdersResponse
```

### Store: Type Aliases

`store.go` reutiliza los tipos de `shop.go` via type aliases:

```go
type StoreImageResponse = ShopImageResponse
type StoreAddressResponse = ShopAddressResponse
type StorePaymentMethodResponse = ShopPaymentMethodResponse
// ... (11 aliases en total)
```

Solo agrega el campo `IsOpen bool` respecto a `ShopResponse`.

### Order: Flattening

`OrderResponseFromModel` aplana `Variants[].Options[]` del dominio en un array plano `SelectedOptions[]` por item, facilitando el consumo del cliente.

## Naming

### Requests

| Recurso   | Create                  | Update                  | Filters                  |
|-----------|-------------------------|-------------------------|--------------------------|
| Auth      | `SignUpRequest`         | —                       | —                        |
| Category  | `CategoryCreateRequest` | `CategoryUpdateRequest` | `CategoryFiltersRequest` |
| Product   | `ProductCreateRequest`  | `ProductUpdateRequest`  | `ProductFiltersRequest`  |
| Order     | `CreateOrderRequest`    | —                       | `OrderFiltersRequest`    |
| Shop      | —                       | `ShopUpdateRequest`     | —                        |

### Responses

- `CreateXResponse { x: XResponse }`
- `UpdateXResponse { x: XResponse }`
- `GetXResponse { x: XResponse }`
- `ListXResponse { xs: []XResponse, next_cursor, has_more, total_count }`

Nota: el nombre del campo en JSON es el recurso en singular/plural
(`order`, `orders`, `product`, `products`, etc).

## Reglas de respuesta

- POST (Create): 201 + CreateXResponse.
- GET (GetByID): 200 + GetXResponse.
- GET (List): 200 + ListXResponse con metadata.
- PUT/PATCH (Update): 200 + UpdateXResponse.
- DELETE: 204 sin body.

## Paginacion

Todos los listados devuelven:

```json
{
  "items_key": [ "..." ],
  "next_cursor": "string",
  "has_more": true,
  "total_count": 123
}
```

Notas:
- `next_cursor` es opcional si no hay mas paginas.
- `total_count` solo se incluye cuando aplique (ej: primera pagina).
- Cursors se decodifican con degradacion graceful en requests.

## Estructura de DTOs

- XResponse representa el recurso publico del API.
- Se permite "flattening" si ayuda a la UX (ej: Order SelectedOptions).
- Campos opcionales deben ser `omitempty`.
- Conversiones son null-safe en ambas direcciones.

## Auth

- `SignInRequest`: email + password → `Validate()` + `ToUser()`
- `SignUpRequest`: user + shop → `Validate()` + `ToUserModel()` + `ToShopModel()`
- `SignInResponse`: `{ auth: { token } }`
- `AuthResponse`: struct con Token (no map generico).

## Validacion de imagenes

Aplicable a categories, products y shops:

- **MIME types validos**: `image/jpeg`, `image/jpg`, `image/png`
- **Tamaño maximo**: 3MB por archivo
- **Categories**: imagen unica opcional
- **Products**: multiples imagenes, al menos una requerida en create
- **Shops**: logo y cover opcionales, tipo validado ("logo" o "cover")

En updates se distingue entre:
- **Imagenes existentes**: validar ID > 0 y URL no vacia.
- **Imagenes nuevas**: validar MIME type y tamaño.

## Checklist para nuevos endpoints

- [ ] Existe Request DTO en `contracts/requests/` con `Validate()` y `ToModel()`.
- [ ] Existe Response DTO en `contracts/responses/` con `FromModel()`.
- [ ] No se devuelve modelo de dominio directamente.
- [ ] Se usa naming consistente (ver tabla arriba).
- [ ] Listados incluyen metadata de paginacion.
- [ ] Imagenes validan MIME type y tamaño.
- [ ] Tests unitarios cubren Validate, ToModel/FromModel, y edge cases.