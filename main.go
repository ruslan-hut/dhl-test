package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DHL24 WebAPI v2 SOAP Integration
// Documentation: https://dhl24.com.pl/en/webapi2/doc/info/zestawieniePolaczenia.html
// WSDL: https://dhl24.com.pl/webapi2?wsdl
// API Requirements: https://narzedzia.dhl.pl/files/dhl24/APIv2_ENG.pdf

// Config structs
type Config struct {
	DHL24 DHL24Config `json:"dhl24"`
}

type DHL24Config struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	AccountNumber string `json:"accountNumber"`
}

// loadConfig reads configuration from config.json file
func loadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w (copy config.example.json to config.json)", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	return &config, nil
}

// Structs for parsing getMyShipments response
type GetMyShipmentsEnvelope struct {
	XMLName xml.Name           `xml:"Envelope"`
	Body    GetMyShipmentsBody `xml:"Body"`
}

type GetMyShipmentsBody struct {
	Response GetMyShipmentsResponse `xml:"getMyShipmentsResponse"`
}

type GetMyShipmentsResponse struct {
	Result GetMyShipmentsResult `xml:"getMyShipmentsResult"`
}

type GetMyShipmentsResult struct {
	Items []ShipmentBasicData `xml:"item"`
}

type ShipmentBasicData struct {
	ShipmentID  string      `xml:"shipmentId"`
	Created     string      `xml:"created"`
	Shipper     AddressInfo `xml:"shipper"`
	Receiver    AddressInfo `xml:"receiver"`
	OrderStatus string      `xml:"orderStatus"`
}

type AddressInfo struct {
	Name            string `xml:"name"`
	PostalCode      string `xml:"postalCode"`
	City            string `xml:"city"`
	Street          string `xml:"street"`
	HouseNumber     string `xml:"houseNumber"`
	ApartmentNumber string `xml:"apartmentNumber"`
	ContactPerson   string `xml:"contactPerson"`
	ContactPhone    string `xml:"contactPhone"`
	ContactEmail    string `xml:"contactEmail"`
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("\nPlease copy config.example.json to config.json and fill in your credentials.")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Test getVersion method - check API version (no auth required)
	testGetVersion(ctx, client)

	// Test createShipment method
	// testCreateShipment(ctx, client, config)

	// Test getMyShipments method - get shipments from last 7 days
	testGetMyShipments(ctx, client, config)
}

// testGetVersion retrieves the DHL24 WebAPI version
// This is the only method that doesn't require authentication
func testGetVersion(ctx context.Context, client *http.Client) {
	const endpoint = "https://dhl24.com.pl/webapi2/provider/service.html?ws=1"

	soapEnvelope := `<?xml version="1.0" encoding="UTF-8"?>
		<soapenv:Envelope
			xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
			xmlns:ns="https://dhl24.com.pl/webapi2/provider/service.html?ws=1">
			<soapenv:Header/>
			<soapenv:Body>
				<ns:getVersion/>
			</soapenv:Body>
		</soapenv:Envelope>`

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewBufferString(soapEnvelope),
	)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// REQUIRED SOAP headers
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	// SOAPAction from WSDL <binding><operation><soap:operation soapAction="..."/>
	req.Header.Set("SOAPAction", "https://dhl24.com.pl/webapi2/provider/service.html?ws=1#getVersion")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("=== getVersion ===")
	fmt.Println("HTTP status:", resp.Status)
	fmt.Println("Response body:")
	fmt.Println(string(body))
	fmt.Println()
}

// testCreateShipment demonstrates calling the createShipment method
// Requires valid DHL24 API credentials from config.json
// To get credentials, register at: https://www.dhl24.com.pl or https://www.sandbox.dhl24.com.pl (test environment)
func testCreateShipment(ctx context.Context, client *http.Client, config *Config) {
	const endpoint = "https://dhl24.com.pl/webapi2/provider/service.html?ws=1"

	username := config.DHL24.Username
	password := config.DHL24.Password
	accountNumber := config.DHL24.AccountNumber

	// Minimal createShipments request (note: plural, wraps shipments in array)
	// Documentation: https://dhl24.com.pl/en/webapi2/doc.html
	// Product codes: https://dhl24.com.pl/en/webapi2/doc/service/createShipment.html
	// Common products: AH (DHL Parcel), PR (Premium), EK (Express 9:00), DR (Express 12:00), etc.
	// Possible responses:
	//   - Fault 100: Invalid credentials
	//   - Fault 101: Missing required parameter
	//   - Fault 131: Product retrieval error (product not available for account)
	soapEnvelope := `<?xml version="1.0" encoding="UTF-8"?>
		<soapenv:Envelope
			xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
			xmlns:ns="https://dhl24.com.pl/webapi2/provider/service.html?ws=1">
			<soapenv:Header/>
			<soapenv:Body>
				<ns:createShipments>
					<authData>
						<username>` + username + `</username>
						<password>` + password + `</password>
					</authData>
					<shipments>
						<item>
						<shipmentInfo>
							<dropOffType>REGULAR_PICKUP</dropOffType>
							<serviceType>EK</serviceType>
							<billing>
								<shippingPaymentType>SHIPPER</shippingPaymentType>
								<billingAccountNumber>` + accountNumber + `</billingAccountNumber>
								<paymentType>BANK_TRANSFER</paymentType>
							</billing>
							<shipmentDate>` + time.Now().AddDate(0, 0, 1).Format("2006-01-02") + `</shipmentDate>
							<shipmentStartHour>09:00</shipmentStartHour>
							<shipmentEndHour>17:00</shipmentEndHour>
							<labelType>BLP</labelType>
						</shipmentInfo>
						<shipper>
							<name>Test Sender</name>
							<postalCode>00-001</postalCode>
							<city>Warsaw</city>
							<street>Test Street</street>
							<houseNumber>1</houseNumber>
							<contactPhone>123456789</contactPhone>
							<contactEmail>sender@example.com</contactEmail>
						</shipper>
						<receiver>
							<country>PL</country>
							<name>Test Receiver</name>
							<postalCode>00-002</postalCode>
							<city>Krakow</city>
							<street>Receiver Street</street>
							<houseNumber>2</houseNumber>
							<contactPhone>987654321</contactPhone>
							<contactEmail>receiver@example.com</contactEmail>
						</receiver>
						<pieceList>
							<item>
								<type>ENVELOPE</type>
								<quantity>1</quantity>
								<weight>0.5</weight>
							</item>
						</pieceList>
						<paymentData>
							<paymentType>BANK_TRANSFER</paymentType>
							<payerType>SHIPPER</payerType>
							<accountNumber>` + accountNumber + `</accountNumber>
							<paymentMethod>BANK_TRANSFER</paymentMethod>
						</paymentData>
						<serviceDefinition>
							<product>AH</product>
						</serviceDefinition>
						<shipmentDate>` + time.Now().AddDate(0, 0, 1).Format("2006-01-02") + `</shipmentDate>
						</item>
					</shipments>
				</ns:createShipments>
			</soapenv:Body>
		</soapenv:Envelope>`

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewBufferString(soapEnvelope),
	)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// REQUIRED SOAP headers
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	// SOAPAction from WSDL <binding><operation><soap:operation soapAction="..."/>
	req.Header.Set("SOAPAction", "https://dhl24.com.pl/webapi2/provider/service.html?ws=1#createShipments")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("=== createShipment ===")
	fmt.Println("HTTP status:", resp.Status)
	fmt.Println("Response body:")
	fmt.Println(string(body))
}

// testGetMyShipments retrieves shipments list for the last 7 days
// Documentation: https://dhl24.com.pl/en/webapi2/doc/info/getMyShipments.html
func testGetMyShipments(ctx context.Context, client *http.Client, config *Config) {
	const endpoint = "https://dhl24.com.pl/webapi2/provider/service.html?ws=1"

	username := config.DHL24.Username
	password := config.DHL24.Password

	// Calculate date range: last 7 days
	createdTo := time.Now().Format("2006-01-02")
	createdFrom := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	// getMyShipments request
	// Returns maximum 100 records per request (use offset for pagination)
	// Documentation: https://dhl24.com.pl/en/webapi2/doc/info/getMyShipments.html
	soapEnvelope := `<?xml version="1.0" encoding="UTF-8"?>
		<soapenv:Envelope
			xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
			xmlns:ns="https://dhl24.com.pl/webapi2/provider/service.html?ws=1">
			<soapenv:Header/>
			<soapenv:Body>
				<ns:getMyShipments>
					<authData>
						<username>` + username + `</username>
						<password>` + password + `</password>
					</authData>
					<createdFrom>` + createdFrom + `</createdFrom>
					<createdTo>` + createdTo + `</createdTo>
					<offset>0</offset>
				</ns:getMyShipments>
			</soapenv:Body>
		</soapenv:Envelope>`

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewBufferString(soapEnvelope),
	)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// REQUIRED SOAP headers
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	// SOAPAction from WSDL <binding><operation><soap:operation soapAction="..."/>
	req.Header.Set("SOAPAction", "https://dhl24.com.pl/webapi2/provider/service.html?ws=1#getMyShipments")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("=== getMyShipments (last 3 days) ===")
	fmt.Printf("Date range: %s to %s\n", createdFrom, createdTo)
	fmt.Println("HTTP status:", resp.Status)
	fmt.Println()

	// Parse the XML response
	var envelope GetMyShipmentsEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		fmt.Println("Error parsing response:", err)
		fmt.Println("Raw response:")
		fmt.Println(string(body))
		return
	}

	shipments := envelope.Body.Response.Result.Items
	fmt.Printf("Found %d shipment(s):\n", len(shipments))
	fmt.Println("================================================================================")

	for i, shipment := range shipments {
		fmt.Printf("\n[%d] Shipment ID: %s\n", i+1, shipment.ShipmentID)
		fmt.Printf("    Created: %s\n", shipment.Created)
		fmt.Printf("    Status: %s\n", shipment.OrderStatus)
		fmt.Println()

		fmt.Println("    FROM:")
		fmt.Printf("      %s\n", shipment.Shipper.Name)
		fmt.Printf("      %s, %s %s", shipment.Shipper.Street, shipment.Shipper.HouseNumber, shipment.Shipper.ApartmentNumber)
		if shipment.Shipper.ApartmentNumber != "" {
			fmt.Printf("/%s", shipment.Shipper.ApartmentNumber)
		}
		fmt.Println()
		fmt.Printf("      %s %s\n", shipment.Shipper.PostalCode, shipment.Shipper.City)
		if shipment.Shipper.ContactPerson != "" {
			fmt.Printf("      Contact: %s\n", shipment.Shipper.ContactPerson)
		}
		if shipment.Shipper.ContactPhone != "" {
			fmt.Printf("      Phone: %s\n", shipment.Shipper.ContactPhone)
		}
		if shipment.Shipper.ContactEmail != "" {
			fmt.Printf("      Email: %s\n", shipment.Shipper.ContactEmail)
		}

		fmt.Println()
		fmt.Println("    TO:")
		fmt.Printf("      %s\n", shipment.Receiver.Name)
		fmt.Printf("      %s, %s", shipment.Receiver.Street, shipment.Receiver.HouseNumber)
		if shipment.Receiver.ApartmentNumber != "" {
			fmt.Printf("/%s", shipment.Receiver.ApartmentNumber)
		}
		fmt.Println()
		fmt.Printf("      %s %s\n", shipment.Receiver.PostalCode, shipment.Receiver.City)
		if shipment.Receiver.ContactPerson != "" && shipment.Receiver.ContactPerson != shipment.Receiver.Name {
			fmt.Printf("      Contact: %s\n", shipment.Receiver.ContactPerson)
		}
		if shipment.Receiver.ContactPhone != "" {
			fmt.Printf("      Phone: %s\n", shipment.Receiver.ContactPhone)
		}
		if shipment.Receiver.ContactEmail != "" {
			fmt.Printf("      Email: %s\n", shipment.Receiver.ContactEmail)
		}

		if i < len(shipments)-1 {
			fmt.Println("\n    ----------------------------------------------------------------------------")
		}
	}

	fmt.Println()
}
