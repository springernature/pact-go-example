package main

import (
	"net/http"
	"fmt"
	"bytes"
	"io/ioutil"
)

func main() {
	var jsonStr = []byte(`{"s":"hello, world"}`)
	req, err := http.NewRequest("POST", "http://localhost:8080/uppercase", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
