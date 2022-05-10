package model

import "time"

type ZdTickets struct {
	Tickets []Ticket `json:"tickets"`
}

type Ticket struct {
	ID                  int           `json:"id"`
	ExternalID          interface{}   `json:"external_id"`
	Via                 Via           `json:"via"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	Type                interface{}   `json:"type"`
	Subject             string        `json:"subject"`
	RawSubject          string        `json:"raw_subject"`
	Description         string        `json:"description"`
	Priority            interface{}   `json:"priority"`
	Status              string        `json:"status"`
	Recipient           interface{}   `json:"recipient"`
	RequesterID         int64         `json:"requester_id"`
	SubmitterID         int64         `json:"submitter_id"`
	AssigneeID          int64         `json:"assignee_id"`
	OrganizationID      int64         `json:"organization_id"`
	CollaboratorIds     []interface{} `json:"collaborator_ids"`
	FollowerIds         []interface{} `json:"follower_ids"`
	EmailCcIds          []interface{} `json:"email_cc_ids"`
	ForumTopicID        interface{}   `json:"forum_topic_id"`
	ProblemID           interface{}   `json:"problem_id"`
	HasIncidents        bool          `json:"has_incidents"`
	IsPublic            bool          `json:"is_public"`
	DueAt               interface{}   `json:"due_at"`
	Tags                []interface{} `json:"tags"`
	CustomFields        []interface{} `json:"custom_fields"`
	SatisfactionRating  interface{}   `json:"satisfaction_rating"`
	SharingAgreementIds []interface{} `json:"sharing_agreement_ids"`
	Fields              []interface{} `json:"fields"`
	FollowupIds         []interface{} `json:"followup_ids"`
	TicketFormID        int64         `json:"ticket_form_id"`
	AllowChannelback    bool          `json:"allow_channelback"`
	AllowAttachments    bool          `json:"allow_attachments"`
	GeneratedTimestamp  int           `json:"generated_timestamp"`
}

type Via struct {
	Channel string `json:"channel"`
	Source  Source `json:"source"`
}

type Source struct {
	From From        `json:"from"`
	To   To          `json:"to"`
	Rel  interface{} `json:"rel"`
}

type From struct{}
type To struct{}
