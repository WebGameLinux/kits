package Components

import (
		"bytes"
		"github.com/prometheus/common/log"
		"github.com/webGameLinux/kits/Contracts"
		"io"
		"io/ioutil"
		"os"
		"unicode"
)

func BeanOf() *Contracts.SupportBean {
		var bean = new(Contracts.SupportBean)
		bean.Boot = true
		bean.Register = true
		return bean
}

// 打开文件
func FileOpen(fs string) (*os.File, error) {
		_, err := GetFileState(fs)
		if err == nil {
				if f, err := os.Open(fs); err == nil {
						return f, nil
				}
		}
		return nil, err
}

// 获取文件状态
func GetFileState(fs string) (os.FileInfo, error) {
		return os.Stat(fs)
}

// 获取文件读取reader
func GetFileReader(fs string) io.Reader {
		if file, err := FileOpen(fs); err == nil {
				defer Close(file)
				if by, err := ioutil.ReadAll(file); err == nil {
						return bytes.NewReader(by)
				}
		}
		return bytes.NewReader([]byte(""))
}

// 文件关闭
func Close(closer io.Closer) {
		if err := closer.Close(); err != nil {
				log.Error(err)
		}
}

// 是否数字
func IsNumber(str string, formatter ...func(string) string) bool {
		if len(formatter) > 0 {
				for _, handler := range formatter {
						if str == "" {
								return false
						}
						str = handler(str)
				}
		}
		if str == "" {
				return false
		}
		chars := []rune(str)
		for _, char := range chars {
				if !unicode.IsNumber(char) {
						return false
				}
		}
		return true
}

