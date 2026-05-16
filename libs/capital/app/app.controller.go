package app

import (
	"os"
	"time"

	"github.com/awesome-goose/goose/types"
	capitalpayment "github.com/thescaffold/gox-apps/libs/capital/app/payment"
	capitalproviders "github.com/thescaffold/gox-apps/libs/capital/app/provider/providers"
	capitalplan "github.com/thescaffold/gox-apps/libs/capital/app/plan"
	capitalplantype "github.com/thescaffold/gox-apps/libs/capital/app/plantype"
	capitalrate "github.com/thescaffold/gox-apps/libs/capital/app/rate"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	capitalwallet "github.com/thescaffold/gox-apps/libs/capital/app/wallet"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/capital/src/app.controller.ts.
// Real provider HTTP integration is wired via PaymentService.RegisterProvider
// post-DI; without registered providers the endpoints return structurally-correct
// records but no external charges happen.
type AppController struct {
	appService     *AppService                     `inject:""`
	paymentService *capitalpayment.PaymentService  `inject:""`
	planEntity     *capitalplan.PlanEntity         `inject:""`
	planTypeEntity *capitalplantype.PlanTypeEntity `inject:""`
	rateEntity     *capitalrate.RateEntity         `inject:""`
	usageEntity    *capitalusage.UsageEntity       `inject:""`
	walletService  *capitalwallet.WalletService    `inject:""`
	lang           *i18n.Service                   `inject:""`
}

// OnRegister fires after goose DI resolves injections. Registers the
// Stripe/Paystack/Flutterwave providers with the injected PaymentService so
// real provider HTTP calls go out from `init`/`verify`/`charge`.
func (c *AppController) OnRegister() {
	capitalproviders.Register(c.paymentService)
}

// providers mirrors the TS `private providers = {...}` map keyed by ProviderName.
func (c *AppController) providers() map[string]map[string]any {
	return map[string]map[string]any{
		string(ProviderFlutterwave): {"name": string(ProviderFlutterwave), "key": os.Getenv("FLW_PUBLIC_KEY")},
		string(ProviderPaystack):    {"name": string(ProviderPaystack), "key": os.Getenv("PSK_PUBLIC_KEY")},
		string(ProviderStripe):      {"name": string(ProviderStripe), "key": os.Getenv("STR_PUBLIC_KEY")},
	}
}

func (c *AppController) title(ctx ntxctx.NTXContext) string {
	return c.lang.Translate("apps.capital.app.title", nil, ctx.Preference)
}

func (c *AppController) translate(key string, ctx ntxctx.NTXContext) string {
	return c.lang.Translate(key, nil, ctx.Preference)
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) — plain string,
// no envelope title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// Reference mirrors TS @Get(':type/reference/:reference').
func (c *AppController) Reference(dto *ReferenceDto) types.Output {
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	pay, _ := c.paymentService.Entity().First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "reference" = ?`,
		userID, clientID, workspaceID, dto.Reference,
	)
	if pay == nil {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.default-provider.error.no-payment", dto.Ctx))
	}
	return response.Success(pay, c.title(dto.Ctx),
		c.translate("apps.capital.app.default-provider.success", dto.Ctx), nil)
}

// GetProviders mirrors TS @Get('providers').
func (c *AppController) GetProviders(dto *ProvidersDto) types.Output {
	return response.Success(c.providers(), c.title(dto.Ctx),
		c.translate("apps.capital.app.providers.success", dto.Ctx), nil)
}

// Init mirrors TS @Post('init') — picks a provider, finds/creates the user
// plan, and starts a payment via PaymentService.Start.
func (c *AppController) Init(dto *InitDto) types.Output {
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	providerName := selectProvider(dto.Ctx.Preference)
	plan, _ := c.appService.FindOrCreateUserPlan(userID, clientID, workspaceID)
	if plan == nil {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.default-provider.error.no-country", dto.Ctx))
	}
	meta := map[string]any{"currency": currencyOf(dto.Ctx.Preference)}
	startOut, err := c.paymentService.Start(plan, dto.Amount, providerName, dto.FirstName, dto.LastName, userID, meta)
	if err != nil {
		return response.InternalServerError(c.title(dto.Ctx), err.Error())
	}
	startOut["amount"] = dto.Amount
	startOut["currency"] = currencyOf(dto.Ctx.Preference)
	for k, v := range c.providers()[providerName] {
		startOut[k] = v
	}
	return response.Success(startOut, c.title(dto.Ctx),
		c.translate("apps.capital.app.default-provider.success", dto.Ctx), nil)
}

// GetDefaultProvider mirrors TS @Post('default-provider') — deprecated mirror
// of Init that takes the same DefaultProviderDto.
func (c *AppController) GetDefaultProvider(dto *DefaultProviderDto) types.Output {
	return c.Init(&InitDto{Ctx: dto.Ctx, Amount: dto.Amount, FirstName: dto.FirstName, LastName: dto.LastName, Meta: dto.Meta})
}

// Verify mirrors TS @Post('verify/:type').
func (c *AppController) Verify(dto *VerifyPaymentDto) types.Output {
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	plan, _ := c.appService.FindOrCreateUserPlan(userID, clientID, workspaceID)
	if plan == nil {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.default-provider.error.payment-not-verified", dto.Ctx))
	}
	ok, err := c.paymentService.Complete(plan, dto.Type, dto.Reference, contextMap(dto.Ctx))
	if err != nil || !ok {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.default-provider.error.payment-not-verified", dto.Ctx))
	}
	return response.Success(map[string]any{
		"reference": dto.Reference,
		"amount":    dto.Amount,
		"currency":  dto.Currency,
	}, c.title(dto.Ctx),
		c.translate("apps.capital.app.default-provider.verify.success", dto.Ctx), nil)
}

// Charge mirrors TS @Post(':type/charge/:providerId').
func (c *AppController) Charge(dto *ChargePaymentDto) types.Output {
	if dto.Type != "wallet" && dto.Type != "provider" {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.post.charge.error.invalid-provider", dto.Ctx))
	}
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	plan, _ := c.appService.FindOrCreateUserPlan(userID, clientID, workspaceID)
	if plan == nil {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.post.charge.error.payment-not-verified", dto.Ctx))
	}
	meta := map[string]any{
		"type":    "apps.capital.payment.link",
		"context": contextMap(dto.Ctx),
	}
	paid, ref, err := c.paymentService.New(
		[]string{paymentTypeFromString(dto.Type)},
		plan, dto.Amount, meta, "", dto.ProviderId, contextMap(dto.Ctx),
	)
	if err != nil || !paid {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.post.charge.error.payment-not-verified", dto.Ctx))
	}
	return response.Success(map[string]any{
		"reference": ref, "amount": dto.Amount, "currency": dto.Currency,
	}, c.title(dto.Ctx),
		c.translate("apps.capital.app.post.charge.success", dto.Ctx), nil)
}

// Pay mirrors TS @Post('pay') — runs an existing-provider charge.
func (c *AppController) Pay(dto *PayDto) types.Output {
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	plan, _ := c.appService.FindOrCreateUserPlan(userID, clientID, workspaceID)
	providerId := ""
	if dto.ProviderId != nil {
		providerId = *dto.ProviderId
	}
	if plan != nil {
		_, _ = c.paymentService.Existing([]string{string(PaymentTypeCharge)}, plan, "", providerId, contextMap(dto.Ctx))
	}
	return response.Success(nil, c.title(dto.Ctx),
		c.translate("apps.capital.app.post.pay.success", dto.Ctx), nil)
}

// Debt mirrors TS @Get('debt').
func (c *AppController) Debt(dto *StatusDto) types.Output {
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	plan, _ := c.appService.FindOrCreateUserPlan(userID, clientID, workspaceID)
	debt, _ := c.paymentService.Debt(plan)
	return response.Success(debt, c.title(dto.Ctx),
		c.translate("apps.capital.app.get.debt.success", dto.Ctx), nil)
}

// Accrual mirrors TS @Get('accrual').
func (c *AppController) Accrual(dto *StatusDto) types.Output {
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	plan, _ := c.appService.FindOrCreateUserPlan(userID, clientID, workspaceID)
	if plan == nil || plan.PeriodType == nil {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.get.accrual.error.invalid-period-type", dto.Ctx))
	}
	out, _ := c.paymentService.Accrual(plan)
	return response.Success(out, c.title(dto.Ctx),
		c.translate("apps.capital.app.get.accrual.success", dto.Ctx), nil)
}

// Status mirrors TS @Get('status') — emits the debt + days/warning/danger flags.
func (c *AppController) Status(dto *StatusDto) types.Output {
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	plan, _ := c.appService.FindOrCreateUserPlan(userID, clientID, workspaceID)
	debt, _ := c.paymentService.Debt(plan)
	total := 0.0
	var from, to time.Time
	if debt != nil {
		total = debt.Amount
		from = debt.From
		to = debt.To
	}
	days := 0
	if !from.IsZero() && !to.IsZero() {
		days = int(to.Sub(from).Hours() / 24)
	}
	return response.Success(map[string]any{
		"total":   total,
		"from":    from,
		"to":      to,
		"days":    days,
		"warning": total > 0,
		"danger":  total > 0 && days > 7,
	}, c.title(dto.Ctx),
		c.translate("apps.capital.app.get.status.success", dto.Ctx), nil)
}

// UserPlan mirrors TS @Get('user-plan').
func (c *AppController) UserPlan(dto *UserPlanDto) types.Output {
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	plan, _ := c.appService.FindOrCreateUserPlan(userID, clientID, workspaceID)
	return response.Success(plan, c.title(dto.Ctx),
		c.translate("apps.capital.app.get.user-plan.success", dto.Ctx), nil)
}

// ComputeUpgrade mirrors TS @Post('user-plan/compute-upgrade').
func (c *AppController) ComputeUpgrade(dto *UpgradeUserPlanDto) types.Output {
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	plan, _ := c.appService.FindOrCreateUserPlan(userID, clientID, workspaceID)
	existingType, _ := c.appService.PlanType(plan)
	if dto.PlanTypeId == nil || *dto.PlanTypeId == "" {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.post.user-plan-upgrade.error.not-found", dto.Ctx))
	}
	if existingType != nil && existingType.Id == *dto.PlanTypeId {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.post.user-plan-upgrade.error.same-plan", dto.Ctx))
	}
	newType, _ := c.planTypeEntity.First(`"id" = ?`, *dto.PlanTypeId)
	if newType == nil {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.post.user-plan-upgrade.error.not-found", dto.Ctx))
	}
	if plan != nil && plan.PeriodType != nil && *plan.PeriodType == "yearly" && dto.PeriodType == "monthly" {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.post.user-plan-upgrade.error.cannot-downgrade", dto.Ctx))
	}
	if existingType != nil && periodAmount(existingType, dto.PeriodType) > periodAmount(newType, dto.PeriodType) {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.post.user-plan-upgrade.error.cannot-downgrade", dto.Ctx))
	}
	deduction, newAmount, diff, ok := c.appService.ComputePlanUpgrade(dto.PeriodType, existingType, newType)
	if !ok {
		return response.BadRequest(c.title(dto.Ctx),
			c.translate("apps.capital.app.post.user-plan-upgrade.error.invalid-request", dto.Ctx))
	}
	return response.Success(map[string]any{
		"deduction": deduction, "newAmount": newAmount, "diff": diff,
	}, c.title(dto.Ctx),
		c.translate("apps.capital.app.post.user-plan-upgrade.success", dto.Ctx), nil)
}

// PlanChange mirrors TS @Post('plan/change').
func (c *AppController) PlanChange(dto *CapitalPlanChangeDto) types.Output {
	clientID := dto.Ctx.ClientID
	out, err := c.appService.ChangeUserPlan(dto.UserId, clientID, dto.WorkspaceId, dto.Currency, dto.PeriodType, dto.PlanType)
	if err != nil || out == nil {
		return response.BadRequest(
			c.translate("apps.controller.interface.capital.plan.change.title", dto.Ctx),
			c.translate("apps.controller.interface.capital.plan.change.error", dto.Ctx),
		)
	}
	return response.Success(out,
		c.translate("apps.controller.interface.capital.plan.change.title", dto.Ctx),
		c.translate("apps.controller.interface.capital.plan.change.success", dto.Ctx), nil)
}

// PlanSubscribe mirrors TS @Post('plan/subscribe').
func (c *AppController) PlanSubscribe(dto *CapitalPlanSubscribeDto) types.Output {
	clientID := dto.Ctx.ClientID
	out, err := c.appService.Subscribe(dto.UserId, clientID, dto.WorkspaceId, dto.Currency, dto.Amount, c.paymentService)
	if err != nil || out == nil {
		return response.BadRequest(
			c.translate("apps.controller.interface.capital.plan.subscription.title", dto.Ctx),
			c.translate("apps.controller.interface.capital.plan.subscription.error", dto.Ctx),
		)
	}
	return response.Success(out,
		c.translate("apps.controller.interface.capital.plan.subscription.title", dto.Ctx),
		c.translate("apps.controller.interface.capital.plan.subscription.success", dto.Ctx), nil)
}

// UsageStart mirrors TS @Post('usage/start').
func (c *AppController) UsageStart(dto *UsageStartDto) types.Output {
	rate, _ := c.rateEntity.First(`"type" = ?`, dto.EntityName)
	now := time.Now().UTC()
	q := 1
	if dto.Quantity != nil {
		if n := parseIntDefault(*dto.Quantity, 1); n > 0 {
			q = n
		}
	}
	rateId := ""
	if rate != nil {
		rateId = rate.Id
	}
	row := &capitalusage.Usage{
		UserId: dto.UserId, ClientId: dto.ClientId, WorkspaceId: dto.WorkspaceId,
		EntityName: dto.EntityName, EntityId: dto.EntityId,
		RateId: rateId, Quantity: q, StartAt: &now, Meta: dto.Meta,
	}
	_ = c.usageEntity.Insert(row)
	return response.Success(row, c.title(dto.Ctx),
		c.translate("apps.capital.app.usage.start.success", dto.Ctx), nil)
}

// UsageUpdate mirrors TS @Post('usage/update').
func (c *AppController) UsageUpdate(dto *UsageStartDto) types.Output {
	row, _ := c.usageEntity.First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "entity_name" = ? AND "entity_id" = ?`,
		dto.UserId, dto.ClientId, dto.WorkspaceId, dto.EntityName, dto.EntityId,
	)
	if row != nil {
		if dto.Quantity != nil {
			row.Quantity = parseIntDefault(*dto.Quantity, row.Quantity)
		}
		if len(dto.Meta) > 0 {
			row.Meta = dto.Meta
		}
		_, _ = c.usageEntity.Update(row, `"id" = ?`, row.Id)
	}
	return response.Success(nil, c.title(dto.Ctx),
		c.translate("apps.capital.app.usage.start.success", dto.Ctx), nil)
}

// UsageStop mirrors TS @Post('usage/stop').
func (c *AppController) UsageStop(dto *UsageStopDto) types.Output {
	row, _ := c.usageEntity.First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "entity_name" = ? AND "entity_id" = ?`,
		dto.UserId, dto.ClientId, dto.WorkspaceId, dto.EntityName, dto.EntityId,
	)
	if row != nil {
		now := time.Now().UTC()
		row.StopAt = &now
		_, _ = c.usageEntity.Update(row, `"id" = ?`, row.Id)
	}
	return response.Success(row, c.title(dto.Ctx),
		c.translate("apps.capital.app.usage.stop.success", dto.Ctx), nil)
}

// contextMap builds the TS-equivalent `context` envelope passed into
// PaymentService methods. PaymentService uses it to gate invoice/receipt
// generation (TS branches on `if context != nil`).
func contextMap(ntx ntxctx.NTXContext) map[string]any {
	out := map[string]any{
		"user":       ntx.User,
		"client":     ntx.Client,
		"workspace":  ntx.Workspace,
		"preference": ntx.Preference,
	}
	return out
}

// identityFromContext extracts the (userId, clientId, workspaceId) tuple.
func identityFromContext(ntx ntxctx.NTXContext) (string, string, string) {
	userID, clientID, workspaceID := "", "", ""
	if ntx.User != nil {
		userID, _ = ntx.User["id"].(string)
	}
	if userID == "" {
		userID = ntx.UserID
	}
	if ntx.Client != nil {
		clientID, _ = ntx.Client["id"].(string)
	}
	if clientID == "" {
		clientID = ntx.ClientID
	}
	if ntx.Workspace != nil {
		workspaceID, _ = ntx.Workspace["id"].(string)
	}
	if workspaceID == "" {
		workspaceID = ntx.WorkspaceID
	}
	return userID, clientID, workspaceID
}

func selectProvider(pref any) string {
	if pref == nil {
		return string(ProviderStripe)
	}
	if c := currencyOf(pref); c == "NGN" {
		return string(ProviderPaystack)
	}
	return string(ProviderStripe)
}

func currencyOf(pref any) string {
	if pref == nil {
		return ""
	}
	if m, ok := pref.(map[string]any); ok {
		if v, ok := m["currency"].(string); ok {
			return v
		}
	}
	return ""
}

func paymentTypeFromString(t string) string {
	switch t {
	case "wallet":
		return string(PaymentTypeWallet)
	case "provider":
		return string(PaymentTypeCharge)
	}
	return string(PaymentTypeCharge)
}

func parseIntDefault(s string, def int) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if n == 0 {
		return def
	}
	return n
}
