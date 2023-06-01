package db

import (
	"context"
	"testing"
	"time"

	"server/utils"

	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	var maxid int64 = 1
	arg := CreateAccountParams{
		Id:        maxid,
		Login:     utils.RandomLogin(),
		Password:  utils.RandomPassword(),
		Address:   utils.RandomLogin(),
		CreatedAt: time.Now(),
		Session:   utils.RandomLogin(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Id, account.Id)
	require.Equal(t, arg.Login, account.Login)
	require.Equal(t, arg.Password, account.Password)
	require.Equal(t, arg.Address, account.Address)
	require.Equal(t, arg.Session, account.Session)

	require.NotZero(t, account.Id)
	require.NotZero(t, account.CreatedAt)
}

func TestGetAccount(t *testing.T) {
	var id int64 = 761
	account, err := testQueries.GetAccount(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, id, account.Id)
	require.Equal(t, "whdevmzm", account.Login)
	require.Equal(t, "gmgsjarhqc", account.Password)
}

func TestGetAccountByLogin(t *testing.T) {
	var login string = "KJDGHLKSJDGJLSDG"
	var id int64 = 643
	account, err := testQueries.GetAccountByLogin(context.Background(), login)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, id, account.Id)
	require.Equal(t, login, account.Login)
	require.Equal(t, "dodik123", account.Password)
}

func TestListAccounts(t *testing.T) {
	arg := ListAccountsParams{
		Limit:  int32(utils.RandomInt(1, 3)),
		Offset: int32(utils.RandomInt(1, 2)),
	}
	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)
}

func TestUpdateAccount(t *testing.T) {
	arg := UpdateAccountParams{
		Id:       643,
		Login:    "DURAK",
		Password: "dodik123",
		Address:  "www.bebra.ru",
		Session:  "SESSION",
	}
	account, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Id, account.Id)
	require.Equal(t, arg.Login, account.Login)
	require.Equal(t, arg.Password, account.Password)
	require.Equal(t, arg.Address, account.Address)
	require.Equal(t, arg.Session, account.Session)

}

func TestDeleteAccount(t *testing.T) {
	err := testQueries.DeleteAccount(context.Background(), 100001)
	require.NoError(t, err)
}

func TestGetLastAccount(t *testing.T) {
	account, err := testQueries.GetLastAccountID(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, account)
	var id int64 = 100000
	require.Equal(t, id, account)
}
