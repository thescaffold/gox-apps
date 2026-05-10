package wallet

import (
	"errors"

	gooErrors "github.com/awesome-goose/goose/errors"
	capitalaccount "github.com/thescaffold/gox-apps/libs/capital/app/account"
	capitaltransaction "github.com/thescaffold/gox-apps/libs/capital/app/transaction"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

type WalletService struct {
	accountEntity     *capitalaccount.AccountEntity         `inject:""`
	transactionEntity *capitaltransaction.TransactionEntity `inject:""`
}

func (s *WalletService) Init(userId, clientId, workspaceId, currency, typ, label string) (*capitalaccount.Account, error) {
	existing, err := s.accountEntity.First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ?`,
		userId, clientId, workspaceId,
	)
	if err != nil && !errors.Is(err, gooErrors.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	ref := utils.Reference("WLT", 36)
	a := &capitalaccount.Account{
		UserId:           userId,
		ClientId:         clientId,
		WorkspaceId:      workspaceId,
		BookBalance:      0,
		AvailableBalance: 0,
		Currency:         currency,
		Reference:        ref,
		Number:           &ref,
		Type:             &typ,
		Label:            &label,
	}
	if err := s.accountEntity.Insert(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *WalletService) Detail(userId string) (*capitalaccount.Account, error) {
	return s.accountEntity.First(`"user_id" = ?`, userId)
}

func (s *WalletService) Balance(userId string) (*capitalaccount.Account, error) {
	return s.accountEntity.First(`"user_id" = ?`, userId)
}

func (s *WalletService) Upgrade(userId string, dailyLimit, monthlyLimit int) (*capitalaccount.Account, error) {
	acc, err := s.accountEntity.First(`"user_id" = ?`, userId)
	if err != nil {
		return nil, err
	}
	acc.DailyLimit = &dailyLimit
	acc.MonthlyLimit = &monthlyLimit
	if _, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *WalletService) Transactions(userId string, page, perPage int) ([]capitaltransaction.Transaction, int64, error) {
	total, err := s.transactionEntity.Count(`"user_id" = ?`, userId)
	if err != nil {
		return nil, 0, err
	}
	txs, err := s.transactionEntity.Page(page, perPage, `"user_id" = ?`, userId)
	if err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}
