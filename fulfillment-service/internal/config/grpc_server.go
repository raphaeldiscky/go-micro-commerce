package config

import (
	"time"

	"github.com/spf13/viper"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
)

// GRPCServerConfig holds the configuration for the gRPC server.
type GRPCServerConfig struct {
	ServiceName       string        `mapstructure:"GRPC_SERVER_SERVICE_NAME"`
	Host              string        `mapstructure:"GRPC_SERVER_HOST"`
	Port              int           `mapstructure:"GRPC_SERVER_PORT"`
	GracePeriod       time.Duration `mapstructure:"GRPC_SERVER_GRACE_PERIOD"`
	ReadHeaderTimeout time.Duration `mapstructure:"GRPC_SERVER_READ_HEADER_TIMEOUT"`
}

// initGRPCServerConfig initializes the gRPC server configuration from environment variables.
func initGRPCServerConfig() *GRPCServerConfig {
	// Set defaults
	viper.SetDefault("GRPC_SERVER_SERVICE_NAME", pkgconstant.GRPCServiceNameFulfillment)
	viper.SetDefault("GRPC_SERVER_HOST", "0.0.0.0")
	viper.SetDefault("GRPC_SERVER_PORT", constant.GRPCPort)
	viper.SetDefault("GRPC_SERVER_GRACE_PERIOD", constant.GRPCGracePeriod)
	viper.SetDefault("GRPC_SERVER_READ_HEADER_TIMEOUT", constant.GRPCReadHeaderTimeout)

	grpcServerConfig := &GRPCServerConfig{}
	if err := viper.Unmarshal(grpcServerConfig); err != nil {
		panic(err)
	}

	return grpcServerConfig
}
