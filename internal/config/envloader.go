package config

import (
    "os"
    "fmt"
    "encoding/json"
    "log"
    "github.com/joho/godotenv"
)

func LoadEnv(filepaths ...string) error {
    err := godotenv.Load(filepaths...)
    if err != nil && !os.IsNotExist(err) {
        log.Printf("Error loading the .env file: %v", err)
        return err
    }
    return nil
}

func GetEnv(key string, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func DumpConfigAsJSON(filename string) (string, error) {
    env, err := godotenv.Read(filename)
    if err != nil {
        return "", fmt.Errorf("error reading .env file: %w", err)
    }

    jsonBytes, err := json.MarshalIndent(env, "", "  ")
    if err != nil {
        return "", fmt.Errorf("error marshaling config to JSON: %w", err)
    }
    return string(jsonBytes), nil
}

