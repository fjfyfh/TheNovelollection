/**************************
Golang数组去重&切片去重
****************************/
package tool

import (
	"sort"
	// "strings"
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/axgle/mahonia"
	"github.com/chain-zhang/pinyin"
)

func GetfontNum(str string) (n int) {

	n = utf8.RuneCountInString(str)
	return n
}

//先将原切片（数组）进行排序，在将相邻的元素进行比较，如果不同则存放在新切片（数组）中。
func RemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	sort.Strings(arr)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

//获取文件大小
func Getfilesize(filepath string) (filesize string) {

	fileinfo, err := os.Stat(filepath)
	if err != nil {
		fmt.Printf("os.Stat %s \n", err)
		return
	}
	size := float32(fileinfo.Size()) / float32(1024*1024) //比特 、字节 、兆
	// fmt.Printf("当前的文档大小是 %.2f MB \t , 名称是 %s \n", filesize, fileinfo.Name())

	buf := fmt.Sprintf("%6.2f", size) // 将数值转化格式化

	// fmt.Printf("buf = %s \n", buf+"MB")
	filesize = buf + "MB"
	return filesize

}

//src为要转换的字符串，srcCode为待转换的编码格式，targetCode为要转换的编码格式
func ConvertTostring(src string, srcCode string, targetCode string) (str string) {

	Decoder := mahonia.NewDecoder(srcCode)
	// fmt.Println(enc.ConvertString("hello,双手都垂到水里"))
	str = Decoder.ConvertString(src)
	return
}

func FontToPinyin(str string) string {

	str, err := pinyin.New(str).Split("").Mode(3).Convert()
	if err != nil {
		// 错误处理
	} /*else {
		// fmt.Println(str)
		fmt.Sprintf("%s", str)
	}*/
	return str
	// str, err := pinyin.New("我是中国人").Split(" ").Mode(pinyin.WithoutTone).Convert()
	// if err != nil {
	// 	// 错误处理
	// } else {
	// 	fmt.Println(str)
	// }

	// str, err := pinyin.New("我是中国人").Split("-").Mode(pinyin.Tone).Convert()
	// if err != nil {
	// 	// 错误处理
	// } else {
	// 	fmt.Println(str)
	// }

	// str, err := pinyin.New("我是中国人").Convert()
	// if err != nil {
	// 	// 错误处理
	// } else {
	// 	fmt.Println(str)
	// }
}
