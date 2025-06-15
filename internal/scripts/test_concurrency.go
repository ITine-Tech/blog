package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type UpdatePostPayload struct {
	Title *string `json:"title" //validate:"omitempty,max=100"`
	Text  *string `json:"text" //validate:"omitempty,max=1000"`
}

func updatePost(postID int, p UpdatePostPayload, wg *sync.WaitGroup) {
	defer wg.Done()

	// Construct the URL for the update endpoint
	url := fmt.Sprintf("http://localhost:3000/posts/%d", postID)

	// Create the JSON payload
	b, _ := json.Marshal(p)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers as needed, for example:
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Update response status:", resp.Status)
}

func main() {
	var wg sync.WaitGroup

	// Assuming the post ID to update is 1
	postID := 9

	// Simulate User A and User B updating the same post concurrently
	wg.Add(2)
	title := "NEW TITLE FROM USER A"
	text := "NEW CONTENT FROM USER B"

	go updatePost(postID, UpdatePostPayload{Title: &title}, &wg)
	go updatePost(postID, UpdatePostPayload{Text: &text}, &wg)
	wg.Wait()
}
