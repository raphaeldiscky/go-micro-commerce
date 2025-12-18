// Package mapper provides mapping functions between OpenAPI types and service DTOs.
package mapper

import (
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/oapi"
)

// ToCalculateShippingRatesDTO maps OpenAPI request to service DTO.
func ToCalculateShippingRatesDTO(
	req *oapi.CalculateShippingRatesRequest,
) *dto.CalculateShippingRatesRequest {
	return &dto.CalculateShippingRatesRequest{
		Currency:    req.Currency,
		CourierID:   constant.CourierID(req.CourierId),
		Destination: toDestination(req.Destination),
		Origin:      toOrigin(req.Origin),
		Package:     toPackage(req.Package),
	}
}

func toDestination(addr oapi.Address) entity.Destination {
	return entity.Destination{
		City:        addr.City,
		State:       addr.State,
		PostalCode:  addr.PostalCode,
		CountryCode: addr.CountryCode,
	}
}

func toOrigin(addr oapi.Address) entity.Origin {
	return entity.Origin{
		City:        addr.City,
		State:       addr.State,
		PostalCode:  addr.PostalCode,
		CountryCode: addr.CountryCode,
	}
}

func toPackage(pkg oapi.Package) entity.Package {
	return entity.Package{
		WeightKG: decimal.RequireFromString(pkg.WeightKg),
		Width:    decimal.RequireFromString(pkg.Width),
		Height:   decimal.RequireFromString(pkg.Height),
		Length:   decimal.RequireFromString(pkg.Length),
		Unit:     string(pkg.Unit),
	}
}
