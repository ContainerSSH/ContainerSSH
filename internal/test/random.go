package test

import (
    "math/rand"
    "strings"
    "time"
)

func RandomString(length int) string {
    // We are only using this user for testing purposes, so no security is required.
    random := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
    runes := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
        "abcdefghijklmnopqrstuvwxyz" +
        "0123456789!%$#-_=+")
    var randomStringBuilder strings.Builder
    for i := 0; i < length; i++ {
        randomStringBuilder.WriteRune(runes[random.Intn(len(runes))])
    }
    return randomStringBuilder.String()
}
