# Frontend Guide: Metrics Date Filtering

## Overview

All metrics endpoints already support date filtering via query parameters. This guide explains how to implement a period selector (e.g., "This month", "Last month", "Custom range") on the frontend.

## Endpoints

All endpoints require authentication (JWT) and shop ownership. Base path: `/shops/{shop_id}/metrics/`

| Endpoint | Method | Description |
|---|---|---|
| `/shops/{shop_id}/metrics/dashboard` | GET | Full dashboard (revenue, orders, top products, customers) |
| `/shops/{shop_id}/metrics/revenue/trend` | GET | Daily revenue trend (chart data) |
| `/shops/{shop_id}/metrics/products/top` | GET | Top selling products |
| `/shops/{shop_id}/metrics/customers/top` | GET | Top customers by spend |

## Query Parameters

All endpoints (except dashboard) accept the following query params:

| Param | Type | Format | Required | Description |
|---|---|---|---|---|
| `date_from` | string | `YYYY-MM-DD` | No | Start date (inclusive). Defaults to server-side logic if omitted. |
| `date_to` | string | `YYYY-MM-DD` | No | End date (inclusive, covers full day until 23:59:59). Defaults to today if omitted. |
| `tz` | string | IANA timezone | No | Timezone for date calculations (e.g., `America/Argentina/Buenos_Aires`). |
| `limit` | integer | positive int | No | Max number of results (for top products/customers). |
| `sort_by` | string | `quantity` or `revenue` | No | Sort criteria for top products only. |

**Dashboard** only accepts `tz` (it computes its own date ranges internally).

## Validation Errors

| Error | HTTP Status | Cause |
|---|---|---|
| `invalid_date_from_format` | 400 | `date_from` is not `YYYY-MM-DD` |
| `invalid_date_to_format` | 400 | `date_to` is not `YYYY-MM-DD` |
| `invalid_limit_format` | 400 | `limit` is not a valid integer |
| `limit_cannot_be_negative` | 400 | `limit` is negative |

## Usage Examples

### Current month (default behavior)
```
GET /shops/1/metrics/revenue/trend
GET /shops/1/metrics/products/top
GET /shops/1/metrics/customers/top
```

### Last month (February 2026)
```
GET /shops/1/metrics/revenue/trend?date_from=2026-02-01&date_to=2026-02-28
GET /shops/1/metrics/products/top?date_from=2026-02-01&date_to=2026-02-28
GET /shops/1/metrics/customers/top?date_from=2026-02-01&date_to=2026-02-28
```

### Last 7 days
```
GET /shops/1/metrics/revenue/trend?date_from=2026-02-22&date_to=2026-03-01
```

### Today only
```
GET /shops/1/metrics/revenue/trend?date_from=2026-03-01&date_to=2026-03-01
```

### Top 5 products by revenue (last month)
```
GET /shops/1/metrics/products/top?date_from=2026-02-01&date_to=2026-02-28&limit=5&sort_by=revenue
```

### Top 10 customers (last month, with timezone)
```
GET /shops/1/metrics/customers/top?date_from=2026-02-01&date_to=2026-02-28&limit=10&tz=America/Argentina/Buenos_Aires
```

### Dashboard with timezone
```
GET /shops/1/metrics/dashboard?tz=America/Argentina/Buenos_Aires
```

## Suggested Period Selector Presets

Recommended presets for the UI dropdown:

| Label | `date_from` | `date_to` |
|---|---|---|
| Today | `today` | `today` |
| Yesterday | `yesterday` | `yesterday` |
| Last 7 days | `today - 6 days` | `today` |
| Last 30 days | `today - 29 days` | `today` |
| This month | 1st of current month | `today` |
| Last month | 1st of previous month | last day of previous month |
| Custom range | user picks start | user picks end |

All dates should be computed on the frontend and sent as `YYYY-MM-DD` strings.

## Response Shapes

### Revenue Trend
```json
{
  "trend": [
    { "date": "2026-02-01", "revenue": 15000.50, "order_count": 12 },
    { "date": "2026-02-02", "revenue": 8300.00, "order_count": 7 }
  ]
}
```

### Top Products
```json
{
  "products": [
    {
      "product_id": 42,
      "product_name": "Product Name",
      "image_url": "https://...",
      "quantity_sold": 150,
      "revenue": 45000.00
    }
  ]
}
```

### Top Customers
```json
{
  "customers": [
    {
      "customer_name": "John Doe",
      "customer_phone": "+5491155667788",
      "order_count": 8,
      "total_spent": 25000.00
    }
  ]
}
```

### Dashboard
```json
{
  "revenue": {
    "today": {
      "current": { "total_revenue": 5000.0, "order_count": 3, "aov": 1666.67 },
      "previous": { "total_revenue": 3000.0, "order_count": 2, "aov": 1500.0 },
      "revenue_change": 66.67,
      "order_change": 50.0,
      "aov_change": 11.11
    },
    "this_week": { "..." : "same structure" },
    "this_month": { "..." : "same structure" }
  },
  "orders": {
    "total_orders": 50,
    "status_distribution": [
      { "status": "pending", "count": 10 },
      { "status": "completed", "count": 35 }
    ],
    "cancellation_rate": 10.0,
    "avg_completion_time_hours": 2.5
  },
  "top_products": [ "...same as top products response" ],
  "payment_distribution": [
    { "name": "Cash", "code": "cash", "order_count": 20, "percentage": 40.0 }
  ],
  "delivery_distribution": [
    { "name": "Pickup", "code": "pickup", "order_count": 30, "percentage": 60.0 }
  ],
  "customers": {
    "total_unique": 25,
    "new_count": 10,
    "returning": 15
  }
}
```

## Implementation Notes

- **`date_to` is inclusive**: The backend converts `date_to` to end-of-day (`23:59:59.999999`), so `date_to=2026-02-28` covers all of February 28th.
- **Omitting dates**: If `date_from` and `date_to` are omitted, the backend applies its own defaults (typically current month or last 30 days depending on the endpoint).
- **Timezone**: Pass `tz` to ensure date boundaries align with the user's local time. Without it, the server uses its own timezone (UTC or server-configured).
- **Dashboard has fixed periods**: The dashboard endpoint always returns today/this_week/this_month comparisons. It doesn't accept `date_from`/`date_to` — only `tz`.
