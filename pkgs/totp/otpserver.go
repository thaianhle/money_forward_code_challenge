package totp

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"hash"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var default_Key = []byte("PdF4MAi.%ruybwDStE7ihe5@z*SbDs69")
var defaultTimeOTP = 45

type TOTP struct {
	digits   int
	hashFunc func() hash.Hash
	interval int64
	secret   []byte
}

func NewTOTP(key ...string) *TOTP {
	t := &TOTP{
		digits:   8,
		hashFunc: sha256.New,
		interval: int64(defaultTimeOTP),
	}
	if len(key) == 0 {
		t.secret = []byte(default_Key)
	} else {
		t.secret = []byte(key[0])
	}
	return t
}

func (t *TOTP) GenTimeBase() int64 {
	v := time.Now().Unix()
	return v
}

func (t *TOTP) ConvertTimeToHex(time int64, tzero ...int64) string {
	var t0 int64 = 0
	if len(tzero) == 1 {
		t0 = tzero[0]
	}
	x := (time - t0) / t.interval
	v := strings.ToUpper(strconv.FormatInt(x, 16))
	for len(v) < 16 {
		v = "0" + v
	}
	return v
}

func (t *TOTP) HextoBytes(hex string) []byte {
	bigInt := new(big.Int)
	bigInt.SetString("10"+hex, 16)
	bArray := bigInt.Bytes()
	var ret []byte = make([]byte, len(bArray)-1)
	for i := 0; i < len(ret); i++ {
		ret[i] = bArray[i+1]
	}
	return ret
}

func (t *TOTP) hexTimeToBytes(hex string) []byte {
	return nil
}
func (t *TOTP) GenCode(user_id string, timeAt int64, tzero ...int64) string {
	timeHex := t.ConvertTimeToHex(timeAt, tzero...)
	counterBytes := t.HextoBytes(timeHex)
	hashser := hmac.New(t.hashFunc, t.secret)
	//fmt.Printf("counter hex: %v\n", counterBytes)
	//hexUserId := hex.EncodeToString([]byte(fmt.Sprintf("%v", user_id)))
	hashser.Write(counterBytes)
	hashser.Write([]byte(user_id))
	hmacHex := hashser.Sum(nil)

	// RFC 6238 - TOTP: Time-Based One-Time Password
	offset := int(hmacHex[len(hmacHex)-1] & 0xf)
	code := ((int(hmacHex[offset]) & 0x7f) << 24) |
		((int(hmacHex[offset+1] & 0xff)) << 16) |
		((int(hmacHex[offset+2] & 0xff)) << 8) |
		(int(hmacHex[offset+3]) & 0xff)

	code = code % int(math.Pow10(t.digits))
	codeString := fmt.Sprintf("%v", code)
	for len(codeString) < t.digits {
		codeString = "0" + codeString
	}
	return codeString
}
func (t *TOTP) GenCodeExpr(user_id string) (string, int64) {
	time_at := t.GenTimeBase()
	time_at_expr := (time_at) + 1*int64(t.interval)
	code := t.GenCode(user_id, time_at)

	fmt.Println("delta: ", time_at_expr-time_at)
	return code, time_at_expr
}
