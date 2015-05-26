package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jasonmoo/ghostmates"
)

const PostmatesClientTimeout = 5 * time.Second

var (
	host = flag.String("host", ":8080", "host:port to listen on")
	cert = flag.String("cert_file", "cert.pem", "key files for tls")
	key  = flag.String("key_file", "key.pem", "key files for tls")

	PostmatesAPIKey    = os.Getenv("POSTMATES_API_KEY")
	GoogleSearchAPIKey = os.Getenv("GOOGLE_API_KEY")
)

func main() {

	flag.Parse()

	wh := ghostmates.NewWebhook()

	go func() {
		for {
			select {
			case e := <-wh.Events.DeliveryStatus:
				fmt.Printf("Received %T event %+v\n", e, e)
			case e := <-wh.Events.DeliveryDeadline:
				fmt.Printf("Received %T event %+v\n", e, e)
			case e := <-wh.Events.CourierUpdate:
				fmt.Printf("Received %T event %+v\n", e, e)
			case e := <-wh.Events.DeliveryReturn:
				fmt.Printf("Received %T event %+v\n", e, e)
			}
		}
	}()

	http.HandleFunc("/_postmates/34f1d48945d369e527701c215901e1519537c7aa", wh.Handler)

	http.HandleFunc("/donate", func(w http.ResponseWriter, req *http.Request) {

		var (
			customer_id    = strings.TrimSpace(req.FormValue("customer_id"))
			pickup_address = strings.TrimSpace(req.FormValue("pickup_address"))
			phone          = strings.TrimSpace(req.FormValue("phone"))
			items          = strings.TrimSpace(req.FormValue("items"))
		)

		if len(customer_id) == 0 || len(pickup_address) == 0 || len(items) == 0 || len(phone) == 0 {
			http.Error(w, "customer_id, pickup_address, phone and items are required fields", http.StatusBadRequest)
			return
		}

		nearby, err := getNearbyDonationCenter(pickup_address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		client := ghostmates.NewClient(PostmatesAPIKey, customer_id, PostmatesClientTimeout)

		quote, err := client.GetQuote(pickup_address, nearby.Results[0].FormattedAddress)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		manifest := ghostmates.NewManifest(items, "Donation Goods")
		pickup := ghostmates.NewDeliverySpot("Donation pickup spot", pickup_address, phone)
		dropoff := ghostmates.NewDeliverySpot(nearby.Results[0].Name, nearby.Results[0].FormattedAddress, phone)

		if err := client.CreateDelivery(manifest, pickup, dropoff, quote); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	})

	log.Fatal(http.ListenAndServeTLS(*host, *cert, *key, nil))

}

type GoogleNearby struct {
	Results []struct {
		Name             string `json:"name"`
		FormattedAddress string `json:"formatted_address"`
	} `json:"results"`
}

func getNearbyDonationCenter(pickup_address string) (*GoogleNearby, error) {

	const GooglePlacesSearchURL = "https://maps.googleapis.com/maps/api/place/textsearch/json"

	vals := url.Values{
		"key":   []string{GoogleSearchAPIKey},
		"query": []string{"Donation Center near " + pickup_address},
	}

	resp, err := http.Get(GooglePlacesSearchURL + "?opennow&" + vals.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Received %d from Google Places API: %s", resp.StatusCode, data)
	}

	nearby := &GoogleNearby{}

	if err := json.NewDecoder(resp.Body).Decode(nearby); err != nil {
		return nil, err
	}

	if len(nearby.Results) == 0 {
		return nil, errors.New("Unable to find a donation center near your pickup location that is open now")
	}

	return nearby, nil

}
