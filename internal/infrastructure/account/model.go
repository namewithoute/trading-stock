package account

// AccountModel (legacy CRUD model for accounts table) removed.
// Persistence for the write side uses account_events (AccountEventModel in event_store.go).
// Persistence for the read side uses account_read_models (AccountReadModelDB in event_store.go).
