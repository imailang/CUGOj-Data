package filetool

import (
	"bytes"
	"io/ioutil"
	"os/exec"
)

// var Compressers = map[string]func(string, string) error{
// 	"zip": ZipComp,
// }
var DeCompressers = map[string](func(string) error){
	".zip": ZipUnComp,
	".tar": TarUnComp,
	".7z":  Q7ZipUnComp,
	".rar": RarUnComp,
}
var Exts = []string{
	".zip",
	".tar",
	".7z",
	".rar",
}

type Error struct {
	Info string
}

func (e Error) Error() string {
	return e.Info
}

// func ZipComp(in string, out string) error {

// 	return nil
// }

func ZipUnComp(in string) error {
	return Exec("unzip", "-o", in, "-d", in[:len(in)-4])
}

func RarUnComp(in string) error {
	return Exec("unrar", "x", in, in[:len(in)-4])
}

func Q7ZipUnComp(in string) error {
	return Exec("7z", "x", in, "-r", "-o"+in[:len(in)-3])
}

func TarUnComp(in string) error {
	return Exec("tar", "-zxvf", in, "-C", in[:len(in)-4])
}

func UnComp(id, in, ext string) error {
	f, ok := DeCompressers[ext]
	if ok {
		return f(Home() + "data/problems/" + id + "/files/" + in)
	}
	return Error{Info: "未知文件后缀"}
}

func Exec(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	command.Stderr = &bytes.Buffer{}
	err := command.Run()
	if err != nil {
		errInfo, _ := ioutil.ReadAll(command.Stderr.(*bytes.Buffer))
		return Error{
			Info: err.Error() + "\n" + string(errInfo),
		}
	}
	return nil
}

func CheckExt(ext string) bool {
	for _, str := range Exts {
		if str == ext {
			return true
		}
	}
	return false
}
