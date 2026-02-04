package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"dhl-test/dhl"
)

func main() {
	// Load configuration
	config, err := dhl.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("\nPlease copy config.example.json to config.json and fill in your credentials.")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Create DHL client
	client := dhl.NewClient(&config.DHL24)

	// Test getVersion method - check API version (no auth required)
	testGetVersion(ctx, client)

	// Test createShipment method
	// testCreateShipment(ctx, client, config)

	// Test getMyShipments method - get shipments from last 7 days
	testGetMyShipments(ctx, client)
}

func testGetVersion(ctx context.Context, client *dhl.Client) {
	version, resp, err := client.GetVersion(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("=== getVersion ===")
	fmt.Println("HTTP status:", resp.Status)
	fmt.Println("API Version:", version)
	fmt.Println()
}

func testCreateShipment(ctx context.Context, client *dhl.Client, config *dhl.Config) {
	// Build shipment from structs
	shipment := dhl.ShipmentItem{
		Shipper: dhl.Address{
			Name:         "ESMALTE INC",
			PostalCode:   "01249",
			City:         "Warsaw",
			Street:       "GOLESZOWSKA",
			HouseNumber:  "3",
			ContactPhone: "123456789",
			ContactEmail: "sender@example.com",
		},
		Receiver: dhl.Address{
			Country:      "PL",
			Name:         "Test Receiver",
			PostalCode:   "01249",
			City:         "Warsaw",
			Street:       "GOLESZOWSKA",
			HouseNumber:  "3",
			ContactPhone: "987654321",
			ContactEmail: "receiver@example.com",
		},
		PieceList: dhl.PieceList{
			Items: []dhl.Piece{
				{
					Type:     "ENVELOPE",
					Quantity: 1,
					Weight:   0.5,
				},
			},
		},
		Payment: dhl.Payment{
			PaymentType:   "BANK_TRANSFER",
			PayerType:     "SHIPPER",
			AccountNumber: config.DHL24.AccountNumber,
			PaymentMethod: "BANK_TRANSFER",
		},
		Service: dhl.Service{
			Product: "AH",
		},
		ShipmentDate:         time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		SkipRestrictionCheck: true,
		Content:              "test content",
	}

	result, resp, err := client.CreateShipment(ctx, shipment)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("=== createShipment ===")
	fmt.Println("HTTP status:", resp.Status)
	fmt.Printf("Created shipment ID: %s\n", result.ShipmentID)
}

func testGetMyShipments(ctx context.Context, client *dhl.Client) {
	shipments, resp, err := client.GetMyShipmentsLastDays(ctx, 7)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("=== getMyShipments (last 7 days) ===")
	fmt.Println("HTTP status:", resp.Status)
	fmt.Println()

	dhl.PrintShipments(shipments)
}
