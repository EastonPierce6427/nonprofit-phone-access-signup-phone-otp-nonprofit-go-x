package nonprofit

import "testing"

func TestAccessFor(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want Access
	}{
		{name: "donor receives receipts", role: Donor, want: Access{ReceiptDelivery: true}},
		{name: "volunteer receives reminders", role: Volunteer, want: Access{VolunteerReminder: true}},
		{name: "organizer reads campaign reports", role: Organizer, want: Access{CampaignReporting: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AccessFor(tt.role)
			if err != nil {
				t.Fatalf("AccessFor(%q): %v", tt.role, err)
			}
			if got != tt.want {
				t.Fatalf("AccessFor(%q) = %+v, want %+v", tt.role, got, tt.want)
			}
		})
	}
}

func TestAccessForRejectsUnknownRole(t *testing.T) {
	if _, err := AccessFor("board-member"); err == nil {
		t.Fatal("AccessFor accepted an unknown role")
	}
}
