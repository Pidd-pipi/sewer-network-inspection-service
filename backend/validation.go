package main

import "fmt"

func validateInspectionStatus(status string) error {
	switch status {
	case "open", "scheduled", "resolved", "closed":
		return nil
	default:
		return fmt.Errorf("status must be open, scheduled, or resolved")
	}
}
