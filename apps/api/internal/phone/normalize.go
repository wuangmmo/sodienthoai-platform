package phone

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidNumber = errors.New("invalid phone number")

func Normalize(input string) (string,error){ return NormalizeForCountry(input,"") }

func NormalizeForCountry(input,country string)(string,error){
	input=strings.TrimSpace(input); if input=="" { return "",ErrInvalidNumber }
	var b strings.Builder
	for i,r:=range input {
		switch {
		case unicode.IsDigit(r): b.WriteRune(r)
		case r=='+'&&i==0:
		case r==' ',r=='-',r=='(',r==')',r=='.':
		default:return "",ErrInvalidNumber
		}
	}
	digits:=b.String()
	if strings.HasPrefix(input,"+") {
		if len(digits)<8||len(digits)>15||digits[0]=='0' { return "",ErrInvalidNumber }
		return "+"+digits,nil
	}
	if strings.EqualFold(country,"VN") {
		if len(digits)<9||len(digits)>11||digits[0]!='0' { return "",ErrInvalidNumber }
		national:=strings.TrimLeft(digits,"0")
		if national=="" { return "",ErrInvalidNumber }
		e164:="+84"+national
		if len(e164)-1>15 { return "",ErrInvalidNumber }
		return e164,nil
	}
	return "",ErrInvalidNumber
}
