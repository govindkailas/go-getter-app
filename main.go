package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	vaultToken := "root"

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
		log.Println("PORT environment variable not set, defaulting to", port)
	}

	vaultUrl := os.Getenv("VAULT_ADDR")
	if vaultUrl == "" {
		vaultUrl = "http://vault:8200"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received Request - Port forwarding is working.")

		// If the JWT path is setup then get the new token from Vault using the k8s Auth
		jwtPath := os.Getenv("JWT_PATH")
		if jwtPath != "" {
			jwtFile, err := ioutil.ReadFile(jwtPath)
			if err != nil {
				fmt.Println("Error reading JWT file at", jwtPath, ": ", err)
				return
			}

			jwt := string(jwtFile)
			fmt.Println("Read JWT:", jwt)

			authPath := "auth/kubernetes/login"

			approle := os.Getenv("APPROLE")
			if approle == "" {
				approle = "webapp"
				log.Println("APPROLE environment variable not set, defaulting to", approle)
			}

			// Create the payload for Vault authentication
			pl := VaultJWTPayload{Role: approle, JWT: jwt}
			jwtPayload, err := json.Marshal(pl)
			if err != nil {
				fmt.Println("Error encoding Vault request JSON:", err)
				return
			}

			// Send a request to Vault to retrieve a token
			vaultLoginResponse := &VaultLoginResponse{}
			err = SendRequest(vaultUrl+"/v1/"+authPath, "", "POST", jwtPayload, vaultLoginResponse)
			if err != nil {
				fmt.Println("Error getting response from Vault k8s login:", err)
				fmt.Println("vault url is", vaultUrl+"/v1/"+authPath)
				fmt.Println("vault login payload is", string(jwtPayload))
				fmt.Println("vaultLoginResponse is ", vaultLoginResponse.Auth.ClientToken)
				return
			}

			vaultToken = vaultLoginResponse.Auth.ClientToken
			fmt.Println("Retrieved token: ", vaultToken)
		}

		secretsPath := os.Getenv("SECRET_PATH")
		if secretsPath == "" {
			secretsPath = "secret/data/webapp/config"
			log.Println("SECRET_PATH environment variable not set, defaulting to", secretsPath)
		}

		// Send a request to Vault using the token to retrieve the secret
		vaultSecretResponse := &VaultSecretResponse{}
		err := SendRequest(vaultUrl+"/v1/"+secretsPath, vaultToken, "GET", nil, &vaultSecretResponse)
		if err != nil {
			fmt.Println("Error getting secret from Vault:", err)
			return
		}

		secretResponseData, ok := vaultSecretResponse.Data.Data.(map[string]interface{})
		if ok {
			for key, value := range secretResponseData {
				fmt.Fprintf(w, "%s:%s ", key, value)
			}
		} else {
			fmt.Println("Error getting the secret from Vault, cannot convert Data to map[string]interface{}")
		}
	})

	log.Println("Listening on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server:", err)
	}
}

func SendRequest(url string, token string, requestType string, payload []byte, target interface{}) error {
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Vault-Token", token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to Vault:", err)
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(target)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"

// 	"github.com/hashicorp/vault/api"
// 	"github.com/hashicorp/vault/api/auth/kubernetes"
// )

// func main() {

// 	http.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
// 		// Call getSecretWithKubernetesAuth()
// 		secret, err := getSecretWithKubernetesAuth()
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		// Write secret to response
// 		fmt.Fprint(w, secret)
// 	})

// 	http.ListenAndServe(":8080", nil)

// }

// // Fetches a key-value secret (kv-v2) after authenticating to Vault with a Kubernetes service account.
// func getSecretWithKubernetesAuth() (string, error) {
// 	// If set, the VAULT_ADDR environment variable will be the address that
// 	// your pod uses to communicate with Vault.
// 	vaultUrl := os.Getenv("VAULT_ADDR")
// 	if vaultUrl == "" {
// 		vaultUrl = "http://vault:8200"
// 	}
// 	config := api.DefaultConfig() // modify for more granular configuration

// 	client, err := api.NewClient(config)
// 	if err != nil {
// 		return "", fmt.Errorf("unable to initialize Vault client: %w", err)
// 	}

// 	// The service-account token will be read from the path where the token's
// 	// Kubernetes Secret is mounted. By default, Kubernetes will mount it to
// 	// /var/run/secrets/kubernetes.io/serviceaccount/token, but an administrator
// 	// may have configured it to be mounted elsewhere.
// 	// In that case, we'll use the option WithServiceAccountTokenPath to look
// 	// for the token there.
// 	approle := os.Getenv("APPROLE")
// 	if approle == "" {
// 		approle = "webapp"
// 		log.Println("APPROLE environment variable not set, defaulting to", approle)
// 	}
// 	k8sAuth, err := kubernetes.NewKubernetesAuth(
// 		"go-app-role", // role name
// 		kubernetes.WithServiceAccountTokenPath("/var/run/secrets/kubernetes.io/serviceaccount/token"),
// 	)
// 	if err != nil {
// 		return "", fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
// 	}

// 	authInfo, err := client.Auth().Login(context.Background(), k8sAuth)
// 	if err != nil {
// 		return "", fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
// 	}
// 	if authInfo == nil {
// 		return "", fmt.Errorf("no auth info was returned after login")
// 	}

// 	// get secret from Vault, from the default mount path for KV v2 in dev mode, "secret"
// 	secretsPath := os.Getenv("SECRET_PATH")
// 	if secretsPath == "" {
// 		secretsPath = "secret/data/webapp/config"
// 		log.Println("SECRET_PATH environment variable not set, defaulting to", secretsPath)
// 	}
// 	secret, err := client.KVv2(secretsPath).Get(context.Background(), "creds")
// 	if err != nil {
// 		return "", fmt.Errorf("unable to read secret: %w", err)
// 	}
// 	//print secret
// 	fmt.Println("Secret from vault is:", secret.Data)
// 	// data map can contain more than one key-value pair,
// 	// in this case we're just grabbing one of them
// 	value, ok := secret.Data["password"].(string)
// 	if !ok {
// 		return "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
// 	}

// 	return value, nil
// }
