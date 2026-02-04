package dhl

import "encoding/xml"

// ============================================================================
// SOAP Envelope Types
// ============================================================================

// SOAPEnvelope represents a SOAP envelope for requests
type SOAPEnvelope struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	Soapenv string   `xml:"xmlns:soapenv,attr"`
	NS      string   `xml:"xmlns:ns,attr"`
	Header  struct{} `xml:"soapenv:Header"`
	Body    SOAPBody `xml:"soapenv:Body"`
}

// SOAPBody wraps the request content
type SOAPBody struct {
	Content interface{}
}

// SOAPResponseEnvelope represents a SOAP envelope for responses
type SOAPResponseEnvelope struct {
	XMLName xml.Name         `xml:"Envelope"`
	Body    SOAPResponseBody `xml:"Body"`
}

// SOAPResponseBody wraps the response content
type SOAPResponseBody struct {
	GetVersionResponse      *GetVersionResponse      `xml:"getVersionResponse,omitempty"`
	CreateShipmentsResponse *CreateShipmentsResponse `xml:"createShipmentsResponse,omitempty"`
	GetMyShipmentsResponse  *GetMyShipmentsResponse  `xml:"getMyShipmentsResponse,omitempty"`
}

// ============================================================================
// Common Types
// ============================================================================

// AuthData contains authentication credentials
type AuthData struct {
	Username string `xml:"username"`
	Password string `xml:"password"`
}

// Address represents shipper or receiver address
type Address struct {
	Country         string `xml:"country,omitempty"`
	Name            string `xml:"name"`
	PostalCode      string `xml:"postalCode"`
	City            string `xml:"city"`
	Street          string `xml:"street"`
	HouseNumber     string `xml:"houseNumber"`
	ApartmentNumber string `xml:"apartmentNumber,omitempty"`
	ContactPerson   string `xml:"contactPerson,omitempty"`
	ContactPhone    string `xml:"contactPhone"`
	ContactEmail    string `xml:"contactEmail"`
}

// Piece represents a single piece in a shipment
type Piece struct {
	Type     string  `xml:"type"`
	Quantity int     `xml:"quantity"`
	Weight   float64 `xml:"weight"`
}

// PieceList contains list of pieces
type PieceList struct {
	Items []Piece `xml:"item"`
}

// Payment contains payment information
type Payment struct {
	PaymentType   string `xml:"paymentType"`
	PayerType     string `xml:"payerType"`
	AccountNumber string `xml:"accountNumber"`
	PaymentMethod string `xml:"paymentMethod"`
}

// Service contains service/product information
type Service struct {
	Product string `xml:"product"`
}

// ============================================================================
// GetVersion Types
// ============================================================================

// GetVersionRequest represents getVersion SOAP request
type GetVersionRequest struct {
	XMLName xml.Name `xml:"ns:getVersion"`
}

// GetVersionResponse represents getVersion SOAP response
type GetVersionResponse struct {
	Version string `xml:"getVersionResult"`
}

// ============================================================================
// CreateShipments Types
// ============================================================================

// CreateShipmentsRequest represents createShipments SOAP request
type CreateShipmentsRequest struct {
	XMLName   xml.Name  `xml:"ns:createShipments"`
	AuthData  AuthData  `xml:"authData"`
	Shipments Shipments `xml:"shipments"`
}

// Shipments contains list of shipment items
type Shipments struct {
	Items []ShipmentItem `xml:"item"`
}

// ShipmentItem represents a single shipment to create
type ShipmentItem struct {
	Shipper              Address   `xml:"shipper"`
	Receiver             Address   `xml:"receiver"`
	PieceList            PieceList `xml:"pieceList"`
	Payment              Payment   `xml:"payment"`
	Service              Service   `xml:"service"`
	ShipmentDate         string    `xml:"shipmentDate"`
	SkipRestrictionCheck bool      `xml:"skipRestrictionCheck"`
	Comment              string    `xml:"comment"`
	Content              string    `xml:"content"`
}

// CreateShipmentsResponse represents createShipments SOAP response
type CreateShipmentsResponse struct {
	Result CreateShipmentsResult `xml:"createShipmentsResult"`
}

// CreateShipmentsResult contains created shipments
type CreateShipmentsResult struct {
	Items []CreatedShipment `xml:"item"`
}

// CreatedShipment represents a successfully created shipment
type CreatedShipment struct {
	ShipmentID  string `xml:"shipmentId"`
	ShipmentNo  string `xml:"shipmentNo,omitempty"`
	OrderStatus string `xml:"orderStatus,omitempty"`
}

// ============================================================================
// GetMyShipments Types
// ============================================================================

// GetMyShipmentsRequest represents getMyShipments SOAP request
type GetMyShipmentsRequest struct {
	XMLName     xml.Name `xml:"ns:getMyShipments"`
	AuthData    AuthData `xml:"authData"`
	CreatedFrom string   `xml:"createdFrom"`
	CreatedTo   string   `xml:"createdTo"`
	Offset      int      `xml:"offset"`
}

// GetMyShipmentsEnvelope represents the SOAP envelope for getMyShipments response
type GetMyShipmentsEnvelope struct {
	XMLName xml.Name           `xml:"Envelope"`
	Body    GetMyShipmentsBody `xml:"Body"`
}

// GetMyShipmentsBody represents the SOAP body for getMyShipments response
type GetMyShipmentsBody struct {
	Response GetMyShipmentsResponse `xml:"getMyShipmentsResponse"`
}

// GetMyShipmentsResponse represents the getMyShipments response
type GetMyShipmentsResponse struct {
	Result GetMyShipmentsResult `xml:"getMyShipmentsResult"`
}

// GetMyShipmentsResult contains the list of shipments
type GetMyShipmentsResult struct {
	Items []ShipmentBasicData `xml:"item"`
}

// ShipmentBasicData represents basic shipment information
type ShipmentBasicData struct {
	ShipmentID  string      `xml:"shipmentId"`
	Created     string      `xml:"created"`
	Shipper     AddressInfo `xml:"shipper"`
	Receiver    AddressInfo `xml:"receiver"`
	OrderStatus string      `xml:"orderStatus"`
}

// AddressInfo represents address information for shipper or receiver (response)
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
