package exception

import (
	"fmt"
)

type CheckExceptionBankTypeAccount struct {
	name      string
	expecteds []string
	errString string
}

func NewCheckExceptionBankTypeAccount(expecteds []string) *CheckExceptionBankTypeAccount {
	return &CheckExceptionBankTypeAccount{
		name:      "InValidExceptionBankTypeAccount",
		expecteds: expecteds,
	}
}

func (i *CheckExceptionBankTypeAccount) getExpectToString() string {
	v := ""
	for i, expect := range i.expecteds {
		v += fmt.Sprintf("%v", expect)
		if i > 0 {
			v += ", or"
		}
	}
	return v
}

func (i *CheckExceptionBankTypeAccount) Check(value any) {
	if i.errString != "" {
		return
	}

	if _, ok := value.(string); !ok {
		i.errString = constructErrString(i.name, "string", "unknown")
	}

	bank_type := value.(string)
	for _, expect := range i.expecteds {
		if expect == bank_type {
			return
		}
	}

	i.errString = constructErrString(i.name, i.getExpectToString(), bank_type)
}

func (i *CheckExceptionBankTypeAccount) Error() string {
	return i.errString
}
