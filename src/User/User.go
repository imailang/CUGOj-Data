package user

import (
	httptool "CUGOj-Data/src/HttpTool"
	sqltool "CUGOj-Data/src/SqlTool"
	"fmt"
	"io/ioutil"
	"net/http"
)

func AddUser(w http.ResponseWriter, r *http.Request) {
	httptool.Add(&sqltool.UserInfo{}, w, r)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	httptool.Delete(&sqltool.UserInfo{}, w, r)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	httptool.Query(&sqltool.UserInfo{}, w, r, "password")
}
func ChangeUser(w http.ResponseWriter, r *http.Request) {
	httptool.Change(&sqltool.UserInfo{}, w, r)
}
func GetBaseUser(w http.ResponseWriter, r *http.Request) {
	httptool.Base(&sqltool.UserInfo{}, w, r)
}
func GetUserList(w http.ResponseWriter, r *http.Request) {
	httptool.List(&[]sqltool.UserInfo{}, w, r, "password")
}
func CheckUsername(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "HTTP请求读取出错"))
		return
	}
	db, err := httptool.GetDB()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var count int64
	result := db.Model(&sqltool.UserInfo{}).Where("Username = ?", string(buf)).Count(&count)
	if result.Error != nil {
		w.Write(httptool.DisposeQueryResult(result))
		return
	}
	w.Write(httptool.SuccessBuf(fmt.Sprint(count)))
}
func CheckNickname(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "HTTP请求读取出错"))
		return
	}
	db, err := httptool.GetDB()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var count int64
	result := db.Model(&sqltool.UserInfo{}).Where("Nickname = ?", string(buf)).Count(&count)
	if result.Error != nil {
		w.Write(httptool.DisposeQueryResult(result))
		return
	}
	w.Write(httptool.SuccessBuf(fmt.Sprint(count)))
}
func CheckEmail(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "HTTP请求读取出错"))
		return
	}
	db, err := httptool.GetDB()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var count int64
	result := db.Model(&sqltool.UserInfo{}).Where("Email = ?", string(buf)).Count(&count)
	if result.Error != nil {
		w.Write(httptool.DisposeQueryResult(result))
		return
	}
	w.Write(httptool.SuccessBuf(fmt.Sprint(count)))
}

func GetUserCount(w http.ResponseWriter, r *http.Request) {
	httptool.Count(&sqltool.UserInfo{}, w, r)
}
