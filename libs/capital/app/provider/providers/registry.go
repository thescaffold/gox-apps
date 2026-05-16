package providers

import (
	capitalpayment "github.com/thescaffold/gox-apps/libs/capital/app/payment"
)

// Register installs Stripe / Paystack / Flutterwave into the supplied
// PaymentService. Host code calls this once after DI completes:
//
//	providers.Register(paymentSvc)
//
// Each provider reads its own env credentials; constructors don't fail when
// keys are missing — calls just return predictable errors.
func Register(svc *capitalpayment.PaymentService) {
	if svc == nil {
		return
	}
	svc.RegisterProvider(NewStripe())
	svc.RegisterProvider(NewPaystack())
	svc.RegisterProvider(NewFlutterwave())
}
