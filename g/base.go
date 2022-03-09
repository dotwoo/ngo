// Description:
// Author: dotwoo@gmail.com
// Github: https://github.com/dotwoo
// Date: 2022-03-09 12:48:39
// FilePath: /ngo/g/check.go
//
package g

// CheckError 提供简介的error判断，如果err != nil则panic
func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

// Containt 判断字符串是否在列表里
func Containt(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
