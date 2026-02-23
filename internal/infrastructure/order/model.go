package order

// OrderModel (legacy CRUD model for the `orders` table) is no longer used for writes.
// The write side uses order_events (OrderEventModel in event_store.go).
// The read  side uses order_read_models (OrderReadModelDB in event_store.go).
//
// The physical `orders` table is preserved for backward compatibility only.
// Drop it (or rename to orders_legacy) once all consumers have been migrated.