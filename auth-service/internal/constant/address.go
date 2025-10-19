package constant

// Address-related constants for validation and business logic.
const (
	// MaxAddressesPerUser defines the maximum number of addresses a user can have.
	MaxAddressesPerUser = 10

	// CountryCodeLength defines the required length for ISO 3166-1 alpha-2 country codes.
	CountryCodeLength = 2

	// MinLatitude defines the minimum valid latitude value.
	MinLatitude = -90.0

	// MaxLatitude defines the maximum valid latitude value.
	MaxLatitude = 90.0

	// MinLongitude defines the minimum valid longitude value.
	MinLongitude = -180.0

	// MaxLongitude defines the maximum valid longitude value.
	MaxLongitude = 180.0
)
