package ghostmates

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type (
	Client struct {
		customer_id string
		client      *http.Client
	}

	DeliverySpot struct {
		Name         string
		BusinessName string
		Address      string
		PhoneNumber  string
		Notes        string
	}

	DeliveryQuote struct {
		Kind            string     `json:"kind"`
		ID              string     `json:"id"`
		Created         *time.Time `json:"created"`
		Expires         *time.Time `json:"expires"`     // Date/Time after which the quote will no longer be accepted.
		DropoffEta      *time.Time `json:"dropoff_eta"` // Estimated drop-off time. This value may increase to several hours if the postmates platform is in high demand.
		DurationSeconds int        `json:"duration"`    // Estimated minutes for this delivery to reach dropoff, this value can be highly variable based upon the current amount of capacity available.
		Currency        string     `json:"currency"`    // Currency the "amount" values are in. (see "Other Standards > Currency")
		Fee             int        `json:"fee"`         // Amount in cents that will be charged if this delivery is created (see "Other Standards > Currency")
	}

	Location struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}

	Courier struct {
		Name        string   `json:"name"`
		ImgHref     string   `json:"img_href"`
		PhoneNumber string   `json:"phone_number"`
		VehicleType string   `json:"vehicle_type"`
		Location    Location `json:"location"`
	}

	Address struct {
		City           string `json:"city"`
		Country        string `json:"country"`
		State          string `json:"state"`
		StreetAddress1 string `json:"street_address_1"`
		StreetAddress2 string `json:"street_address_2"`
		ZipCode        string `json:"zip_code"`
	}

	Spot struct {
		Address         string   `json:"address"`
		DetailedAddress Address  `json:"detailed_address"`
		Location        Location `json:"location"`
		Name            string   `json:"name"`
		Notes           string   `json:"notes"`
		PhoneNumber     string   `json:"phone_number"`
	}

	Manifest struct {
		Description string `json:"description"` // A free form body describing the package
		Reference   string `json:"reference"`   // Developer provided identifier for the courier to reference when picking up the package
	}

	RelatedDelivery struct {
		ID           string `json:"id"`           // Unique identifier
		Relationship string `json:"relationship"` // Indicating the nature of the job identified in related_deliveries
	}

	Delivery struct {
		Kind                     string            `json:"kind"`
		ID                       string            `json:"id"`
		Status                   string            `json:"status"`
		Created                  *time.Time        `json:"created"`
		Updated                  *time.Time        `json:"updated"`
		DropoffEta               *time.Time        `json:"dropoff_eta"`                 // Estimated time the courier will arrive at the dropoff location.
		PickupEta                *time.Time        `json:"pickup_eta"`                  // Estimated time the courier will arrive at the pickup location.
		DropoffDeadline          *time.Time        `json:"dropoff_deadline"`            // Based on the delivery window from the delivery quote. If the dropoff_eta goes beyond this dropoff_deadline, our customer service team will be notified. We may extend this value to indicate a known problem.
		Complete                 bool              `json:"complete"`                    // false if the delivery is ongoing, and you can expect additional updates.
		Courier                  Courier           `json:"courier"`                     // Delivery's courier information once the delivery is assigned to a courier
		Currency                 string            `json:"currency"`                    // Currency the "amount" values are in. (see "Other Standards > Currency")
		CustomerSignatureImgHref string            `json:"customer_signature_img_href"` // A link to an image of the delivery confirmation signature
		Dropoff                  Spot              `json:"dropoff"`
		DropoffIdentifier        string            `json:"dropoff_identifier"` // This field identifies who received delivery at dropoff location
		Fee                      int               `json:"fee"`                // Amount in cents that will be charged if this delivery is created (see "Other Standards > Currency")
		Manifest                 Manifest          `json:"manifest"`           // An invoice of goods and its identifier
		Pickup                   Spot              `json:"pickup"`
		QuoteID                  string            `json:"quote_id"`           // ID for the Delivery Quote if one was provided when creating this Delivery
		RelatedDeliveries        []RelatedDelivery `json:"related_deliveries"` // A collection describing other jobs that share an association

		// testing flag from Postmates
		LiveMode bool `json:"live_mode"`
	}

	Deliveries struct {
		Object     string      `json:"object"`
		URL        string      `json:"url"`
		NextHref   string      `json:"next_href"`
		TotalCount int         `json:"total_count"`
		Data       []*Delivery `json:"data"`
	}
)

const (
	APIVersion = "20150519"
	APIHost    = "api.postmates.com"

	StatusPending        = "pending"         // We've accepted the delivery and will be assigning it to a courier.
	StatusPickup         = "pickup"          // Courier is assigned and is en route to pick up the items
	StatusPickupComplete = "pickup_complete" // Courier has picked up the items
	StatusDropoff        = "dropoff"         // Courier is moving towards the dropoff
	StatusCanceled       = "canceled"        // Items won't be delivered. Deliveries are either canceled by the customer or by our customer service team.
	StatusDelivered      = "delivered"       // Items were delivered successfully.
	StatusReturned       = "returned"        // The delivery was canceled and a new job created to return items to sender. (See related_deliveries in delivery object.)

	VehicleBicycle    = "bicycle"
	VehicleCar        = "car"
	VehicleVan        = "van"
	VehicleTruck      = "truck"
	VehicleScooter    = "scooter"
	VehicleMotorcycle = "motorcycle"

	RelatedDeliveryRelationshipTypeOriginal = "original" // the indicated job is the forward leg of the relationsihp
	RelatedDeliveryRelationshipTypeReturned = "returned" // the indicated job is the return leg of the relationship

	DeliveryQuoteKind = "delivery_quote"
	DeliveryKind      = "delivery"

	OngoingFilter = "ongoing"
	AllFilter     = ""

	AllDeliveries = -1
)

func NewManifest(description, reference string) *Manifest {
	return &Manifest{
		Description: description,
		Reference:   reference,
	}
}

// optional params not represented in signature
func NewDeliverySpot(name, address, phonenumber string) *DeliverySpot {
	return &DeliverySpot{
		Name:        name,
		Address:     address,
		PhoneNumber: phonenumber,
	}
}

func NewClient(customer_id, api_key string, timeout time.Duration) *Client {

	return &Client{
		customer_id: customer_id,
		client: &http.Client{
			Transport: postmatesTransport(func(req *http.Request) (*http.Response, error) {
				req.Host = APIHost
				req.URL.Host = APIHost
				req.URL.Scheme = "https"
				req.Header.Set("X-Postmates-Version", APIVersion)
				req.SetBasicAuth(api_key, "")
				return http.DefaultTransport.RoundTrip(req)
			}),
			Timeout: timeout,
		},
	}

}

func (c *Client) GetQuote(pickup_address, dropoff_address string) (*DeliveryQuote, error) {

	// POST /v1/customers/:customer_id/delivery_quotes

	// pickup_address="20 McAllister St, San Francisco, CA"
	// dropoff_address="101 Market St, San Francisco, CA"

	// You'll receive a DeliveryQuote response.

	resp, err := c.client.PostForm(
		"/v1/customers/"+url.QueryEscape(c.customer_id)+"/delivery_quotes",
		url.Values{
			"pickup_address":  []string{pickup_address},
			"dropoff_address": []string{dropoff_address},
		},
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewError(resp)
	}

	dq := &DeliveryQuote{}
	if err := json.NewDecoder(resp.Body).Decode(dq); err != nil {
		return nil, err
	}

	return dq, nil

}

func (c *Client) CreateDelivery(manifest *Manifest, pickup, dropoff *DeliverySpot, quote *DeliveryQuote) error {

	// POST /v1/customers/:customer_id/deliveries

	// manifest="a box of kittens"
	// manifest_reference="Optional reference that identifies the box of kittens"
	// pickup_name="The Warehouse"
	// pickup_address="20 McAllister St, San Francisco, CA"
	// pickup_phone_number="555-555-5555"
	// pickup_business_name="Optional Pickup Business Name, Inc."
	// pickup_notes="Optional note that this is Invoice #123"
	// dropoff_name="Alice"
	// dropoff_address="101 Market St, San Francisco, CA"
	// dropoff_phone_number="415-555-1234"
	// dropoff_business_name="Optional Dropoff Business Name, Inc."
	// dropoff_notes="Optional note to ring the bell"
	// quote_id=qUdje83jhdk

	resp, err := c.client.PostForm(
		"/v1/customers/"+url.QueryEscape(c.customer_id)+"/deliveries",
		url.Values{
			"manifest":              []string{manifest.Description},
			"manifest_reference":    []string{manifest.Reference},
			"pickup_name":           []string{pickup.Name},
			"pickup_address":        []string{pickup.Address},
			"pickup_phone_number":   []string{pickup.PhoneNumber},
			"pickup_business_name":  []string{pickup.BusinessName},
			"pickup_notes":          []string{pickup.Notes},
			"dropoff_name":          []string{dropoff.Name},
			"dropoff_address":       []string{dropoff.Address},
			"dropoff_phone_number":  []string{dropoff.PhoneNumber},
			"dropoff_business_name": []string{dropoff.BusinessName},
			"dropoff_notes":         []string{dropoff.Notes},
			"quote_id":              []string{quote.ID},
		},
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewError(resp)
	}

	return nil

}

func (c *Client) GetOngoingDeliveries(n int) ([]*Delivery, error) {
	return c.GetDeliveries(OngoingFilter, n)
}

func (c *Client) GetDeliveries(filter string, n int) ([]*Delivery, error) {

	// GET /v1/customers/:customer_id/deliveries
	// This endpoint currently supports one query argument:
	// ?filter=ongoing

	var (
		da []*Delivery
		u  = "/v1/customers/" + url.QueryEscape(c.customer_id) + "/deliveries?filter=" + url.QueryEscape(filter)
	)

	for {
		resp, err := c.client.Get(u)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			defer resp.Body.Close()
			return nil, NewError(resp)
		}

		ds := &Deliveries{}
		if err := json.NewDecoder(resp.Body).Decode(ds); err != nil {
			resp.Body.Close()
			return nil, err
		}

		// no defer since we're in a loop
		resp.Body.Close()

		da = append(da, ds.Data...)

		// if n == AllDeliveries we go till the wheels fall off
		// sanity check ds.Data length
		if (n != AllDeliveries && len(da) >= n) || len(ds.NextHref) == 0 || len(ds.Data) == 0 {
			break
		}

		u = ds.NextHref
	}

	if n != AllDeliveries && len(da) > n {
		da = da[:n]
	}

	return da, nil

}

func (c *Client) GetDelivery(delivery_id string) (*Delivery, error) {

	// GET /v1/customers/:customer_id/deliveries/:delivery_id
	// Returns: Delivery Object

	resp, err := c.client.Get("/v1/customers/" + url.QueryEscape(c.customer_id) + "/deliveries/" + url.QueryEscape(delivery_id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewError(resp)
	}

	d := &Delivery{}
	if err := json.NewDecoder(resp.Body).Decode(d); err != nil {
		return nil, err
	}

	return d, nil

}

func (c *Client) CancelDelivery(delivery_id string) (*Delivery, error) {

	// POST /v1/customers/:customer_id/deliveries/:delivery_id/cancel
	// Returns: Delivery Object

	resp, err := c.client.PostForm("/v1/customers/"+url.QueryEscape(c.customer_id)+"/deliveries/"+url.QueryEscape(delivery_id)+"/cancel", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewError(resp)
	}

	d := &Delivery{}
	if err := json.NewDecoder(resp.Body).Decode(d); err != nil {
		return nil, err
	}

	return d, nil

}

func (c *Client) ReturnDelivery(delivery_id string) (*Delivery, error) {

	// POST /v1/customers/:customer_id/deliveries/:delivery_id/return
	// Returns: Delivery Object (the new return delivery)

	resp, err := c.client.PostForm("/v1/customers/"+url.QueryEscape(c.customer_id)+"/deliveries/"+url.QueryEscape(delivery_id)+"/return", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewError(resp)
	}

	d := &Delivery{}
	if err := json.NewDecoder(resp.Body).Decode(d); err != nil {
		return nil, err
	}

	return d, nil

}
