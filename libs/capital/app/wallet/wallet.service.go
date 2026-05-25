package wallet

import (
	"encoding/json"
	"errors"
	"time"

	gooErrors "github.com/awesome-goose/goose/errors"
	capitalaccount "github.com/thescaffold/gox-apps/libs/capital/app/account"
	capitaltransaction "github.com/thescaffold/gox-apps/libs/capital/app/transaction"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// WalletTransaction type/status values mirror TS app.dto.ts
// WalletTransactionType / WalletTransactionStatusType.
const (
	txTypeCR         = "cr"
	txTypeDR         = "dr"
	txStatusComplete = "complete"
	txStatusPending  = "pending"
	txStatusReversed = "reversed"
)

var (
	errAccountNotFound  = errors.New("account not found")
	errAccountClosed    = errors.New("account closed")
	errAccountSuspended = errors.New("account suspended")
	errTxNotFound       = errors.New("transaction not found")
	errInvalidAmount    = errors.New("invalid amount")
	errInsufficient     = errors.New("insufficient funds")
)

type WalletService struct {
	accountEntity     *capitalaccount.AccountEntity         `inject:""`
	transactionEntity *capitaltransaction.TransactionEntity `inject:""`
}

// checkAccount mirrors TS WalletService.checkAccount: rejects a missing,
// closed (unless ignoreClosed) or suspended (unless ignoreSuspended) account.
func (s *WalletService) checkAccount(acc *capitalaccount.Account, ignoreClosed, ignoreSuspended bool) error {
	if acc == nil {
		return errAccountNotFound
	}
	if acc.ClosedAt != nil && !ignoreClosed {
		return errAccountClosed
	}
	if acc.SuspendedAt != nil && !ignoreSuspended {
		return errAccountSuspended
	}
	return nil
}

// newTx builds a Transaction row, attaching status, narration (meta["narration"])
// and the meta blob as desc — mirroring TS which persists status + narration + desc.
func (s *WalletService) newTx(acc *capitalaccount.Account, ref string, amt, book, avail int, currency, typ, status string, meta map[string]any) *capitaltransaction.Transaction {
	tx := &capitaltransaction.Transaction{
		UserId:           acc.UserId,
		AccountId:        acc.Id,
		Amount:           amt,
		BookBalance:      book,
		AvailableBalance: avail,
		Currency:         currency,
		Reference:        ref,
		Type:             typ,
		Status:           &status,
		TransactionAt:    time.Now().UTC(),
	}
	if meta != nil {
		if n, ok := meta["narration"].(string); ok && n != "" {
			tx.Narration = &n
		}
		if b, err := json.Marshal(meta); err == nil {
			d := string(b)
			tx.Desc = &d
		}
	}
	return tx
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

// ── FUNDING (credit) ──────────────────────────────────────────────────────────
// Mirrors TS WalletService.fund.{default,lien,execute,reverse}. NOTE: TS wraps
// each in a pessimistic-write DB transaction; goose's Entity API exposes no row
// locks, so (per the idiomatic-Go directive — behavioural, not lock-for-lock,
// parity) these run as sequential ops, matching the pre-existing gox pattern.

// FundDefault credits an account directly (book + available both increase).
func (s *WalletService) FundDefault(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	acc, _ := s.accountEntity.First(`"user_id" = ?`, userId)
	if err := s.checkAccount(acc, false, false); err != nil {
		return nil, err
	}
	amt := int(amount)
	book := acc.BookBalance + amt
	avail := acc.AvailableBalance + amt
	tx := s.newTx(acc, reference, amt, book, avail, currency, txTypeCR, txStatusComplete, meta)
	if err := s.transactionEntity.Insert(tx); err != nil {
		return nil, err
	}
	acc.BookBalance, acc.AvailableBalance = book, avail
	if _, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id); err != nil {
		return nil, err
	}
	return tx, nil
}

// Fund is the public name kept for callers; aliases FundDefault.
func (s *WalletService) Fund(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	return s.FundDefault(userId, reference, amount, currency, meta)
}

// FundLien puts a lien on funds being credited: book increases, available does
// NOT; the transaction is Pending until executed or reversed.
func (s *WalletService) FundLien(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	acc, _ := s.accountEntity.First(`"user_id" = ?`, userId)
	if err := s.checkAccount(acc, false, false); err != nil {
		return nil, err
	}
	amt := int(amount)
	book := acc.BookBalance + amt
	tx := s.newTx(acc, reference, amt, book, acc.AvailableBalance, currency, txTypeCR, txStatusPending, meta)
	if err := s.transactionEntity.Insert(tx); err != nil {
		return nil, err
	}
	acc.BookBalance = book
	if _, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id); err != nil {
		return nil, err
	}
	return tx, nil
}

// FundExecute executes a pending credit lien, making the funds available.
func (s *WalletService) FundExecute(userId, reference string, amount float64, currency string, meta map[string]any) error {
	pending, _ := s.transactionEntity.First(`"reference" = ? AND "type" = ? AND "status" = ?`, reference, txTypeCR, txStatusPending)
	if pending == nil {
		return errTxNotFound
	}
	amt := int(amount)
	if amt > pending.Amount {
		return errInvalidAmount
	}
	acc, _ := s.accountEntity.First(`"id" = ?`, pending.AccountId)
	if err := s.checkAccount(acc, false, false); err != nil {
		return err
	}
	avail := acc.AvailableBalance + amt
	exe := s.newTx(acc, "EXE-"+reference, amt, acc.BookBalance, avail, currency, txTypeCR, txStatusComplete, meta)
	if err := s.transactionEntity.Insert(exe); err != nil {
		return err
	}
	s.settlePending(pending, amt, txStatusComplete)
	acc.AvailableBalance = avail
	_, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id)
	return err
}

// FundReverse reverses a pending credit lien, removing the funds from book.
func (s *WalletService) FundReverse(userId, reference string, amount float64, currency string, meta map[string]any) error {
	pending, _ := s.transactionEntity.First(`"reference" = ? AND "type" = ? AND "status" = ?`, reference, txTypeCR, txStatusPending)
	if pending == nil {
		return errTxNotFound
	}
	amt := int(amount)
	if amt > pending.Amount {
		return errInvalidAmount
	}
	acc, _ := s.accountEntity.First(`"id" = ?`, pending.AccountId)
	if err := s.checkAccount(acc, false, false); err != nil {
		return err
	}
	book := acc.BookBalance - amt
	rvl := s.newTx(acc, "RVL-"+reference, amt, book, acc.AvailableBalance, currency, txTypeDR, txStatusComplete, meta)
	if err := s.transactionEntity.Insert(rvl); err != nil {
		return err
	}
	s.settlePending(pending, amt, txStatusReversed)
	acc.BookBalance = book
	_, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id)
	return err
}

// ── WITHDRAWAL (debit) ─────────────────────────────────────────────────────────

// WithdrawDefault debits an account directly (book + available both decrease).
func (s *WalletService) WithdrawDefault(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	acc, _ := s.accountEntity.First(`"user_id" = ?`, userId)
	if err := s.checkAccount(acc, false, false); err != nil {
		return nil, err
	}
	amt := int(amount)
	if amt > acc.AvailableBalance {
		return nil, errInsufficient
	}
	book := acc.BookBalance - amt
	avail := acc.AvailableBalance - amt
	tx := s.newTx(acc, reference, amt, book, avail, currency, txTypeDR, txStatusComplete, meta)
	if err := s.transactionEntity.Insert(tx); err != nil {
		return nil, err
	}
	acc.BookBalance, acc.AvailableBalance = book, avail
	if _, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id); err != nil {
		return nil, err
	}
	return tx, nil
}

// Withdraw is the public name kept for callers; aliases WithdrawDefault.
func (s *WalletService) Withdraw(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	return s.WithdrawDefault(userId, reference, amount, currency, meta)
}

// WithdrawLien puts a lien on funds to be withdrawn: available decreases, book
// does NOT; the transaction is Pending.
func (s *WalletService) WithdrawLien(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	acc, _ := s.accountEntity.First(`"user_id" = ?`, userId)
	if err := s.checkAccount(acc, false, false); err != nil {
		return nil, err
	}
	amt := int(amount)
	if amt > acc.AvailableBalance {
		return nil, errInsufficient
	}
	avail := acc.AvailableBalance - amt
	tx := s.newTx(acc, reference, amt, acc.BookBalance, avail, currency, txTypeDR, txStatusPending, meta)
	if err := s.transactionEntity.Insert(tx); err != nil {
		return nil, err
	}
	acc.AvailableBalance = avail
	if _, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id); err != nil {
		return nil, err
	}
	return tx, nil
}

// WithdrawExecute executes a pending withdrawal lien, debiting the book balance.
func (s *WalletService) WithdrawExecute(userId, reference string, amount float64, currency string, meta map[string]any) error {
	pending, _ := s.transactionEntity.First(`"reference" = ? AND "type" = ? AND "status" = ?`, reference, txTypeDR, txStatusPending)
	if pending == nil {
		return errTxNotFound
	}
	amt := int(amount)
	if amt > pending.Amount {
		return errInvalidAmount
	}
	acc, _ := s.accountEntity.First(`"id" = ?`, pending.AccountId)
	if err := s.checkAccount(acc, false, false); err != nil {
		return err
	}
	book := acc.BookBalance - amt
	exe := s.newTx(acc, "EXE-"+reference, amt, book, acc.AvailableBalance, currency, txTypeDR, txStatusComplete, meta)
	if err := s.transactionEntity.Insert(exe); err != nil {
		return err
	}
	s.settlePending(pending, amt, txStatusComplete)
	acc.BookBalance = book
	_, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id)
	return err
}

// WithdrawReverse reverses a pending withdrawal lien, making funds available again.
func (s *WalletService) WithdrawReverse(userId, reference string, amount float64, currency string, meta map[string]any) error {
	pending, _ := s.transactionEntity.First(`"reference" = ? AND "type" = ? AND "status" = ?`, reference, txTypeDR, txStatusPending)
	if pending == nil {
		return errTxNotFound
	}
	amt := int(amount)
	if amt > pending.Amount {
		return errInvalidAmount
	}
	acc, _ := s.accountEntity.First(`"id" = ?`, pending.AccountId)
	if err := s.checkAccount(acc, false, false); err != nil {
		return err
	}
	avail := acc.AvailableBalance + amt
	rvl := s.newTx(acc, "RVL-"+reference, amt, acc.BookBalance, avail, currency, txTypeCR, txStatusComplete, meta)
	if err := s.transactionEntity.Insert(rvl); err != nil {
		return err
	}
	s.settlePending(pending, amt, txStatusReversed)
	acc.AvailableBalance = avail
	_, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id)
	return err
}

// DoubleDefault credits then debits the account by the same amount (net-zero
// balance) — a paired CR (<ref>-CR) and DR (<ref>-DR), both Complete. Mirrors
// TS WalletService.double.default, used by the payment verify/charge/record legs.
func (s *WalletService) DoubleDefault(userId, reference string, amount float64, currency string, meta map[string]any) (*capitaltransaction.Transaction, error) {
	acc, _ := s.accountEntity.First(`"user_id" = ?`, userId)
	if err := s.checkAccount(acc, false, false); err != nil {
		return nil, err
	}
	amt := int(amount)
	// Credit leg.
	creditBook := acc.BookBalance + amt
	creditAvail := acc.AvailableBalance + amt
	cr := s.newTx(acc, reference+"-CR", amt, creditBook, creditAvail, currency, txTypeCR, txStatusComplete, meta)
	if err := s.transactionEntity.Insert(cr); err != nil {
		return nil, err
	}
	acc.BookBalance, acc.AvailableBalance = creditBook, creditAvail
	if _, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id); err != nil {
		return nil, err
	}
	// Debit leg (back to the original balance).
	debitBook := creditBook - amt
	debitAvail := creditAvail - amt
	dr := s.newTx(acc, reference+"-DR", amt, debitBook, debitAvail, currency, txTypeDR, txStatusComplete, meta)
	if err := s.transactionEntity.Insert(dr); err != nil {
		return nil, err
	}
	acc.BookBalance, acc.AvailableBalance = debitBook, debitAvail
	if _, err := s.accountEntity.Update(acc, `"id" = ?`, acc.Id); err != nil {
		return nil, err
	}
	return dr, nil
}

// settlePending applies the TS "update original pending transaction" step:
// reduce its amount by the settled amount; if fully settled set finalStatus,
// else keep it Pending.
func (s *WalletService) settlePending(pending *capitaltransaction.Transaction, amt int, finalStatus string) {
	remaining := pending.Amount - amt
	status := txStatusPending
	if remaining == 0 {
		status = finalStatus
	}
	pending.Amount = remaining
	pending.Status = &status
	_, _ = s.transactionEntity.Update(pending, `"id" = ?`, pending.Id)
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
