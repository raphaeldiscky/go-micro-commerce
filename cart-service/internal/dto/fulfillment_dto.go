// Package dto contains data transfer objects for fulfillment service.
package dto

// GetShippingCostRequest represents a request to calculate shipping costs.
type GetShippingCostRequest struct {
	Currency               string
	CourierID              string
	DestinationCity        string
	DestinationState       string
	DestinationPostalCode  string
	DestinationCountryCode string
	OriginCity             string
	OriginState            string
	OriginPostalCode       string
	OriginCountryCode      string
	WeightKG               string
	Width                  string
	Height                 string
	Length                 string
	Unit                   string
}

// GetShippingCostResponse represents the response from shipping cost calculation.
type GetShippingCostResponse struct {
	Success      bool
	ShippingCost float64
	Currency     string
	ErrorMessage string
}
