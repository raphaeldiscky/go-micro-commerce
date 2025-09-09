package config

import (
	"log"

	"github.com/spf13/viper"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// GRPCServerConfig holds the configuration for the gRPC server.
type GRPCServerConfig struct {
	ServiceName string `mapstructure:"GRPC_SERVER_SERVICE_NAME"`
	Host        string `mapstructure:"GRPC_SERVER_HOST"`
	Port        int    `mapstructure:"GRPC_SERVER_PORT"`
	GracePeriod int    `mapstructure:"GRPC_SERVER_GRACE_PERIOD"`
}

// initGRPCServerConfig initializes the gRPC server configuration from environment variables.
func initGRPCServerConfig() *GRPCServerConfig {
	// Set defaults
	viper.SetDefault("GRPC_SERVER_SERVICE_NAME", pkgconstant.GRPCServiceNameFulfillment)
	viper.SetDefault("GRPC_SERVER_HOST", "0.0.0.0")
	viper.SetDefault("GRPC_SERVER_PORT", 50055)
	viper.SetDefault("GRPC_SERVER_GRACE_PERIOD", 10)

	grpcServerConfig := &GRPCServerConfig{}

	if err := viper.Unmarshal(&grpcServerConfig); err != nil {
		log.Fatalf("error mapping grpc server config: %v", err)
	}

	return grpcServerConfig
}
