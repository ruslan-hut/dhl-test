# DHL24 WebAPI Go Client

A simple Go client for testing DHL24 WebAPI v2 SOAP integration.

## Features

- ✅ Get API version (`getVersion`)
- ✅ Create shipments (`createShipments`)
- ✅ List shipments (`getMyShipments`)
- 📦 Clean console output with parsed shipment details

## Prerequisites

- Go 1.16 or higher
- DHL24 API credentials (username, password, account number)

## Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd dhl-test
```

2. Copy the example configuration file:
```bash
cp config.example.json config.json
```

3. Edit `config.json` with your DHL24 credentials:
```json
{
  "dhl24": {
    "username": "your_username",
    "password": "your_password",
    "accountNumber": "your_account_number"
  }
}
```

## Getting DHL24 API Credentials

To obtain API credentials:
- **Production**: Contact DHL24 at https://www.dhl24.com.pl/en/contact.html (select "WebAPI - obtaining access")
- **Test Environment**: Register at https://www.sandbox.dhl24.com.pl

## Usage

Run all tests:
```bash
go run main.go
```

This will:
1. Get the API version
2. ~~Create a test shipment~~ (commented out by default)
3. List shipments from the last 7 days

## Configuration

Edit `main.go` to enable/disable specific tests:

```go
func main() {
    // ...

    testGetVersion(ctx, client)           // ✅ Active
    // testCreateShipment(ctx, client)    // 💤 Disabled
    testGetMyShipments(ctx, client)       // ✅ Active
}
```

## Available Service Types (Products)

Common DHL24 product codes:
- **EK** - Connect shipment (Parcel Connect)
- **CP** - Connect Plus shipment
- **AH** - Domestic shipment
- **PR** - Premium product
- **PI** - International shipment
- **09** - Domestic 09
- **12** - Domestic 12
- **SP** - Delivery to DHL point

## Documentation

- [DHL24 WebAPI v2 Documentation](https://dhl24.com.pl/en/webapi2/doc.html)
- [WSDL](https://dhl24.com.pl/webapi2?wsdl)
- [API Requirements PDF](https://narzedzia.dhl.pl/files/dhl24/APIv2_ENG.pdf)

## Project Structure

```
dhl-test/
├── main.go                 # Main application code
├── config.json             # Local configuration (gitignored)
├── config.example.json     # Example configuration
├── go.mod                  # Go module file
├── .gitignore             # Git ignore rules
└── README.md              # This file
```

## Common Errors

- **Fault 100**: Invalid credentials
- **Fault 101**: Missing required parameter
- **Fault 131**: Product not available for account

## License

MIT License

## Contributing

Pull requests are welcome! Please ensure your code follows Go best practices.
