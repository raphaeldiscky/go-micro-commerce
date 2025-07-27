# Mock Generation

This directory contains automatically generated mocks using [uber-go/mock](https://github.com/uber-go/mock).

## Regenerating Mocks

To regenerate all mocks:

```bash
make mocks
```

## Adding New Service Mocks

1. Add the `//go:generate` directive to your interface file:

```go
package interfaces

//go:generate mockgen -source=your_service.go -destination=../../mocks/mock_your_service.go -package=mocks

type YourService interface {
    DoSomething() error
}
```

2. Run the mock generation:

```bash
make mocks
```

3. Use the generated mock in your tests:

```go
func TestSomething(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockService := mocks.NewMockYourService(ctrl)
    mockService.EXPECT().DoSomething().Return(nil).Times(1)

    // Your test logic here
}
```

## Generated Files

- `mock_product_service.go` - Mock for ProductService interface
- `mock_seller_service.go` - Mock for SellerService interface

> **Note**: Do not edit these files manually as they are automatically generated and will be overwritten.
