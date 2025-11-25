package accts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"scriptmang/drumstick/internal/testutils"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type create_acct_testcase struct {
	TestName                               string
	AttribNames                            []string
	Fname, Lname, Username, Address, Email string
	Password                               []byte
	ErrLst                                 []error
}

type testcase struct {
	path   string
	errLst []error
}

func printAcctErrs(attribNames []string, errMsgs []error) []error {
	var tempErrLst []error
	for i, errMsg := range errMsgs {
		tempErrLst = append(tempErrLst, fmt.Errorf("error:%s:%w", attribNames[i], errMsg))
	}
	return tempErrLst
}

func Test_encryptPassword(t *testing.T) {

	longPswd := bcrypt.ErrPasswordTooLong
	samplePswd1 := []byte("superfragilistic alidocious superfragilistic alidocious superfragilistic alidocious superfragilistic alidocious")
	samplePswd2 := []byte("TheRightAmount2")

	_, actualErr1 := encryptPassword(samplePswd1)
	assert.EqualError(t, actualErr1, longPswd.Error())

	_, actualErr2 := encryptPassword(samplePswd2)
	if actualErr2 != nil {
		t.Fatalf("error:pswd:%s", actualErr2.Error())
	}
}

func bind_TestCase(t *testing.T, dir string, file string) create_acct_testcase {
	t.Helper()
	relFilePath := filepath.Join("testdata", dir, file)
	jsonData, readErr := os.ReadFile(relFilePath)
	if readErr != nil {
		t.Fatal(readErr)
	}

	var testcase create_acct_testcase
	bindingErr := json.Unmarshal(jsonData, &testcase)
	if bindingErr != nil {
		t.Fatal(bindingErr)
	}
	return testcase
}

func Test_accountcreation(t *testing.T) {
	testCaseLst := []testcase{
		{"case01_fname_lname_empty.json", []error{ErrEmptyField, ErrEmptyField}},
		{"case02_fname_has_spaces.json", []error{ErrHasPunct}},
		{"case03_fname_has_nums.json", []error{ErrHasNums}},
		{"case04_fname_has_symbols.json", []error{ErrHasSymbols}},
		{"case05_lname_has_spaces.json", []error{ErrHasPunct}},
		{"case06_lname_has_nums.json", []error{ErrHasNums}},
		{"case07_lname_has_symbols.json", []error{ErrHasSymbols}},
		{"case08_addr_empty.json", []error{ErrEmptyField}},
		{"case09_addr_has_symbols.json", []error{ErrHasSymbols}},
		{"case10_email_empty.json", []error{ErrEmptyField}},
		{"case11_email_bad_addr.json", []error{ErrReqEndingAddr}},
		{"case12_email_miss_symbol.json", []error{ErrReqEmailSymbol}},
		{"case13_pswd_empty.json", []error{EmptyPswd}},
		{"case14_pswd_miss_capital.json", []error{PswdHasNoCapLetters}},
		{"case15_pswd_too_short.json", []error{PswdTooShort, PswdHasNoCapLetters, PswdHasNoDigits}},
		{"case16_pswd_too_long.json", []error{PswdTooLong}},
		{"case17_pswd_miss_digits.json", []error{PswdHasNoCapLetters, PswdHasNoDigits}},
		{"case18_pswd_has_symbols.json", []error{PswdHasSymbols}},
	}
	tests := []create_acct_testcase{}
	for _, tc := range testCaseLst {
		test := bind_TestCase(t, "invalidAccts", tc.path)
		test.ErrLst = tc.errLst
		tests = append(tests, test)
	}

	// range over table tests and validate the right errs are being thrown
	for _, tt := range tests {
		t.Run(tt.TestName, func(t *testing.T) {
			tempAcct := Account{Fname: tt.Fname, Lname: tt.Lname, Username: tt.Username, Address: tt.Address, Email: tt.Email, Password: tt.Password}
			actualErrLst := VetAllFields(tempAcct)
			expectedErrLst := printAcctErrs(tt.AttribNames, tt.ErrLst)
			assert.Equal(t, expectedErrLst, actualErrLst, expectedErrLst)
		})
	}
}

func Test_userlogin(t *testing.T) {
	pool := testutils.TestPool(t)
	defer pool.Close()

	ctx := context.Background()
	_, resetAndTestTxErr := pool.Exec(ctx, "TRUNCATE user_account RESTART IDENTITY CASCADE")
	if resetAndTestTxErr != nil {
		t.Fatal(resetAndTestTxErr)
	}
	// logic for the good pswd testcase
	relFilePath := filepath.Join("testdata", "dummy_acct.json")
	jsonData, readErr := os.ReadFile(relFilePath)
	if readErr != nil {
		t.Fatal(readErr)
	}

	var tempAcct Account
	bindingErr := json.Unmarshal(jsonData, &tempAcct)
	if bindingErr != nil {
		t.Fatal(bindingErr)
	}

	tx, txErr := pool.Begin(ctx)
	if txErr != nil {
		t.Fatal(txErr)
	}

	_, acctErr := CreateAcct(ctx, tx, tempAcct)
	if acctErr != nil {
		tx.Rollback(ctx)
		t.Fatal(acctErr)
	}

	txErr = tx.Commit(ctx)
	if txErr != nil {
		t.Fatal(txErr)
	}

	tests := []struct {
		testName    string
		username    string
		password    string
		expectedErr error
	}{
		{"Empty values", "", "", InvalidUserCreds},
		{"bad password", "dummy2", "dummy54Dwdgfdgf", InvalidUserCreds},
		{"bad username", "dummy2", "mldKMuffinDrop4", InvalidUserCreds},
		{"password is too long", "dummy2", "lkdrWkkkkkkkkk", InvalidUserCreds},
		{"password is too short", "dummy2", "dfdf", InvalidUserCreds},
		{"password is good", "dummy@gmail.com", "superSecretP432", nil},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			actualErr := CompareUserCreds(ctx, pool, tt.username, tt.password)
			if !errors.Is(actualErr, tt.expectedErr) {
				t.Fatalf("got: %v , expected: %v ", actualErr, tt.expectedErr)
			}
		})
	}

	_, resetAndTestTxErr = pool.Exec(ctx, "TRUNCATE user_account RESTART IDENTITY CASCADE")
	if resetAndTestTxErr != nil {
		t.Fatal(resetAndTestTxErr)
	}
}

// test userid retrieval by email
func Test_UserIDByEmail(t *testing.T) {

	pool := testutils.TestPool(t)
	defer pool.Close()

	testutils.ResetAndTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
		rscNotFound := errors.New("error: resource not found: id does not exist")
		relFilePath := filepath.Join("testdata", "dummy_acct.json")
		jsonData, readErr := os.ReadFile(relFilePath)
		if readErr != nil {
			t.Fatal(readErr)
		}

		var tempAcct Account
		bindingErr := json.Unmarshal(jsonData, &tempAcct)
		if bindingErr != nil {
			t.Fatal(bindingErr)
		}

		_, acctErr := CreateAcct(ctx, tx, tempAcct)
		if acctErr != nil {
			t.Fatal(acctErr)
		}
		actualID, actualErr := UserIDByEmail(ctx, tx, tempAcct.Email)

		if actualID == 1 {
			return
		}

		if errors.Is(actualErr, rscNotFound) {
			t.Fatal(rscNotFound)
		} else if actualErr != nil {
			t.Fatal(actualErr)
		}

	})
}
