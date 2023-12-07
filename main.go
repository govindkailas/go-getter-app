// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"log"

	vault "github.com/hashicorp/vault/api"
)

func main() {
	config := vault.DefaultConfig()

	config.Address = "http://127.0.0.1:8200"

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	// Authenticate
	// WARNING: This quickstart uses the root token for our Vault dev server.
	// Don't do this in production!
	client.SetToken("dev-only-token")

	secretData := map[string]interface{}{
		"password": "Hashi123",
	}

	ctx := context.Background()

	// // Write a secret
	// _, err = client.KVv2("secret").Put(ctx, "my-secret-password", secretData)
	// if err != nil {
	// 	log.Fatalf("unable to write secret: %v", err)
	// }

	// log.Println("Secret written successfully.")

	// // Read a secret
	// secret, err := client.KVv2("secret").Get(ctx, "my-secret-password")
	// if err != nil {
	// 	log.Fatalf("unable to read secret: %v", err)
	// }

	// value, ok := secret.Data["password"].(string)
	// if !ok {
	// 	log.Fatalf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
	// }

	// if value != "Hashi123" {
	// 	log.Fatalf("unexpected password value %q retrieved from vault", value)
	// }

	log.Println("Access granted!")
}


// package main

// import (
//     "fmt"
//     "os"

//     vault "github.com/hashicorp/vault/api"
//     auth "github.com/hashicorp/vault/api/auth/kubernetes"
// )

// // Fetches a key-value secret (kv-v2) after authenticating to Vault with a Kubernetes service account.
// // For a more in-depth setup explanation, please see the relevant readme in the hashicorp/vault-examples repo.
// func getSecretWithKubernetesAuth() (string, error) {
//     // If set, the VAULT_ADDR environment variable will be the address that
//     // your pod uses to communicate with Vault.
//     config := vault.DefaultConfig() // modify for more granular configuration

//     client, err := vault.NewClient(config)
//     if err != nil {
//         return "", fmt.Errorf("unable to initialize Vault client: %w", err)
//     }

//     // The service-account token will be read from the path where the token's
//     // Kubernetes Secret is mounted. By default, Kubernetes will mount it to
//     // /var/run/secrets/kubernetes.io/serviceaccount/token, but an administrator
//     // may have configured it to be mounted elsewhere.
//     // In that case, we'll use the option WithServiceAccountTokenPath to look
//     // for the token there.
//     k8sAuth, err := auth.NewKubernetesAuth(
//         "go-app-role",
//         auth.WithServiceAccountTokenPath("path/to/service-account-token"),
//     )
//     if err != nil {
//         return "", fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
//     }

//     authInfo, err := client.Auth().Login(context.TODO(), k8sAuth)
//     if err != nil {
//         return "", fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
//     }
//     if authInfo == nil {
//         return "", fmt.Errorf("no auth info was returned after login")
//     }

//     // get secret from Vault, from the default mount path for KV v2 in dev mode, "secret"
//     secret, err := client.KVv2("secret").Get(context.Background(), "creds")
//     if err != nil {
//         return "", fmt.Errorf("unable to read secret: %w", err)
//     }

//     // data map can contain more than one key-value pair,
//     // in this case we're just grabbing one of them
//     value, ok := secret.Data["password"].(string)
//     if !ok {
//         return "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
//     }

//     return value, nil
// }