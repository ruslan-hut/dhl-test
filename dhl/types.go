package dhl

import "encoding/xml"

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

// AddressInfo represents address information for shipper or receiver
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
