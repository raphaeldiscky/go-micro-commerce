package gateway

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// DebugServices returns information about discovered services.
func (gw *Gateway) DebugServices() echo.HandlerFunc {
	return func(c echo.Context) error {
		services := []string{
			"auth-service",
			"product-service",
			"order-service",
			"notification-service",
			"fulfillment-service",
			"payment-service",
		}
		result := make(map[string]interface{})

		for _, serviceName := range services {
			endpoint, err := gw.serviceDiscovery.GetServiceEndpoint(serviceName)
			if err != nil {
				result[serviceName] = map[string]string{
					"status": "unavailable",
					"error":  err.Error(),
				}
			} else {
				result[serviceName] = map[string]string{
					"status":   "available",
					"endpoint": endpoint,
				}
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"services":  result,
			"timestamp": time.Now().Unix(),
		})
	}
}
