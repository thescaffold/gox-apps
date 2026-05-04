package wallet

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages-core/response"
)

type WalletController struct {
	walletService *WalletService `inject:""`
}

func (c *WalletController) Init(dto *WalletInitDto) types.Output {
	if dto.NTX.UserID == "" {
		return response.BadRequest("wallet", "user not found")
	}

	currency := ""
	if dto.NTX.Preference != nil {
		if v, ok := dto.NTX.Preference["currency"].(string); ok {
			currency = v
		}
	}

	typ := dto.Type
	if typ == "" {
		typ = "default"
	}
	label := dto.Label
	if label == "" {
		label = "default"
	}

	wallet, err := c.walletService.Init(
		dto.NTX.UserID, dto.NTX.ClientID, dto.NTX.WorkspaceID,
		currency, typ, label,
	)
	if err != nil {
		return response.InternalServerError("wallet", err.Error())
	}
	return response.Success(wallet, "wallet", "wallet initialized", nil)
}

func (c *WalletController) Detail(dto *WalletDetailDto) types.Output {
	if dto.NTX.UserID == "" {
		return response.BadRequest("wallet", "user not found")
	}
	account, err := c.walletService.Detail(dto.NTX.UserID)
	if err != nil {
		return response.NotFound("wallet", "account not found")
	}
	return response.Success(account, "wallet", "ok", nil)
}

func (c *WalletController) Balance(dto *WalletBalanceDto) types.Output {
	if dto.NTX.UserID == "" {
		return response.BadRequest("wallet", "user not found")
	}
	account, err := c.walletService.Balance(dto.NTX.UserID)
	if err != nil {
		return response.NotFound("wallet", "account not found")
	}
	return response.Success(account, "wallet", "ok", nil)
}

func (c *WalletController) Upgrade(dto *WalletUpgradeDto) types.Output {
	if dto.NTX.UserID == "" {
		return response.BadRequest("wallet", "user not found")
	}
	account, err := c.walletService.Upgrade(dto.NTX.UserID, dto.DailyLimit, dto.MonthlyLimit)
	if err != nil {
		return response.InternalServerError("wallet", err.Error())
	}
	return response.Success(account, "wallet", "limits updated", nil)
}

func (c *WalletController) Transactions(dto *WalletTransactionsDto) types.Output {
	page := dto.Page
	if page < 1 {
		page = 1
	}
	perPage := dto.PerPage
	if perPage < 1 {
		perPage = 20
	}

	txs, total, err := c.walletService.Transactions(dto.NTX.UserID, page, perPage)
	if err != nil {
		return response.InternalServerError("wallet", err.Error())
	}
	return response.Paginated(txs, page, perPage, total, "wallet", "ok")
}
