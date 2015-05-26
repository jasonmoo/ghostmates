package ghostmates

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	TestManifestDescription = "Manifest description"
	TestManifestReference   = "Manifest reference"

	TestPickupName                 = "Pickup spot"
	TestPickupAddress              = "555 W 18th St, New York, NY 10011"
	TestPickupStreetAddress1       = "555 W 18th St"
	TestPickupCity                 = "New York"
	TestPickupState                = "NY"
	TestPickupZipCode              = "10011"
	TestPickupCountry              = "US"
	TestPickupLat                  = 40.74527
	TestPickupLng                  = -74.007889
	TestPickupPhoneNumber          = "555-222-3333"
	TestPickupOptionalBusinessName = "Pickup LLC"
	TestPickupOptionalNotes        = "Pickup notes"

	TestDropoffName                 = "Dropoff spot"
	TestDropoffAddress              = "620 8th Ave, New York, NY, 10018"
	TestDropoffStreetAddress1       = "620 8th Ave"
	TestDropoffCity                 = "New York"
	TestDropoffState                = "NY"
	TestDropoffZipCode              = "10018"
	TestDropoffCountry              = "US"
	TestDropoffLat                  = 40.75626
	TestDropoffLng                  = -73.990501
	TestDropoffPhoneNumber          = "620-222-3333"
	TestDropoffOptionalBusinessName = "Dropoff LLC"
	TestDropoffOptionalNotes        = "Dropoff notes"
)

var (
	TestAPIKey     string
	TestCustomerId string
	TestTimeout    = 2 * time.Second
)

func init() {
	TestAPIKey = os.Getenv("POSTMATES_API_KEY")
	TestCustomerId = os.Getenv("POSTMATES_CUSTOMER_ID")
	if len(TestAPIKey) == 0 || len(TestCustomerId) == 0 {
		log.Fatalf("environment variables POSTMATES_API_KEY and POSTMATES_CUSTOMER_ID required to run tests")
	}
}

func TestGetQuote(t *testing.T) {

	client := NewClient(TestCustomerId, TestAPIKey, TestTimeout)

	dq, err := client.GetQuote(TestPickupAddress, TestDropoffAddress)
	if err != nil {
		t.Error(err)
	}

	if dq.Kind != DeliveryQuoteKind {
		t.Errorf("Expected %q, got %q", DeliveryQuoteKind, dq.Kind)
	}
	if len(dq.ID) == 0 {
		t.Errorf("Expected non-zero length ID")
	}
	if dq.Created.IsZero() {
		t.Errorf("Expected non-zero 'created' timestamp")
	}
	if dq.Expires.IsZero() {
		t.Errorf("Expected non-zero 'expires' timestamp")
	}
	if dq.DropoffEta.IsZero() {
		t.Errorf("Expected non-zero 'dropoff_eta' timestamp")
	}
	if dq.DurationSeconds == 0 {
		t.Errorf("Expected non-zero 'duration'")
	}
	if len(dq.Currency) == 0 {
		t.Errorf("Expected a currency type")
	}
	if dq.Fee == 0 {
		t.Errorf("Expected non-zero 'fee'")
	}

}

func TestClientTimeout(t *testing.T) {

	// get client with absurdly low timeout
	client := NewClient(TestCustomerId, TestAPIKey, time.Nanosecond)

	_, err := client.GetQuote(TestPickupAddress, TestDropoffAddress)
	if err == nil {
		t.Errorf("Expected a canceled error, got nil")
	}
	if !strings.Contains(err.Error(), "canceled") {
		t.Errorf("Expected a canceled error, got %q", err.Error())
	}

}

func TestCreateDelivery(t *testing.T) {

	client := NewClient(TestCustomerId, TestAPIKey, TestTimeout)

	quote, err := client.GetQuote(TestPickupAddress, TestDropoffAddress)
	if err != nil {
		t.Error(err)
	}

	manifest := NewManifest(TestManifestDescription, TestManifestReference)

	pickup := NewDeliverySpot(TestPickupName, TestPickupAddress, TestPickupPhoneNumber)
	pickup.BusinessName = TestPickupOptionalBusinessName
	pickup.Notes = TestPickupOptionalNotes

	dropoff := NewDeliverySpot(TestDropoffName, TestDropoffAddress, TestDropoffPhoneNumber)
	dropoff.BusinessName = TestDropoffOptionalBusinessName
	dropoff.Notes = TestDropoffOptionalNotes

	err = client.CreateDelivery(manifest, pickup, dropoff, quote)
	if err != nil {
		t.Error(err)
	}

}

func TestGetDeliveries(t *testing.T) {

	client := NewClient(TestCustomerId, TestAPIKey, TestTimeout)

	ds, err := client.GetDeliveries(AllFilter, 1)
	if err != nil {
		t.Error(err)
	}

	if len(ds) != 1 {
		t.Errorf("Expected 1 delivery, got %d", len(ds))
		t.Fail()
	}
	if ds[0].LiveMode {
		t.Errorf("woah live mode enabled!  this could be very bad")
		t.Fail()
	}
	if ds[0].Kind != DeliveryKind {
		t.Errorf("Expected %q, got %q", DeliveryKind, ds[0].Kind)
	}
	if len(ds[0].ID) == 0 {
		t.Errorf("Expected an ID")
	}
	if len(ds[0].Status) == 0 {
		t.Errorf("Expected a Status")
	}
	if len(ds[0].Currency) == 0 {
		t.Errorf("Expected a Currency")
	}
	if len(ds[0].QuoteID) == 0 {
		t.Errorf("Expected a QuoteID")
	}
	if ds[0].Created == nil || ds[0].Created.IsZero() {
		t.Errorf("Expected non-nil, non-zero Created timestamp")
	}

	if ds[0].Status != StatusPending {

		if ds[0].Updated == nil || ds[0].Updated.IsZero() {
			t.Errorf("Expected non-nil, non-zero Updated timestamp")
		}
		if ds[0].DropoffEta == nil || ds[0].DropoffEta.IsZero() {
			t.Errorf("Expected non-nil, non-zero DropoffEta timestamp")
		}
		if ds[0].PickupEta == nil || ds[0].PickupEta.IsZero() {
			t.Errorf("Expected non-nil, non-zero PickupEta timestamp")
		}
		if len(ds[0].Courier.Name) == 0 {
			t.Errorf("Expected a Name")
		}
		if len(ds[0].Courier.ImgHref) == 0 {
			t.Errorf("Expected a ImgHref")
		}
		if len(ds[0].Courier.PhoneNumber) == 0 {
			t.Errorf("Expected a PhoneNumber")
		}
		if len(ds[0].Courier.VehicleType) == 0 {
			t.Errorf("Expected a VehicleType")
		}
		if ds[0].Courier.Location.Lat == 0 || ds[0].Courier.Location.Lng == 0 {
			t.Errorf("Expected non-zero lat/lng")
		}

	}

	if ds[0].DropoffDeadline == nil || ds[0].DropoffDeadline.IsZero() {
		t.Errorf("Expected non-nil, non-zero PickupDeadline timestamp")
	}
	if ds[0].Fee == 0 {
		t.Errorf("Expected a non-zero Fee")
	}

	if ds[0].Pickup.Address != TestPickupStreetAddress1 {
		t.Errorf("Expected %q, got %q", TestPickupStreetAddress1, ds[0].Pickup.Address)
	}
	if ds[0].Pickup.DetailedAddress.City != TestPickupCity {
		t.Errorf("Expected %q, got %q", TestPickupCity, ds[0].Pickup.DetailedAddress.City)
	}
	if ds[0].Pickup.DetailedAddress.Country != TestPickupCountry {
		t.Errorf("Expected %q, got %q", TestPickupCountry, ds[0].Pickup.DetailedAddress.Country)
	}
	if ds[0].Pickup.DetailedAddress.State != TestPickupState {
		t.Errorf("Expected %q, got %q", TestPickupState, ds[0].Pickup.DetailedAddress.State)
	}
	if ds[0].Pickup.DetailedAddress.StreetAddress1 != TestPickupStreetAddress1 {
		t.Errorf("Expected %q, got %q", TestPickupStreetAddress1, ds[0].Pickup.DetailedAddress.StreetAddress1)
	}
	if ds[0].Pickup.DetailedAddress.ZipCode != TestPickupZipCode {
		t.Errorf("Expected %q, got %q", TestPickupZipCode, ds[0].Pickup.DetailedAddress.ZipCode)
	}
	if ds[0].Pickup.Location.Lat != TestPickupLat {
		t.Errorf("Expected %q, got %q", TestPickupLat, ds[0].Pickup.Location.Lat)
	}
	if ds[0].Pickup.Location.Lng != TestPickupLng {
		t.Errorf("Expected %q, got %q", TestPickupLng, ds[0].Pickup.Location.Lng)
	}
	if ds[0].Pickup.Name != TestPickupName+", "+TestPickupOptionalBusinessName {
		t.Errorf("Expected %q, got %q", TestPickupName+", "+TestPickupOptionalBusinessName, ds[0].Pickup.Name)
	}
	if ds[0].Pickup.Notes != TestPickupOptionalNotes {
		t.Errorf("Expected %q, got %q", TestPickupOptionalNotes, ds[0].Pickup.Notes)
	}
	if ds[0].Pickup.PhoneNumber != TestPickupPhoneNumber {
		t.Errorf("Expected %q, got %q", TestPickupPhoneNumber, ds[0].Pickup.PhoneNumber)
	}

	if ds[0].DropoffDeadline == nil || ds[0].DropoffDeadline.IsZero() {
		t.Errorf("Expected non-nil, non-zero DropoffDeadline timestamp")
	}
	if ds[0].Fee == 0 {
		t.Errorf("Expected a non-zero Fee")
	}
	if ds[0].Dropoff.Address != TestDropoffStreetAddress1 {
		t.Errorf("Expected %q, got %q", TestDropoffStreetAddress1, ds[0].Dropoff.Address)
	}
	if ds[0].Dropoff.DetailedAddress.City != TestDropoffCity {
		t.Errorf("Expected %q, got %q", TestDropoffCity, ds[0].Dropoff.DetailedAddress.City)
	}
	if ds[0].Dropoff.DetailedAddress.Country != TestDropoffCountry {
		t.Errorf("Expected %q, got %q", TestDropoffCountry, ds[0].Dropoff.DetailedAddress.Country)
	}
	if ds[0].Dropoff.DetailedAddress.State != TestDropoffState {
		t.Errorf("Expected %q, got %q", TestDropoffState, ds[0].Dropoff.DetailedAddress.State)
	}
	if ds[0].Dropoff.DetailedAddress.StreetAddress1 != TestDropoffStreetAddress1 {
		t.Errorf("Expected %q, got %q", TestDropoffStreetAddress1, ds[0].Dropoff.DetailedAddress.StreetAddress1)
	}
	if ds[0].Dropoff.DetailedAddress.ZipCode != TestDropoffZipCode {
		t.Errorf("Expected %q, got %q", TestDropoffZipCode, ds[0].Dropoff.DetailedAddress.ZipCode)
	}
	if ds[0].Dropoff.Location.Lat != TestDropoffLat {
		t.Errorf("Expected %q, got %q", TestDropoffLat, ds[0].Dropoff.Location.Lat)
	}
	if ds[0].Dropoff.Location.Lng != TestDropoffLng {
		t.Errorf("Expected %q, got %q", TestDropoffLng, ds[0].Dropoff.Location.Lng)
	}
	if ds[0].Dropoff.Name != TestDropoffName+", "+TestDropoffOptionalBusinessName {
		t.Errorf("Expected %q, got %q", TestDropoffName+", "+TestDropoffOptionalBusinessName, ds[0].Dropoff.Name)
	}
	if ds[0].Dropoff.Notes != TestDropoffOptionalNotes {
		t.Errorf("Expected %q, got %q", TestDropoffOptionalNotes, ds[0].Dropoff.Notes)
	}
	if ds[0].Dropoff.PhoneNumber != TestDropoffPhoneNumber {
		t.Errorf("Expected %q, got %q", TestDropoffPhoneNumber, ds[0].Dropoff.PhoneNumber)
	}

	if ds[0].Manifest.Description != TestManifestDescription {
		t.Errorf("Expected %q, got %q", ds[0].Manifest.Description, TestManifestDescription)
	}
	if ds[0].Manifest.Reference != TestManifestReference {
		t.Errorf("Expected %q, got %q", ds[0].Manifest.Reference, TestManifestReference)
	}

	if len(ds[0].RelatedDeliveries) != 0 {
		t.Errorf("Expected 0 RelatedDeliveries, got %d", len(ds[0].RelatedDeliveries))
	}

}

func TestGetAllDeliveries(t *testing.T) {

	client := NewClient(TestCustomerId, TestAPIKey, TestTimeout)

	ds, err := client.GetDeliveries(AllFilter, AllDeliveries)
	if err != nil {
		t.Error(err)
	}

	if len(ds) == 0 {
		t.Errorf("Expected multiple deliveries, got %d", len(ds))
	}

}

func TestGetOngoingDeliveries(t *testing.T) {

	client := NewClient(TestCustomerId, TestAPIKey, TestTimeout)

	_, err := client.GetOngoingDeliveries(AllDeliveries)
	if err != nil {
		t.Error(err)
	}

}

func TestGetDelivery(t *testing.T) {

	client := NewClient(TestCustomerId, TestAPIKey, TestTimeout)

	ds, err := client.GetDeliveries(AllFilter, 1)
	if err != nil {
		t.Error(err)
	}

	d, err := client.GetDelivery(ds[0].ID)
	if err != nil {
		t.Error(err)
	}

	if d.LiveMode {
		t.Errorf("woah live mode enabled!  this could be very bad")
		t.Fail()
	}
	if d.Kind != DeliveryKind {
		t.Errorf("Expected %q, got %q", DeliveryKind, d.Kind)
	}
	if len(d.ID) == 0 {
		t.Errorf("Expected an ID")
	}
	if len(d.Status) == 0 {
		t.Errorf("Expected a Status")
	}
	if len(d.Currency) == 0 {
		t.Errorf("Expected a Currency")
	}
	if len(d.QuoteID) == 0 {
		t.Errorf("Expected a QuoteID")
	}
	if d.Created == nil || d.Created.IsZero() {
		t.Errorf("Expected non-nil, non-zero Created timestamp")
	}

	if d.Status != StatusPending {

		if d.Updated == nil || d.Updated.IsZero() {
			t.Errorf("Expected non-nil, non-zero Updated timestamp")
		}
		if d.DropoffEta == nil || d.DropoffEta.IsZero() {
			t.Errorf("Expected non-nil, non-zero DropoffEta timestamp")
		}
		if d.PickupEta == nil || d.PickupEta.IsZero() {
			t.Errorf("Expected non-nil, non-zero PickupEta timestamp")
		}
		if len(d.Courier.Name) == 0 {
			t.Errorf("Expected a Name")
		}
		if len(d.Courier.ImgHref) == 0 {
			t.Errorf("Expected a ImgHref")
		}
		if len(d.Courier.PhoneNumber) == 0 {
			t.Errorf("Expected a PhoneNumber")
		}
		if len(d.Courier.VehicleType) == 0 {
			t.Errorf("Expected a VehicleType")
		}
		if d.Courier.Location.Lat == 0 || d.Courier.Location.Lng == 0 {
			t.Errorf("Expected non-zero lat/lng")
		}

	}

	if d.DropoffDeadline == nil || d.DropoffDeadline.IsZero() {
		t.Errorf("Expected non-nil, non-zero PickupDeadline timestamp")
	}
	if d.Fee == 0 {
		t.Errorf("Expected a non-zero Fee")
	}

	if d.Pickup.Address != TestPickupStreetAddress1 {
		t.Errorf("Expected %q, got %q", TestPickupStreetAddress1, d.Pickup.Address)
	}
	if d.Pickup.DetailedAddress.City != TestPickupCity {
		t.Errorf("Expected %q, got %q", TestPickupCity, d.Pickup.DetailedAddress.City)
	}
	if d.Pickup.DetailedAddress.Country != TestPickupCountry {
		t.Errorf("Expected %q, got %q", TestPickupCountry, d.Pickup.DetailedAddress.Country)
	}
	if d.Pickup.DetailedAddress.State != TestPickupState {
		t.Errorf("Expected %q, got %q", TestPickupState, d.Pickup.DetailedAddress.State)
	}
	if d.Pickup.DetailedAddress.StreetAddress1 != TestPickupStreetAddress1 {
		t.Errorf("Expected %q, got %q", TestPickupStreetAddress1, d.Pickup.DetailedAddress.StreetAddress1)
	}
	if d.Pickup.DetailedAddress.ZipCode != TestPickupZipCode {
		t.Errorf("Expected %q, got %q", TestPickupZipCode, d.Pickup.DetailedAddress.ZipCode)
	}
	if d.Pickup.Location.Lat != TestPickupLat {
		t.Errorf("Expected %q, got %q", TestPickupLat, d.Pickup.Location.Lat)
	}
	if d.Pickup.Location.Lng != TestPickupLng {
		t.Errorf("Expected %q, got %q", TestPickupLng, d.Pickup.Location.Lng)
	}
	if d.Pickup.Name != TestPickupName+", "+TestPickupOptionalBusinessName {
		t.Errorf("Expected %q, got %q", TestPickupName+", "+TestPickupOptionalBusinessName, d.Pickup.Name)
	}
	if d.Pickup.Notes != TestPickupOptionalNotes {
		t.Errorf("Expected %q, got %q", TestPickupOptionalNotes, d.Pickup.Notes)
	}
	if d.Pickup.PhoneNumber != TestPickupPhoneNumber {
		t.Errorf("Expected %q, got %q", TestPickupPhoneNumber, d.Pickup.PhoneNumber)
	}

	if d.DropoffDeadline == nil || d.DropoffDeadline.IsZero() {
		t.Errorf("Expected non-nil, non-zero DropoffDeadline timestamp")
	}
	if d.Fee == 0 {
		t.Errorf("Expected a non-zero Fee")
	}
	if d.Dropoff.Address != TestDropoffStreetAddress1 {
		t.Errorf("Expected %q, got %q", TestDropoffStreetAddress1, d.Dropoff.Address)
	}
	if d.Dropoff.DetailedAddress.City != TestDropoffCity {
		t.Errorf("Expected %q, got %q", TestDropoffCity, d.Dropoff.DetailedAddress.City)
	}
	if d.Dropoff.DetailedAddress.Country != TestDropoffCountry {
		t.Errorf("Expected %q, got %q", TestDropoffCountry, d.Dropoff.DetailedAddress.Country)
	}
	if d.Dropoff.DetailedAddress.State != TestDropoffState {
		t.Errorf("Expected %q, got %q", TestDropoffState, d.Dropoff.DetailedAddress.State)
	}
	if d.Dropoff.DetailedAddress.StreetAddress1 != TestDropoffStreetAddress1 {
		t.Errorf("Expected %q, got %q", TestDropoffStreetAddress1, d.Dropoff.DetailedAddress.StreetAddress1)
	}
	if d.Dropoff.DetailedAddress.ZipCode != TestDropoffZipCode {
		t.Errorf("Expected %q, got %q", TestDropoffZipCode, d.Dropoff.DetailedAddress.ZipCode)
	}
	if d.Dropoff.Location.Lat != TestDropoffLat {
		t.Errorf("Expected %q, got %q", TestDropoffLat, d.Dropoff.Location.Lat)
	}
	if d.Dropoff.Location.Lng != TestDropoffLng {
		t.Errorf("Expected %q, got %q", TestDropoffLng, d.Dropoff.Location.Lng)
	}
	if d.Dropoff.Name != TestDropoffName+", "+TestDropoffOptionalBusinessName {
		t.Errorf("Expected %q, got %q", TestDropoffName+", "+TestDropoffOptionalBusinessName, d.Dropoff.Name)
	}
	if d.Dropoff.Notes != TestDropoffOptionalNotes {
		t.Errorf("Expected %q, got %q", TestDropoffOptionalNotes, d.Dropoff.Notes)
	}
	if d.Dropoff.PhoneNumber != TestDropoffPhoneNumber {
		t.Errorf("Expected %q, got %q", TestDropoffPhoneNumber, d.Dropoff.PhoneNumber)
	}

	if d.Manifest.Description != TestManifestDescription {
		t.Errorf("Expected %q, got %q", d.Manifest.Description, TestManifestDescription)
	}
	if d.Manifest.Reference != TestManifestReference {
		t.Errorf("Expected %q, got %q", d.Manifest.Reference, TestManifestReference)
	}

	if len(d.RelatedDeliveries) != 0 {
		t.Errorf("Expected 0 RelatedDeliveries, got %d", len(d.RelatedDeliveries))
	}

}

func TestCancelDelivery(t *testing.T) {

	client := NewClient(TestCustomerId, TestAPIKey, TestTimeout)

	da, err := client.GetDeliveries(AllFilter, 1)
	if err != nil {
		t.Error(err)
	}

	newd, err := client.CancelDelivery(da[0].ID)
	if err != nil {
		t.Error(err)
	}
	if newd.Status != StatusCanceled {
		t.Errorf("Expected %q, got %q", StatusCanceled, newd.Status)
	}

}

func TestReturnDelivery(t *testing.T) {

	// t.Skip("jk")

	var (
		client   = NewClient(TestCustomerId, TestAPIKey, TestTimeout)
		manifest = NewManifest(TestManifestDescription, TestManifestReference)
		pickup   = NewDeliverySpot(TestPickupName, TestPickupAddress, TestPickupPhoneNumber)
		dropoff  = NewDeliverySpot(TestDropoffName, TestDropoffAddress, TestDropoffPhoneNumber)
		da       []*Delivery
	)

	quote, err := client.GetQuote(TestPickupAddress, TestDropoffAddress)
	if err != nil {
		t.Error(err)
	}

	// create a delivery to return
	err = client.CreateDelivery(manifest, pickup, dropoff, quote)
	if err != nil {
		t.Error(err)
	}

	// wait for it to get picked up
	// this bit could be handled by the webhook event trigger in a production scenario
	fmt.Println("Delivery created, waiting for dropoff state before triggering return (~100sec)")
	for {
		time.Sleep(2 * time.Second)
		fmt.Print(".")
		da, err = client.GetOngoingDeliveries(1)
		if err != nil {
			t.Error(err)
		}
		if len(da) > 0 && da[0].Status == StatusDropoff {
			break
		}
	}
	fmt.Println()

	// trigger return
	d, err := client.ReturnDelivery(da[0].ID)
	if err != nil {
		t.Error(err)
	}

	if d.ID == da[0].ID {
		t.Errorf("Expected a new delivery object")
	}

	// cancel it and call it a day
	_, err = client.CancelDelivery(d.ID)
	if err != nil {
		t.Error(err)
	}

}
