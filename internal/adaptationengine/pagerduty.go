package adaptationengine

import (
	"context"
	"fmt"
	"github.com/PagerDuty/go-pagerduty"
)

type UrgencyLevel string

var (
	HighUrgencyLevel UrgencyLevel = "high"
	LowUrgencyLevel  UrgencyLevel = "low"
)

func (s Client) RaisePager(urgencyLevel UrgencyLevel, message string) error {
	from := s.pagerDutySender
	_, err := s.pagerDutyClient.CreateIncidentWithContext(context.Background(), from, &pagerduty.CreateIncidentOptions{
		Title:   "Sensitive Incident",
		Urgency: string(urgencyLevel),
		Service: &pagerduty.APIReference{
			ID:   "P03T0Q9",
			Type: "service_reference",
		},
		Body: &pagerduty.APIDetails{
			Type:    "incident_body",
			Details: message,
		},
	})
	if err != nil {
		return fmt.Errorf("error occurred while creating the pager corresponding to '%s': %w", from, err)
	}
	return nil
}
