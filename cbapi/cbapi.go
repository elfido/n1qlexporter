package cbapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	tr     = &http.Transport{}
	client = &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
)

// Auth Generic Authentication holder
type Auth struct {
	Username string
	Password string
}

// ToMillis Convers a duration string to int64
func ToMillis(t string) int64 {
	parsed, err := time.ParseDuration(t)
	if err == nil {
		return int64(parsed) / 1000000
	}
	return -1
}

// GetAPI generic HTTP caller for GET operations
func GetAPI(url string, serverAuth *Auth) []byte {
	request, _ := http.NewRequest("GET", url, nil)
	request.SetBasicAuth(serverAuth.Username, serverAuth.Password)
	res, err := client.Do(request)
	if err != nil {
		fmt.Printf("Failed to scrap server %s", err.Error())
		return []byte{}
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		bytes, _ := ioutil.ReadAll(res.Body)
		return bytes
	}
	return []byte{}
}
