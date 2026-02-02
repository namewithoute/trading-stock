package domain

type OrderState string

const (
	Pending  OrderState = "PENDING"
	Filled   OrderState = "FILLED"
	Canceled OrderState = "CANCELED"
)
