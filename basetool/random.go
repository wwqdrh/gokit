package basetool

import (
	"math/rand"
	"time"
)

const (
	Letters  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numbers  = "0123456789"
	randChar = "0123456789abcdefghigklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// 生成随机字符串
func RandString(l int) string {
	strList := []byte(randChar)

	result := []byte{}
	i := 0

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	charLen := len(strList)
	for i < l {
		new := strList[r.Intn(charLen)]
		result = append(result, new)
		i = i + 1
	}
	return string(result)
}

func RandomLetters(length int) string {
	res := make([]byte, length)
	for i := 0; i < length; i++ {
		res[i] = Letters[rand.Intn(len(Letters))]
	}
	return string(res)
}

func RandomNumbers(length int) string {
	res := make([]byte, length)
	for i := 0; i < length; i++ {
		res[i] = Numbers[rand.Intn(len(Numbers))]
	}
	return string(res)
}
