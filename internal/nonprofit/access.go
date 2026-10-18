package nonprofit

import "fmt"

type Role string

const (
	Donor     Role = "donor"
	Volunteer Role = "volunteer"
	Organizer Role = "organizer"
)

type Access struct {
	ReceiptDelivery   bool `json:"receipt_delivery"`
	VolunteerReminder bool `json:"volunteer_reminder"`
	CampaignReporting bool `json:"campaign_reporting"`
}

func AccessFor(role Role) (Access, error) {
	switch role {
	case Donor:
		return Access{ReceiptDelivery: true}, nil
	case Volunteer:
		return Access{VolunteerReminder: true}, nil
	case Organizer:
		return Access{CampaignReporting: true}, nil
	default:
		return Access{}, fmt.Errorf("unknown nonprofit role %q", role)
	}
}
