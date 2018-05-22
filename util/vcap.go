package util

import (
	"encoding/json"
	"fmt"
)

// ParseVcapServices parses raw JSON VCAP_SERVICES into a useable object
func ParseVcapServices(data []byte) (*VcapServices, error) {
	services := VcapServices{}
	err := json.Unmarshal(data, &services)
	return &services, err
}

// VcapServices is a parsed VCAP_SERVICES JSON configuration
type VcapServices map[string][]VcapService

// FindServiceByName finds a service within VCAP_SERVICES, wherever it is nestled
func (s VcapServices) FindServiceByName(name string) *VcapService {
	for _, serviceArray := range s {
		for _, service := range serviceArray {
			if service.Name == name {
				return &service
			}
		}
	}
	return nil
}

// VcapService is a parsed individual VCAP service; not all fields are parsed here
type VcapService struct {
	Name        string          `json:"name"`
	Credentials VcapCredentials `json:"credentials"`
}

// VcapCredentials is a parsed map of VCAP credentials for a service
type VcapCredentials map[string]interface{}

// String recovers the value at the given key, assuming it is a string
func (c VcapCredentials) String(key string) (string, error) {
	if val, ok := c[key]; !ok {
		return "", fmt.Errorf("Credential key does not exist: %s", key)
	} else if valStr, ok := val.(string); ok {
		return valStr, nil
	} else {
		return "", fmt.Errorf("Could not convert value to string: key=%s, value=%v", key, val)
	}
}

// Int recovers the value at the given key, assuming it is an int
func (c VcapCredentials) Int(key string) (int, error) {
	if val, ok := c[key]; !ok {
		return 0, fmt.Errorf("Credential key does not exist: %s", key)
	} else if valInt, ok := val.(int); ok {
		return valInt, nil
	} else {
		return 0, fmt.Errorf("Could not convert value to int: key=%s, value=%v", key, val)
	}

}
