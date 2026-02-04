// Package dhl provides DHL24 WebAPI v2 SOAP Integration
// Documentation: https://dhl24.com.pl/en/webapi2/doc/info/zestawieniePolaczenia.html
// WSDL: https://dhl24.com.pl/webapi2?wsdl
// API Requirements: https://narzedzia.dhl.pl/files/dhl24/APIv2_ENG.pdf
package dhl

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// Endpoint is the DHL24 WebAPI endpoint
	Endpoint = "https://dhl24.com.pl/webapi2/provider/service.html?ws=1"
)

// Client represents a DHL24 API client
type Client struct {
	httpClient    *http.Client
	config        *DHL24Config
	debugFiles    bool
	debugFilesDir string
}

// NewClient creates a new DHL24 API client
func NewClient(config *DHL24Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		config:        config,
		debugFiles:    config.DebugFiles,
		debugFilesDir: config.DebugFilesDir,
	}
}

// getExecutableDir returns the directory where the executable is located
func getExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// writeDebugFile writes payload to a file with timestamp in the specified directory
// If dir is empty, defaults to the executable directory
func (c *Client) writeDebugFile(prefix string, payload []byte) {
	dir := c.debugFilesDir
	if dir == "" {
		dir = getExecutableDir()
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Warning: failed to create debug directory %s: %v\n", dir, err)
		return
	}

	timestamp := time.Now().Format("20060102_150405.000")
	filename := fmt.Sprintf("%s_%s.xml", prefix, timestamp)
	fullPath := filepath.Join(dir, filename)

	if err := os.WriteFile(fullPath, payload, 0644); err != nil {
		fmt.Printf("Warning: failed to write debug file %s: %v\n", fullPath, err)
	} else {
		fmt.Printf("Debug: wrote %s\n", fullPath)
	}
}

// doRequest performs an HTTP request and optionally logs request/response to files
func (c *Client) doRequest(ctx context.Context, body []byte, soapAction string, operationName string) ([]byte, *http.Response, error) {
	if c.debugFiles {
		c.writeDebugFile(operationName+"_request", body)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, fmt.Errorf("error reading response: %w", err)
	}

	if c.debugFiles {
		c.writeDebugFile(operationName+"_response", respBody)
	}

	return respBody, resp, nil
}

// GetVersion retrieves the DHL24 WebAPI version
// This is the only method that doesn't require authentication
func (c *Client) GetVersion(ctx context.Context) ([]byte, *http.Response, error) {
	soapEnvelope := `<?xml version="1.0" encoding="UTF-8"?>
		<soapenv:Envelope
			xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
			xmlns:ns="https://dhl24.com.pl/webapi2/provider/service.html?ws=1">
			<soapenv:Header/>
			<soapenv:Body>
				<ns:getVersion/>
			</soapenv:Body>
		</soapenv:Envelope>`

	return c.doRequest(ctx, []byte(soapEnvelope), Endpoint+"#getVersion", "getVersion")
}

// CreateShipment creates a new shipment
// Documentation: https://dhl24.com.pl/en/webapi2/doc.html
// Product codes: https://dhl24.com.pl/en/webapi2/doc/service/createShipment.html
// Common products: AH (DHL Parcel), PR (Premium), EK (Express 9:00), DR (Express 12:00), etc.
// Possible responses:
//   - Fault 100: Invalid credentials
//   - Fault 101: Missing required parameter
//   - Fault 131: Product retrieval error (product not available for account)
func (c *Client) CreateShipment(ctx context.Context) ([]byte, *http.Response, error) {
	username := c.config.Username
	password := c.config.Password
	accountNumber := c.config.AccountNumber

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
						<shipper>
							<name>ESMALTE INC</name>
							<postalCode>01249</postalCode>
							<city>Warsaw</city>
							<street>GOLESZOWSKA</street>
							<houseNumber>3</houseNumber>
							<contactPhone>123456789</contactPhone>
							<contactEmail>sender@example.com</contactEmail>
						</shipper>
						<receiver>
							<country>PL</country>
							<name>NIÑO PERESOZO</name>
							<postalCode>01249</postalCode>
							<city>Warsaw</city>
							<street>GOLESZOWSKA</street>
							<houseNumber>3</houseNumber>
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
						<payment>
							<paymentType>BANK_TRANSFER</paymentType>
							<payerType>SHIPPER</payerType>
							<accountNumber>` + accountNumber + `</accountNumber>
							<paymentMethod>BANK_TRANSFER</paymentMethod>
						</payment>
						<service>
							<product>AH</product>
						</service>
						<shipmentDate>` + time.Now().AddDate(0, 0, 1).Format("2006-01-02") + `</shipmentDate>
						<skipRestrictionCheck>true</skipRestrictionCheck>
						<comment></comment>
               			<content>zawartosc testowa</content>
						</item>
					</shipments>
				</ns:createShipments>
			</soapenv:Body>
		</soapenv:Envelope>`

	return c.doRequest(ctx, []byte(soapEnvelope), Endpoint+"#createShipments", "createShipments")
}

// GetMyShipments retrieves shipments list for the specified date range
// Documentation: https://dhl24.com.pl/en/webapi2/doc/info/getMyShipments.html
// Returns maximum 100 records per request (use offset for pagination)
func (c *Client) GetMyShipments(ctx context.Context, createdFrom, createdTo string, offset int) ([]ShipmentBasicData, *http.Response, error) {
	username := c.config.Username
	password := c.config.Password

	soapEnvelope := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
		<soapenv:Envelope
			xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
			xmlns:ns="https://dhl24.com.pl/webapi2/provider/service.html?ws=1">
			<soapenv:Header/>
			<soapenv:Body>
				<ns:getMyShipments>
					<authData>
						<username>%s</username>
						<password>%s</password>
					</authData>
					<createdFrom>%s</createdFrom>
					<createdTo>%s</createdTo>
					<offset>%d</offset>
				</ns:getMyShipments>
			</soapenv:Body>
		</soapenv:Envelope>`, username, password, createdFrom, createdTo, offset)

	body, resp, err := c.doRequest(ctx, []byte(soapEnvelope), Endpoint+"#getMyShipments", "getMyShipments")
	if err != nil {
		return nil, resp, err
	}

	var envelope GetMyShipmentsEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return nil, resp, fmt.Errorf("error parsing response: %w", err)
	}

	return envelope.Body.Response.Result.Items, resp, nil
}

// GetMyShipmentsLastDays retrieves shipments from the last N days
func (c *Client) GetMyShipmentsLastDays(ctx context.Context, days int) ([]ShipmentBasicData, *http.Response, error) {
	createdTo := time.Now().Format("2006-01-02")
	createdFrom := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	return c.GetMyShipments(ctx, createdFrom, createdTo, 0)
}

// PrintShipments prints shipments in a compact one-line format
func PrintShipments(shipments []ShipmentBasicData) {
	fmt.Printf("Found %d shipment(s):\n", len(shipments))
	for _, shipment := range shipments {
		fmt.Printf("%-30s | %s | %-20s | %s\n", shipment.ShipmentID, shipment.Created, shipment.OrderStatus, shipment.Receiver.Name)
	}
}
