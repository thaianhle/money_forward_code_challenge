package exception

import (
	"fmt"
	//"go-micro/internal/domain/transaction/models"
)

type CheckExceptionTransactionType struct {
	name      string
	expecteds []string
	errString string
}

func NewCheckExceptionTransactionType(expecteds []string) *CheckExceptionTransactionType {
	return &CheckExceptionTransactionType{
		name:      "InValidExceptionTransactionType",
		expecteds: expecteds,
	}
}

func constructErrString(name string, expect any, got any) string {
	// ClassNameError - ErrType ->>> expects: expect_value, !got: got_value
	return fmt.Sprintf("[%s] ->>> expects: [%v], !got: [%v]", name, expect, got)
}

func (i *CheckExceptionTransactionType) getExpectToString() string {
	v := ""
	for i, expect := range i.expecteds {
		if i > 0 {
			v += ", "
		}
		v += fmt.Sprintf("%v", expect)
	}

	return v
}
func (i *CheckExceptionTransactionType) Check(value any) {
	if i.errString != "" {
		return
	}

	if _, ok := value.(string); !ok {
		i.errString = constructErrString(i.name, "string", "unknown")
	}

	transactionType := fmt.Sprintf("%s", value)

	for _, expect := range i.expecteds {
		if expect == transactionType {
			return
		}
	}

	i.errString = constructErrString(i.name, i.getExpectToString(), transactionType)
}

func (i *CheckExceptionTransactionType) Error() string {
	return i.errString
}
