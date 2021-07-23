package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"kf/book"
)

const (
	sendEndpoint    string = "/api/v3/sendorder"
	cancelEndpoint  string = "/api/v3/cancelorder"
	openPosEndpoint string = "/api/v3/openpositions"

	pathPrefix string = "/derivatives"

	sendPath    string = pathPrefix + sendEndpoint
	cancelPath  string = pathPrefix + cancelEndpoint
	openPosPath string = pathPrefix + openPosEndpoint
)

type API struct {
	auth Auth
	http *http.Client
	sig  string
}

func New(auth Auth, sig string) API {
	return API{auth: auth, http: &http.Client{}, sig: sig}
}

func (api *API) GetOpenPos() (float64, error) {
	nonce, authent := api.auth.Authentication(openPosEndpoint, "")
	res, err := api.sendRequest(openPosPath, "", nonce, authent)
	if err != nil {
		return 0, err
	}

	var jsonRes map[string]interface{}
	json.NewDecoder(res.Body).Decode(&jsonRes)
	if jsonRes["result"].(string) != "success" {
		return 0, fmt.Errorf("failed to send an order: %s", jsonRes["error"].(string))
	}

	pos := jsonRes["openPositions"].([]interface{})
	if len(pos) == 0 {
		return 0, nil
	} else {
		bitcoin := pos[0].(map[string]interface{})
		qty := bitcoin["size"].(float64)
		if bitcoin["side"].(string) == "short" {
			return -qty, nil
		} else {
			return qty, nil
		}
	}
}

func (api *API) SendOrder(order book.Order) (orderID string, clientID string, err error) {
	clientID = api.newClientID()
	var side string
	if order.Type == book.Ask {
		side = "sell"
	} else {
		side = "buy"
	}

	post := "orderType=lmt&symbol=PI_XBTUSD&side=" + side + "&size=" + fmt.Sprintf("%f", order.Quantity) + "&limitPrice=" + fmt.Sprintf("%f", order.Price) + "&triggerSignal=mark&cliOrdId=" + clientID + "&reduceOnly=false"
	nonce, authent := api.auth.Authentication(sendEndpoint, post)
	res, err := api.sendRequest(sendPath, post, nonce, authent)
	if err != nil {
		return "", "", err
	}

	var jsonRes map[string]interface{}
	json.NewDecoder(res.Body).Decode(&jsonRes)
	if jsonRes["result"].(string) != "success" {
		return "", "", fmt.Errorf("failed to send an order: %s", jsonRes["error"].(string))
	}

	status := jsonRes["sendStatus"].(map[string]interface{})
	return status["order_id"].(string), status["cliOrdId"].(string), nil
}

func (api *API) CancelOrder(orderID string, clientID string) (bool, error) {
	post := "order_id=" + orderID + "&cliOrdId=" + clientID
	nonce, authent := api.auth.Authentication(cancelEndpoint, post)
	res, err := api.sendRequest(cancelPath, post, nonce, authent)
	if err != nil {
		return false, err
	}

	var jsonRes map[string]interface{}
	json.NewDecoder(res.Body).Decode(&jsonRes)
	if jsonRes["result"].(string) != "success" {
		return false, fmt.Errorf("failed to cancel an order: %s", jsonRes["error"].(string))
	}

	status := jsonRes["cancelStatus"].(map[string]interface{})
	return status["status"].(string) != "notFound", nil
}

func (api *API) newClientID() string {
	return api.sig + strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
}

func (api *API) sendRequest(path string, postData string, nonce string, authent string) (*http.Response, error) {
	url := RestURL + path
	if len(postData) > 0 {
		url += "?" + postData
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("APIKey", api.auth.APIKey)
	req.Header.Set("Nonce", nonce)
	req.Header.Set("Authent", authent)
	return api.http.Do(req)
}
