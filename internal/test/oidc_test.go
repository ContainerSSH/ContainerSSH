package test //nolint:testpackage

import (
    "testing"
)

func TestCreateJWT(t *testing.T) {
    data := map[string]interface{}{
        "sub":  "1234567890",
        "name": "Test User",
        "iat":  1660371788,
    }
    jwt := createJWT(data, []byte("Aemixeidairee6ohth0Paesh1Baey7Tu"))
    if jwt != "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE2NjAzNzE3ODgsIm5hbWUiOiJUZXN0IFVzZXIiLCJzdWIiOiIxMjM0N"+
        "TY3ODkwIn0.SwNzxU5P3beHJ0MeoTEnvmshfrQrBF9YRHu-YAjwCNI" {
        t.Fatalf("Incorrect JWT token: %s", jwt)
    }
}
