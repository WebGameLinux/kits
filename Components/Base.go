package Components

import (
		"bytes"
		"github.com/prometheus/common/log"
		"github.com/webGameLinux/kits/Contracts"
		"io"
		"io/ioutil"
		"os"
		"path/filepath"
		"unicode"
)

const (
		FileNotExists       = -1
		FilePermission      = -2
		FileExistsError     = -3
		TimeOutError        = -4
		FileStateOtherError = -5
		IsDirFlag           = 2
		FileFlag            = 1
		IsEmptyFileFlag     = 10
)

type SupportBeanExtend struct {
		Contracts.SupportBean
}

func BeanOf() Contracts.SupportInterface {
		var bean = new(SupportBeanExtend)
		bean.Boot = true
		bean.Register = true
		return bean
}

func (this *SupportBeanExtend) HasRegister() bool {
		return this.Register
}

func (this *SupportBeanExtend) HasBoot() bool {
		return this.Boot
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

func IsFile(fs string) int {
		state, err := GetFileState(fs)
		if err != nil {
				if os.IsNotExist(err) {
						return FileNotExists
				}
				if os.IsPermission(err) {
						return FilePermission
				}
				if os.IsExist(err) {
						return FileExistsError
				}
				if os.IsTimeout(err) {
						return TimeOutError
				}
				return FileStateOtherError
		}
		if state.IsDir() {
				return IsDirFlag
		}
		if state.Size() > 0 {
				return FileFlag
		}
		return IsEmptyFileFlag
}

func IsEmptyFile(fs string) bool {
		if flag := IsFile(fs); flag == IsEmptyFileFlag {
				return true
		}
		return false
}

func IsDir(fs string) bool {
		if flag := IsFile(fs); flag == IsDirFlag {
				return true
		}
		return false
}

func IsEmptyDir(dir string) bool {
		if !IsDir(dir) {
				return false
		}
		if size, err := GetDirSize(dir); err == nil && size > 0 {
				return false
		}
		return true
}

func GetDirSize(dir string) (int64, error) {
		var size int64
		err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
				if !info.IsDir() {
						size += info.Size()
				}
				return err
		})
		return size, err
}

// 构造文件组
func MakeFiles(fs string, filter ...func(fs string) bool) []string {
		flag := IsFile(fs)
		if flag <= 0 {
				return []string{}
		}
		if flag == IsDirFlag {
				var files []string
				err := filepath.Walk(fs, func(path string, info os.FileInfo, err error) error {
						if info.IsDir() {
								return err
						}
						file, e := filepath.Abs(path)
						if e != nil {
								return err
						}
						if len(filter) > 0 {
								for _, fn := range filter {
										if !fn(file) {
												return err
										}
								}
						}
						files = append(files, file)
						return err
				})
				if err == nil {
						return files
				}
				return []string{}
		}
		return []string{fs}
}

// 是否debug
func Debug() bool  {
		value:=EnvironmentProviderOf().Get(Contracts.AppDebug,"false")
		boolean :=BooleanOf(value)
		if boolean.Invalid() {
				return false
		}
		return boolean.ValueOf()
}