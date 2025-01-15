package main

import (
    "log"
    "github.com/oauth2-proxy/mockoidc"
)

func main() {
    // Start the mock OIDC server
    m, err := mockoidc.Run()
    if err != nil {
        log.Fatalf("Failed to start mock OIDC server: %v", err)
    }
    defer m.Shutdown()

    // Retrieve the issuer URL
    issuer := m.Issuer()

    // Output the issuer URL for reference
    log.Printf("Mock OIDC server is running at: %s", issuer)

    // Example: Integrate with your application or run tests here
    // ...

    // Keep the server running
    select {}
}

