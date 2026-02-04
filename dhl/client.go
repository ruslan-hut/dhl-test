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

	// SOAP namespace constants
	soapenvNS = "http://schemas.xmlsoap.org/soap/envelope/"
	dhlNS     = "https://dhl24.com.pl/webapi2/provider/service.html?ws=1"
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

// marshalSOAPRequest creates a SOAP envelope with the given body and marshals it to XML
func (c *Client) marshalSOAPRequest(body interface{}) ([]byte, error) {
	envelope := SOAPEnvelope{
		Soapenv: soapenvNS,
		NS:      dhlNS,
		Body:    SOAPBody{Content: body},
	}

	xmlData, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling SOAP request: %w", err)
	}

	// Add XML declaration
	return append([]byte(xml.Header), xmlData...), nil
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

// authData returns AuthData populated from client config
func (c *Client) authData() AuthData {
	return AuthData{
		Username: c.config.Username,
		Password: c.config.Password,
	}
}

// GetVersion retrieves the DHL24 WebAPI version
// This is the only method that doesn't require authentication
func (c *Client) GetVersion(ctx context.Context) (string, *http.Response, error) {
	reqBody, err := c.marshalSOAPRequest(GetVersionRequest{})
	if err != nil {
		return "", nil, err
	}

	body, resp, err := c.doRequest(ctx, reqBody, Endpoint+"#getVersion", "getVersion")
	if err != nil {
		return "", resp, err
	}

	var envelope SOAPResponseEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return "", resp, fmt.Errorf("error parsing response: %w", err)
	}

	if envelope.Body.GetVersionResponse == nil {
		return "", resp, fmt.Errorf("empty getVersion response")
	}

	return envelope.Body.GetVersionResponse.Version, resp, nil
}

// CreateShipments creates new shipments
// Documentation: https://dhl24.com.pl/en/webapi2/doc.html
// Product codes: https://dhl24.com.pl/en/webapi2/doc/service/createShipment.html
// Common products: AH (DHL Parcel), PR (Premium), EK (Express 9:00), DR (Express 12:00), etc.
// Possible responses:
//   - Fault 100: Invalid credentials
//   - Fault 101: Missing required parameter
//   - Fault 131: Product retrieval error (product not available for account)
func (c *Client) CreateShipments(ctx context.Context, shipments []ShipmentItem) ([]CreatedShipment, *http.Response, error) {
	request := CreateShipmentsRequest{
		AuthData: c.authData(),
		Shipments: Shipments{
			Items: shipments,
		},
	}

	reqBody, err := c.marshalSOAPRequest(request)
	if err != nil {
		return nil, nil, err
	}

	body, resp, err := c.doRequest(ctx, reqBody, Endpoint+"#createShipments", "createShipments")
	if err != nil {
		return nil, resp, err
	}

	var envelope SOAPResponseEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return nil, resp, fmt.Errorf("error parsing response: %w", err)
	}

	if envelope.Body.CreateShipmentsResponse == nil {
		return nil, resp, fmt.Errorf("empty createShipments response")
	}

	return envelope.Body.CreateShipmentsResponse.Result.Items, resp, nil
}

// CreateShipment creates a single shipment (convenience wrapper)
func (c *Client) CreateShipment(ctx context.Context, shipment ShipmentItem) (*CreatedShipment, *http.Response, error) {
	results, resp, err := c.CreateShipments(ctx, []ShipmentItem{shipment})
	if err != nil {
		return nil, resp, err
	}

	if len(results) == 0 {
		return nil, resp, fmt.Errorf("no shipment created")
	}

	return &results[0], resp, nil
}

// GetMyShipments retrieves shipments list for the specified date range
// Documentation: https://dhl24.com.pl/en/webapi2/doc/info/getMyShipments.html
// Returns maximum 100 records per request (use offset for pagination)
func (c *Client) GetMyShipments(ctx context.Context, createdFrom, createdTo string, offset int) ([]ShipmentBasicData, *http.Response, error) {
	request := GetMyShipmentsRequest{
		AuthData:    c.authData(),
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		Offset:      offset,
	}

	reqBody, err := c.marshalSOAPRequest(request)
	if err != nil {
		return nil, nil, err
	}

	body, resp, err := c.doRequest(ctx, reqBody, Endpoint+"#getMyShipments", "getMyShipments")
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
