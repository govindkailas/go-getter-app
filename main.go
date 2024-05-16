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
	vaultToken := os.Getenv("VAULT_TOKEN")
	if vaultToken == "" {
		vaultToken = "root"
		log.Println("VAULT_TOKEN environment variable not set, defaulting to ", vaultToken)
	}

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
		log.Println("PORT environment variable not set, defaulting to ", port)
	}

	vaultUrl := os.Getenv("VAULT_ADDR")
	if vaultUrl == "" {
		vaultUrl = "http://vault:8200"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received Request on default handler /")

		// If the JWT path is setup then get the new token from Vault using the k8s Auth
		jwtPath := os.Getenv("JWT_PATH")
		if jwtPath != "" {
			jwtFile, err := ioutil.ReadFile(jwtPath)
			if err != nil {
				log.Println("Error reading JWT file at", jwtPath, ": ", err)
				return
			}

			jwt := string(jwtFile)
			log.Println("Read JWT:", jwt)

			authPath := "auth/kubernetes/login"

			approle := os.Getenv("APPROLE")
			if approle == "" {
				approle = "webapp"
				log.Println("APPROLE environment variable not set, defaulting to ", approle)
			}

			// Create the payload for Vault authentication
			pl := VaultJWTPayload{Role: approle, JWT: jwt}
			jwtPayload, err := json.Marshal(pl)
			if err != nil {
				log.Println("Error encoding Vault request JSON:", err)
				return
			}

			// Send a request to Vault to retrieve a token
			vaultLoginResponse := &VaultLoginResponse{}
			if err != nil {
				log.Println("Error sending request to Vault:", err)
				if res.StatusCode == http.StatusForbidden {
					log.Println("Invalid JWT token, please check if the token is expired or if you have access to the requested resource.")
				}
				return err
			}

			err = SendRequest(vaultUrl+authPath, "", "POST", jwtPayload, &vaultLoginResponse)
			if err != nil {
				log.Println("Error sending request to Vault:", err)
				return err
			}

			vaultToken = vaultLoginResponse.Auth.ClientToken
			log.Println("Retrieved vault login token: ", vaultToken)
		}

		secretsPath := os.Getenv("SECRET_PATH")
		if secretsPath == "" {
			secretsPath = "secret/data/webapp/config"
			log.Println("SECRET_PATH environment variable not set, defaulting to ", secretsPath)
		}

		if vaultSecretResponse.Data == nil {
			log.Println("Error getting secret from Vault: empty response")
			return}

		secretResponseData, ok := vaultSecretResponse.Data.Data.(map[string]interface{})
		if ok {
			for key, value := range secretResponseData {
				// log.Println(w, "%s:%s ", key, value)
				fmt.Fprintf(w, "%s:%s ", key, value)
			}
		} else {
			log.Println("Error getting the secret from Vault, cannot convert Data to map[string]interface{}")
		}
	})

	//Add one more handler for getting a secret from Vault based on a userid passed in the POST call
	http.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		log.Println("Received Request on hander /secret with POST method")
		//Parse the userid from the request body
		decoder := json.NewDecoder(r.Body)
		var request struct {
			// suppose the request body is {"userId": "1234567890"}
			UserId string `json:"userId"`
		}
		err := decoder.Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//Send a request to Vault to retrieve the secret for given userid
		secretPath := "go-app/data/" + request.UserId
		vaultSecretResponse := &VaultSecretResponse{}
		err = SendRequest(vaultUrl+"/v1/"+secretPath, vaultToken, "GET", nil, &vaultSecretResponse)
		if err != nil {
			log.Println("Error getting secret from Vault:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//Write the secret response to the client
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.Encode(vaultSecretResponse)
		//log the response
		log.Println("Responding back to client with: ", vaultSecretResponse)
	})

	log.Println("Listening on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server:", err)
	}
}

func SendRequest(url string, token string, requestType string, payload []byte, target interface{}) error {
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Vault-Token", token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request to Vault:", err)
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(target)
}
