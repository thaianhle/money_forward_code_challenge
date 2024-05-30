package exception

import "fmt"

type CheckErrAmountValue struct {
	name      string
	min       float32
	max       float32
	errString string
}

func NewCheckErrAmountValue(min float32, max float32) *CheckErrAmountValue {
	return &CheckErrAmountValue{
		name: "InValidErrAmountValue",
		min:  min,
		max:  max,
	}
}

func (i *CheckErrAmountValue) getExpectToString() string {
	return fmt.Sprintf("(min=%s, max=%s)", fmt.Sprintf("%.0f", i.min), fmt.Sprintf("%.0f", i.max))
}

func (i *CheckErrAmountValue) Check(value any) {
	if i.errString != "" {
		return
	}

	if _, ok := value.(float32); !ok {
		i.errString = constructErrString(i.name, "float32", "unknown")
		return
	}

	amount := value.(float32)
	if amount < i.min || amount > i.max {
		i.errString = constructErrString(i.name, i.getExpectToString(), amount)
	}
}

func (i *CheckErrAmountValue) Error() string {
	return i.errString
}
