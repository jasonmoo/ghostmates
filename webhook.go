package ghostmates

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"time"
)

type (
	Webhook struct {
		Handler http.HandlerFunc
		Events  Events
	}

	Events struct {
		DeliveryStatus   <-chan *DeliveryStatusEvent
		DeliveryDeadline <-chan *DeliveryDeadlineEvent
		CourierUpdate    <-chan *CourierUpdateEvent
		DeliveryReturn   <-chan *DeliveryReturnEvent
	}

	// delivery_status - Sent each time the status field on a delivery changes.
	// Event will contain the field status and include an updated delivery object as the data.
	DeliveryStatusEvent struct {
		ID         string     `json:"id"`
		Created    *time.Time `json:"created"`
		DeliveryID string     `json:"delivery_id"`
		LiveMode   bool       `json:"live_mode"`

		Status   string    `json:"status"`
		Delivery *Delivery `json:"data"`
	}

	// delivery_deadline - Sent when the delivery deadline for a delivery has changed.
	// Event will contain the field dropoff_deadline and include an updated delivery object as the data.
	DeliveryDeadlineEvent struct {
		ID         string     `json:"id"`
		Created    *time.Time `json:"created"`
		DeliveryID string     `json:"delivery_id"`
		LiveMode   bool       `json:"live_mode"`

		DropoffDeadline *time.Time `json:"dropoff_deadline"`
		Delivery        *Delivery  `json:"data"`
	}

	// courier_update - Sent periodicially as a courier changes locations.
	// Event will contain the field location with lat and lng fields. An udpated delivery object will be included as the data.
	CourierUpdateEvent struct {
		ID         string     `json:"id"`
		Created    *time.Time `json:"created"`
		DeliveryID string     `json:"delivery_id"`
		LiveMode   bool       `json:"live_mode"`

		Location Location  `json:"location"`
		Delivery *Delivery `json:"data"`
	}

	// delivery_return - Sent when a delivery has been returned to pickup location
	// Event will contain the field status and include an updated delivery object as the data.
	DeliveryReturnEvent struct {
		ID         string     `json:"id"`
		Created    *time.Time `json:"created"`
		DeliveryID string     `json:"delivery_id"`
		LiveMode   bool       `json:"live_mode"`

		Status   string    `json:"status"`
		Delivery *Delivery `json:"data"`
	}
)

const (
	DeliveryStatusEventKind   = "event.delivery_status"
	DeliveryDeadlineEventKind = "event.delivery_deadline"
	CourierUpdateEventKind    = "event.courier_update"
	DeliveryReturnEventKind   = "event.delivery_return"

	MaxBodyLength = 64 << 10 // 64kb max event payload size
)

var (
	DefaultBufferLength = 512

	ErrUnsupportedEventKind = errors.New("unsupported event kind")
)

func NewWebhook() *Webhook {

	DeliveryStatusEventChan := make(chan *DeliveryStatusEvent, DefaultBufferLength)
	DeliveryDeadlineEventChan := make(chan *DeliveryDeadlineEvent, DefaultBufferLength)
	CourierUpdateEventChan := make(chan *CourierUpdateEvent, DefaultBufferLength)
	DeliveryReturnEventChan := make(chan *DeliveryReturnEvent, DefaultBufferLength)

	wh := &Webhook{
		Events: Events{
			DeliveryStatus:   DeliveryStatusEventChan,
			DeliveryDeadline: DeliveryDeadlineEventChan,
			CourierUpdate:    CourierUpdateEventChan,
			DeliveryReturn:   DeliveryReturnEventChan,
		},
	}

	wh.Handler = func(w http.ResponseWriter, req *http.Request) {

		switch req.Method {
		case "POST":

			data, err := ioutil.ReadAll(&io.LimitedReader{R: req.Body, N: MaxBodyLength})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var k struct {
				Kind string `json:"kind"`
			}
			if err := json.Unmarshal(data, &k); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			switch k.Kind {
			case DeliveryStatusEventKind:
				v := &DeliveryStatusEvent{}
				if err := json.Unmarshal(data, v); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				select {
				case DeliveryStatusEventChan <- v:
				default:
					// writes should not block, discard overflow
				}
			case DeliveryDeadlineEventKind:
				v := &DeliveryDeadlineEvent{}
				if err := json.Unmarshal(data, v); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				select {
				case DeliveryDeadlineEventChan <- v:
				default:
					// writes should not block, discard overflow
				}
			case CourierUpdateEventKind:
				v := &CourierUpdateEvent{}
				if err := json.Unmarshal(data, v); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				select {
				case CourierUpdateEventChan <- v:
				default:
					// writes should not block, discard overflow
				}
			case DeliveryReturnEventKind:
				v := &DeliveryReturnEvent{}
				if err := json.Unmarshal(data, v); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				select {
				case DeliveryReturnEventChan <- v:
				default:
					// writes should not block, discard overflow
				}
			default:
				http.Error(w, ErrUnsupportedEventKind.Error(), http.StatusBadRequest)
				return
			}

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

	}

	return wh

}
