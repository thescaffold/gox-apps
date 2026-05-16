package wallet

import (
	"errors"
	"time"

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

// Fund mirrors TS WalletService.fund — credits a user's wallet by inserting
// a credit Transaction row and bumping the available balance. Used by the
// `apps.capital.card.link` / `payment.link` / `gate.link` payment.pay
// subscription branches.
func (s *WalletService) Fund(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	acc, err := s.accountEntity.First(`"user_id" = ?`, userId)
	if err != nil || acc == nil {
		return nil, err
	}
	amt := int(amount)
	tx := &capitaltransaction.Transaction{
		UserId:           acc.UserId,
		AccountId:        acc.Id,
		Reference:        reference,
		Amount:           amt,
		BookBalance:      acc.BookBalance + amt,
		AvailableBalance: acc.AvailableBalance + amt,
		Currency:         currency,
		Type:             "cr",
		TransactionAt:    time.Now().UTC(),
	}
	if err := s.transactionEntity.Insert(tx); err != nil {
		return nil, err
	}
	acc.AvailableBalance += amt
	acc.BookBalance += amt
	if _, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id); err != nil {
		return nil, err
	}
	return tx, nil
}

// FundDefault matches the TS `walletService.fund.default(...)` callsite.
// Delegates to Fund — name preserved so subscription handlers read as TS.
func (s *WalletService) FundDefault(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	return s.Fund(userId, reference, amount, currency, meta)
}

// Withdraw mirrors TS WalletService.withdraw — debits the user's wallet by
// inserting a debit Transaction and reducing the available balance. Returns
// an error when the wallet doesn't exist or the balance can't cover the
// withdrawal (TS-equivalent insufficient-funds guard).
func (s *WalletService) Withdraw(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	acc, err := s.accountEntity.First(`"user_id" = ?`, userId)
	if err != nil || acc == nil {
		return nil, errors.New("wallet not found")
	}
	amt := int(amount)
	if acc.AvailableBalance < amt {
		return nil, errors.New("insufficient funds")
	}
	tx := &capitaltransaction.Transaction{
		UserId:           acc.UserId,
		AccountId:        acc.Id,
		Reference:        reference,
		Amount:           amt,
		BookBalance:      acc.BookBalance - amt,
		AvailableBalance: acc.AvailableBalance - amt,
		Currency:         currency,
		Type:             "dr",
		TransactionAt:    time.Now().UTC(),
	}
	if err := s.transactionEntity.Insert(tx); err != nil {
		return nil, err
	}
	acc.AvailableBalance -= amt
	acc.BookBalance -= amt
	if _, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id); err != nil {
		return nil, err
	}
	return tx, nil
}

// WithdrawDefault matches the TS `walletService.withdraw.default(...)` callsite.
func (s *WalletService) WithdrawDefault(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	return s.Withdraw(userId, reference, amount, currency, meta)
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
