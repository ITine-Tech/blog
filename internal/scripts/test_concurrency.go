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

// updatePost sends a PATCH request to update a post with the given postID.
// It uses a WaitGroup to synchronize goroutines.
// The payload can contain either a new title or text, or both.
// This is a simulation of concurrent updates to the same post by different users.
func updatePost(postID int, p UpdatePostPayload, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("http://localhost:3000/posts/%d", postID)

	b, _ := json.Marshal(p)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

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

	postID := 1

	// Simulate User A and User B updating the same post concurrently
	wg.Add(2)
	title := "NEW TITLE FROM USER A"
	text := "NEW CONTENT FROM USER B"

	go updatePost(postID, UpdatePostPayload{Title: &title}, &wg)
	go updatePost(postID, UpdatePostPayload{Text: &text}, &wg)
	wg.Wait()
}
