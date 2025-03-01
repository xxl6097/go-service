package utils

import (
	"fmt"
	"math"
	"strings"
)

// PrintByteArrayAsConstant 把字节数组以常量字节数组的形式打印出来
func PrintByteArrayAsConstant(bytes []byte) string {
	sb := strings.Builder{}
	sb.WriteString("[]byte{")
	for i, b := range bytes {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("0x%02X", b))
	}
	sb.WriteString("}")
	return sb.String()
}

// DivideAndCeil 函数用于进行除法并向上取整
func DivideAndCeil(a, b int) int {
	// 将整数转换为 float64 类型进行除法运算
	result := float64(a) / float64(b)
	// 使用 math.Ceil 函数进行向上取整
	result = math.Ceil(result)
	// 将结果转换回整数类型
	return int(result)
}

func Divide(a, b int) int {
	return DivideAndCeil(a, b) * b
}
