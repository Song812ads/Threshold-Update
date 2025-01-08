package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func GetRequest(c http.Client, url string) bool {

	resp, err := c.Get(url)
	var send = false

	if err != nil {
		// Handle the error (e.g., network issue, service down)
		fmt.Println("Waiting for service to begin...")
	} else {
		// Check if the response status code is 200
		if resp.StatusCode == 200 {
			send = true
			// fmt.Println("Send success")
			resp.Body.Close() // Close the response body when done
		} else {
			// If status code is not 200, handle the error or retry
			fmt.Printf("Received status %d. Retrying...\n", resp.StatusCode)
			resp.Body.Close() // Close the response body even on failure
		}
	}
	return send
}

func PostRequest(c http.Client, url string, data []byte) error {
	reader := bytes.NewReader(data)

	resp, err := c.Post(url, "application/json", reader)
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 { // 200 OK
		// WriteError(w, resp.StatusCode, fmt.Errorf("Received non-OK status code: %d", resp.StatusCode))

		// fmt.Println(resp.StatusCode)
		return fmt.Errorf("May have already exist")
	}

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to read response body: %v", err)
	// }

	return nil

}

func SendPutRequest(url string, data []byte) (string, error) {
	// Create a new HTTP client
	client := &http.Client{}

	// Create the PUT request
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create PUT request: %v", err)
	}

	// Set headers if necessary (e.g., Content-Type)
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send PUT request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Return the response as a string
	return string(body), nil
}
