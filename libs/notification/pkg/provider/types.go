package provider

type EmailMessage struct {
	From    string
	To      string
	Subject string
	Text    string
	HTML    string
}

type SMSMessage struct {
	From    string
	To      string
	Subject string
	Text    string
	HTML    string
}

type WebMessage struct {
	To      string
	Subject string
	HTML    string
}

type MobileMessage struct {
	To      string
	Subject string
	Body    string
}

type ChannelResult struct {
	Channel string
	Status  bool
}
