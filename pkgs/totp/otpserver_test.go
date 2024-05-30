package totp

import (
	"sync"
	"testing"
	"time"
)

func TestOTP(t *testing.T) {
	totp := NewTOTP("12345678901234567890")

	user_ids := []string{"test1", "test2", "test3", "test4", "test5"}

	waitgroup := &sync.WaitGroup{}
	codes := make([][2]string, len(user_ids))
	for index, user_id := range user_ids {
		tnow := time.Now()
		tzero := int64(tnow.Second())
		first_code := totp.GenCode(user_id, tnow.Unix(), tzero)
		waitgroup.Add(1)
		go recheck_otp(user_id, tzero, first_code, totp, &codes, index, waitgroup)
	}

	waitgroup.Wait()

	for i, code := range codes {
		success := code[0] == code[1]
		v := "FAILED"
		if success {
			v = "SUCCESS"
		}
		t.Logf("case [user_id = %v]: success[%v] %v==%v\n", user_ids[i], v, code[0], code[1])
	}
}

func recheck_otp(user_id string, tzero int64, first_code string, totp *TOTP, after_codes *[][2]string, index int, waitgroup *sync.WaitGroup) {
	defer waitgroup.Done()
	time.Sleep(time.Second)
	tnow := time.Now()
	code := totp.GenCode(user_id, tnow.Unix(), int64(tzero))
	(*after_codes)[index][0] = first_code
	(*after_codes)[index][1] = code
}
