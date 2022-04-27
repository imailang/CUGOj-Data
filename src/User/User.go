package user

import (
	httptool "CUGOj-Data/src/HttpTool"
	sqltool "CUGOj-Data/src/SqlTool"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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

func GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "HTTP请求读取出错"))
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	model := sqltool.UserInfo{}
	result := db.Where("username=?", string(buf)).First(&model)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	buf = httptool.GetJson(w, r, model, "Password")
	if buf == nil {
		return
	}
	w.Write(buf)
}

func CheckPassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	model := sqltool.UserInfo{}
	if !httptool.UnMarshal(w, r, &model) {
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	result := db.Where("username=? and binary password=?", model.Username, model.Password).First(&model)

	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	w.Write(httptool.SuccessBuf(fmt.Sprint(model.ID)))
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
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	var count int64
	result := db.Model(&sqltool.UserInfo{}).Where("Username = ?", string(buf)).Count(&count)
	if !httptool.DisposeQueryResult(w, r, result) {
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
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	var count int64
	result := db.Model(&sqltool.UserInfo{}).Where("Nickname = ?", string(buf)).Count(&count)
	if !httptool.DisposeQueryResult(w, r, result) {
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
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	var count int64
	result := db.Model(&sqltool.UserInfo{}).Where("Email = ?", string(buf)).Count(&count)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	w.Write(httptool.SuccessBuf(fmt.Sprint(count)))
}

func GetUserCount(w http.ResponseWriter, r *http.Request) {
	httptool.Count(&sqltool.UserInfo{}, w, r)
}

func GetSubmmitCount(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ID := 0
	if !httptool.UnMarshal(w, r, &ID) {
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	model := sqltool.Judge{}
	res := make(map[string]int64)
	tmp := int64(0)

	result := db.Model(model).Where("uid = ?", ID).Where("submit_time >= ?", time.Now().Add(-7*24*time.Hour)).Count(&tmp)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	res["seven"] = tmp

	result = db.Model(model).Where("uid = ?", ID).Where("submit_time >= ?", time.Now().Add(-30*24*time.Hour)).Count(&tmp)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	res["thirty"] = tmp

	result = db.Model(model).Where("uid = ?", ID).Count(&tmp)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	res["all"] = tmp

	result = db.Model(model).Where("status = ?", "AC").Where("uid = ?", ID).Where("submit_time >= ?", time.Now().Add(-7*24*time.Hour)).Count(&tmp)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	res["sevenAC"] = tmp

	result = db.Model(model).Where("status = ?", "AC").Where("uid = ?", ID).Where("submit_time >= ?", time.Now().Add(-30*24*time.Hour)).Count(&tmp)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	res["thirtyAC"] = tmp

	result = db.Model(model).Where("status = ?", "AC").Where("uid = ?", ID).Count(&tmp)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	res["allAC"] = tmp

	buf := httptool.GetJson(w, r, &res)
	if buf == nil {
		return
	}
	w.Write(httptool.SuccessBuf(string(buf)))
}

func CheckAdminByID(ID uint) bool {
	db := sqltool.GetDB()
	if db == nil {
		return false
	}
	model := sqltool.UserInfo{}
	result := db.Find(&model, ID)
	if result.Error != nil {
		return false
	}
	if result.RowsAffected == 0 {
		return false
	}
	titles := strings.Split(model.Title, ";")
	for _, str := range titles {
		if str == "管理员" {
			return true
		}
	}
	return false
}
