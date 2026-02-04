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
	// testGetVersion(ctx, client)

	// Test createShipment method
	// testCreateShipment(ctx, client)

	// Test getMyShipments method - get shipments from last 7 days
	testGetMyShipments(ctx, client)
}

func testGetVersion(ctx context.Context, client *dhl.Client) {
	body, resp, err := client.GetVersion(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("=== getVersion ===")
	fmt.Println("HTTP status:", resp.Status)
	fmt.Println("Response body:")
	fmt.Println(string(body))
	fmt.Println()
}

func testCreateShipment(ctx context.Context, client *dhl.Client) {
	body, resp, err := client.CreateShipment(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("=== createShipment ===")
	fmt.Println("HTTP status:", resp.Status)
	fmt.Println("Response body:")
	fmt.Println(string(body))
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
