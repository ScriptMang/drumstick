package accts

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"scriptmang/drumstick/internal/templateRenderer"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func dict(vals ...any) (map[string]any, error) {
	if len(vals)%2 != 0 {
		return nil, fmt.Errorf("dictionary requires an even number of arguments")
	}

	m := make(map[string]any, len(vals)/2)
	for i := 0; i < len(vals); i += 2 {
		ky, ok := vals[i].(string)
		if !ok {
			return nil, fmt.Errorf("dictionary keys must be string type")
		}
		m[ky] = vals[i+1]
	}
	return m, nil
}

func setupEchoClient() *echo.Echo {
	tm := &templateRenderer.TemplateManager{
		Templates: template.Must(template.New("").Funcs(template.FuncMap{
			"dict": dict,
		}).ParseGlob("../../ui/html/pages/*[^#?!|].tmpl")),
	}

	r := echo.New()

	r.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.Renderer = tm

	return r
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

func Test_accountcreation(t *testing.T) {
	r := setupEchoClient()

	// field errors
	errEmptyField := errors.New("field is empty")
	errHasNums := errors.New("field can't contain any numbers")
	errHasSymbols := errors.New("field can't contain any symbols")
	errHasPunct := errors.New("field has punctuation")

	// email errs
	errReqSymbol := errors.New("email is missing an '@' symbol.")
	errReqEndingAddr := errors.New("email doesn't match any of the ending addresses.")

	// password errs
	emptyPswd := errors.New("field can't be empty")
	shortPswd := errors.New("field too short")
	longPswd := errors.New("field too long")
	pswdHasNoDigits := errors.New("field missing at least 1 digit")
	pswdHasSymbols := errors.New("field can't contain any symbols")
	missingCapitalLetter := errors.New("field is missing at least 1 capital letter")

	tests := []struct {
		testName                               string
		attribNames                            []string
		fname, lname, username, address, email string
		password                               []byte
		errLst                                 []error
	}{
		{"Empty Attributes", []string{"fname", "lname"}, "", "", "testUser1", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errEmptyField, errEmptyField}},
		{"No Spaces Found in Fname", []string{"fname"}, " Jon", "Martin", "testUser2", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasPunct}},
		{"No Spaces Found in Lname", []string{"lname"}, "Jon", " Martin", "testUser3", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasPunct}},
		{"No Spaces Found in Fname and Lname", []string{"fname", "lname"}, " Jon", " Martin", "testUser4", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasPunct, errHasPunct}},
		{"No Symbols in Fname", []string{"fname"}, "Jon@", "Martin", "testUser5", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasSymbols}},
		{"No Symbols in Lname", []string{"lname"}, "Jon", "@Martin", "testUser6", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasSymbols}},
		{"No Numbers in Fname", []string{"fname"}, "Jon3", "Martin", "testUser7", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasNums}},
		{"No Numbers in Lname", []string{"lname"}, "Jon", "Martin5", "testUser8", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasNums}},
		{"No Symbols in Address", []string{"address"}, "Jon", "Martin", "testUser9", "@409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasSymbols}},
		{"Missing '@' Symbol in Email", []string{"email"}, "Jon", "Martin", "testUser10", "409 Alistar Road", "dummy59gmail.com", []byte("passwordPomp432"), []error{errReqSymbol}},
		{"Invalid ending address in Email", []string{"email"}, "Jon", "Martin", "testUser12", "409 Alistar Road", "dummy@gmail.dum", []byte("passwordPomp432"), []error{errReqEndingAddr}},
		{"Missing at least One Capital Letter in Password", []string{"password"}, "Jon", "Martin", "testUser13", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordpomp432"), []error{missingCapitalLetter}},
		{"Password is empty", []string{"password"}, "Jon", "Martin", "testUser14", "409 Alistar Road", "dummy59@gmail.com", []byte(""), []error{emptyPswd}},
		{"Password is too short", []string{"password", "password", "password"}, "Jon", "Martin", "testUser15", "409 Alistar Road", "dummy59@gmail.com", []byte("prompt"), []error{shortPswd, missingCapitalLetter, pswdHasNoDigits}},
		{"Password is too long", []string{"password"}, "Jon", "Martin", "testUser16", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp4323349484734743743233"), []error{longPswd}},
		{"Password has no digits", []string{"password", "password", "password"}, "Jon", "Martin", "testUser17", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordpomppppp"), []error{missingCapitalLetter, pswdHasNoDigits}}, {"Password has symbols", []string{"password"}, "Jon", "Martin", "testUser18", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp43!"), []error{pswdHasSymbols}},
	}

	// range over table tests and validate the right errs are being thrown
	for _, tt := range tests {
		req, _ := http.NewRequest("POST", "/view", nil)
		w := httptest.NewRecorder()
		c := r.NewContext(req, w)
		c.SetPath("/posts")
		t.Run(tt.testName, func(t *testing.T) {
			tempAcct := Account{Fname: tt.fname, Lname: tt.lname, Username: tt.username, Address: tt.address, Email: tt.email, Password: tt.password}
			actualErrLst := VetAllFields(tempAcct)
			expectedErrLst := printAcctErrs(tt.attribNames, tt.errLst)
			assert.Equal(t, expectedErrLst, actualErrLst, expectedErrLst)

		})
	}
}

func Test_userlogin(t *testing.T) {
	r := setupEchoClient()

	invalidUserCreds := errors.New("incorrect email or password.")
	failedDBConn := errors.New("database error: couldn’t connect to drumstick")
	tests := []struct {
		testName string
		username string
		password string
	}{
		{"Empty values", "", ""},
		{"bad password", "dummy2", "dummy54Dwdgfdgf"},
		{"bad username", "dummy2", "mldKMuffinDrop4"},
		{"password is too long", "dummy2", "lkdrWkkkkkkkkk"},
		{"password is too short", "dummy2", "dfdf"},
		{"password is good", "dummy@gmail.com", "superSecretP432"},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest("POST", "/posts", nil)
		w := httptest.NewRecorder()
		c := r.NewContext(req, w)
		c.SetPath("/posts")
		t.Run(tt.testName, func(t *testing.T) {
			actualErr := CompareUserCreds(tt.username, tt.password)

			if errors.Is(actualErr, invalidUserCreds) {
				assert.EqualError(t, actualErr, invalidUserCreds.Error())
			}

			if errors.Is(actualErr, failedDBConn) {
				assert.EqualError(t, actualErr, failedDBConn.Error())
			}

			if actualErr == nil {
				// the case that the usercreds are legit
				assert.Equal(t, nil, actualErr)
			}
		})
	}
}

// test userid retrieval by email
func Test_UserIDByEmail(t *testing.T) {
	rscNotFound := errors.New("error: resource not found: id does not exist")
	sample := "dummy@gmail.com"
	actual, actualErr := UserIDByEmail(sample)

	if errors.Is(actualErr, pgx.ErrNoRows) {
		assert.EqualError(t, actualErr, pgx.ErrNoRows.Error())
	}

	if errors.Is(actualErr, rscNotFound) {
		assert.EqualError(t, actualErr, rscNotFound.Error())
	}

	if actualErr == nil {
		assert.Equal(t, 1, actual)
	}
}
