package account

// accountRepository (legacy CRUD repository) removed.
// All account writes now go through EventSourcingService (see event_sourcing_service.go).
// All account reads now go through ReadModelRepository (see event_store.go).
