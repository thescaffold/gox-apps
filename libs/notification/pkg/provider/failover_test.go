package provider

import (
	"testing"
	"time"
)

func TestSplitProviders(t *testing.T) {
	got := splitProviders(" termii , africastalking,, ")
	if len(got) != 2 || got[0] != "termii" || got[1] != "africastalking" {
		t.Fatalf("%v", got)
	}
}

func TestSMSFailsOverToTheNextProviderAndSingleNamesBehaveAsBefore(t *testing.T) {
	resetCircuits()
	var order []string
	RegisterSMSProvider("down", func(SMSMessage) bool { order = append(order, "down"); return false })
	RegisterSMSProvider("up", func(SMSMessage) bool { order = append(order, "up"); return true })
	s := &ProviderService{}

	if !s.smsProvider("down,up", SMSMessage{To: "+234"}) {
		t.Fatal("should have failed over to the second provider")
	}
	if len(order) != 2 || order[0] != "down" || order[1] != "up" {
		t.Fatalf("order %v", order)
	}
	order = nil
	if s.smsProvider("down", SMSMessage{}) {
		t.Fatal("a single failing provider must still report failure")
	}
	if len(order) != 1 {
		t.Fatalf("%v", order)
	}
}

func TestBreakerSkipsAFailingProviderThenRetriesAfterCooldown(t *testing.T) {
	resetCircuits()
	t.Setenv("NOTIFICATION_BREAKER_FAILS", "2")
	t.Setenv("NOTIFICATION_BREAKER_COOLDOWN_SECS", "30")
	now := time.Unix(1000, 0)
	nowFn = func() time.Time { return now }
	defer func() { nowFn = time.Now }()

	calls := map[string]int{}
	RegisterEmailProvider("flaky", func(EmailMessage) bool { calls["flaky"]++; return false })
	RegisterEmailProvider("steady", func(EmailMessage) bool { calls["steady"]++; return true })
	s := &ProviderService{}

	s.emailProvider("flaky,steady", EmailMessage{})
	s.emailProvider("flaky,steady", EmailMessage{}) // second failure opens the circuit
	before := calls["flaky"]
	s.emailProvider("flaky,steady", EmailMessage{})
	if calls["flaky"] != before {
		t.Fatal("an open circuit was still called")
	}
	if calls["steady"] != 3 {
		t.Fatalf("steady should have carried all three: %v", calls)
	}
	now = now.Add(31 * time.Second)
	s.emailProvider("flaky,steady", EmailMessage{})
	if calls["flaky"] != before+1 {
		t.Fatal("no trial after the cooldown")
	}
}

func TestWhenEveryCircuitIsOpenTheFirstIsStillTried(t *testing.T) {
	resetCircuits()
	t.Setenv("NOTIFICATION_BREAKER_FAILS", "1")
	n := 0
	RegisterEmailProvider("only", func(EmailMessage) bool { n++; return false })
	s := &ProviderService{}
	s.emailProvider("only", EmailMessage{})
	s.emailProvider("only", EmailMessage{})
	if n != 2 {
		t.Fatalf("a message must not be dropped silently: %d attempts", n)
	}
}

func TestCaptureStoresMessagesButIsRefusedInProduction(t *testing.T) {
	ClearCaptures()
	t.Setenv("NODE_ENV", "development")
	s := &ProviderService{}
	if !s.smsProvider("capture", SMSMessage{To: "+2348012345678", Text: "Your code is 123456"}) {
		t.Fatal("capture refused outside production")
	}
	s.emailProvider("capture", EmailMessage{To: "a@example.test", Subject: "Hi", HTML: "<b>x</b>"})
	got := Captures("+2348012345678", 10)
	if len(got) != 1 || got[0].Text != "Your code is 123456" || got[0].Channel != "sms" {
		t.Fatalf("%+v", got)
	}
	if all := Captures("", 10); len(all) != 2 || all[0].Channel != "email" {
		t.Fatalf("newest first expected: %+v", all)
	}

	t.Setenv("NODE_ENV", "production")
	if s.smsProvider("capture", SMSMessage{To: "x"}) {
		t.Fatal("capture must be refused in production")
	}
	if CaptureEnabled() {
		t.Fatal("the read endpoint must be off in production")
	}
	t.Setenv("NODE_ENV", "development")
	t.Setenv("NOTIFICATION_CAPTURE_TOKEN", "")
	if CaptureEnabled() {
		t.Fatal("the read endpoint needs a token configured")
	}
	t.Setenv("NOTIFICATION_CAPTURE_TOKEN", "secret")
	if !CaptureEnabled() {
		t.Fatal("expected enabled")
	}
}
