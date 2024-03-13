/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"sigs.k8s.io/referencegrant-poc/cmd/controller"
	"sigs.k8s.io/referencegrant-poc/pkg/handlers"
	"sigs.k8s.io/referencegrant-poc/pkg/store"
)

func waitForFile(filePath string) {
	for {
		_, err := os.Stat(filePath)
		if err == nil {
			fmt.Println("file exists")
			break
		} else if os.IsNotExist(err) {
			// File doesn't exist yet, wait for a while before checking again
			time.Sleep(3 * time.Second)
		} else {
			// Some other error occurred, print and exit
			fmt.Println("Error:", err)
			return
		}
	}
}

func main() {
	authStore := store.NewAuthStore()

	http.HandleFunc("/health", handlers.HandleHealthCheck)
	http.HandleFunc("/authorize", handlers.AuthzHandler(authStore))
	go func() {
		fmt.Println("Starting server on port 8080...")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	}()

	waitForFile(os.Getenv("KUBECONFIG"))

	controller.NewController(authStore)

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// go ctrl.Start(ctx)

}
