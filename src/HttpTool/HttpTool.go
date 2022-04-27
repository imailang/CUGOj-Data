package httptool

import (
	properties "CUGOj-Data/src/Properties"
	sqltool "CUGOj-Data/src/SqlTool"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type Response struct {
	Statu string
	Info  string
}

func UnMarshal(w http.ResponseWriter, r *http.Request, model any) bool {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		buf, _ = json.Marshal(&Response{
			Statu: "019",
			Info:  err.Error(),
		})
		w.Write(buf)
		return false
	}
	err = json.Unmarshal(buf, model)
	if err != nil {
		buf, _ = json.Marshal(&Response{
			Statu: "018",
			Info:  err.Error(),
		})
		w.Write(buf)
		return false
	}
	return true
}

type Error struct {
	Info string
}

func (e Error) Error() string {
	return e.Info
}

func ResponseBuf(status, info string) []byte {
	buf, _ := json.Marshal(&Response{
		Statu: status,
		Info:  info,
	})
	return buf
}

func DbError() []byte {
	buf, _ := json.Marshal(&Response{
		Statu: "020",
		Info:  "数据库链接失效",
	})
	return buf
}

func GetDB(w http.ResponseWriter, r *http.Request) *gorm.DB {
	db := sqltool.GetDB()
	if db == nil {
		buf, _ := json.Marshal(&Response{
			Statu: "020",
			Info:  "数据库链接失效",
		})
		w.Write(buf)
		return nil
	}
	return db
}

func DisposeAddResult(w http.ResponseWriter, r *http.Request, result *gorm.DB) bool {
	if result.Error != nil {
		buf, _ := json.Marshal(&Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return false
	}
	return true
}

func DisposeQueryResult(w http.ResponseWriter, r *http.Request, result *gorm.DB) bool {
	if result.Error != nil {
		buf, _ := json.Marshal(&Response{
			Statu: "022",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return false
	}
	return true
}

func SuccessBuf(info string) []byte {
	buf, _ := json.Marshal(&Response{
		Statu: "000",
		Info:  info,
	})
	return buf
}

func GetJson(w http.ResponseWriter, r *http.Request, model any, ignores ...string) []byte {
	buf, err := json.Marshal(model)
	if err != nil {
		w.Write(ResponseBuf("005", "Json序列化出现问题"))
		return nil
	}
	tmp := make(map[string]interface{})
	err = json.Unmarshal(buf, &tmp)
	if err != nil {
		return buf
	}
	for _, key := range ignores {
		delete(tmp, key)
	}
	buf, err = json.Marshal(&tmp)
	if err != nil {
		w.Write(ResponseBuf("005", "Json序列化出现问题"))
		return nil
	}
	return buf
}

func Add(model sqltool.SqlModel, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if !UnMarshal(w, r, model) {
		return
	}

	db := GetDB(w, r)
	if db == nil {
		return
	}

	result := db.Create(model)
	if !DisposeAddResult(w, r, result) {
		return
	}
	w.Write(SuccessBuf(model.GetID()))
}

func Query(model any, w http.ResponseWriter, r *http.Request, omi ...string) {
	defer r.Body.Close()

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(ResponseBuf("019", "Http请求读取出错"))
		return
	}

	db := GetDB(w, r)
	if db == nil {
		return
	}
	for _, str := range omi {
		db = db.Omit(str)
	}

	ID, err := strconv.Atoi(string(buf))
	if err != nil {
		w.Write(ResponseBuf("019", "主键参数应为数字"))
		return
	}

	result := db.Find(model, ID)

	if !DisposeQueryResult(w, r, result) {
		return
	}
	if result.RowsAffected == 0 {
		w.Write(ResponseBuf("022", "查询的数据不存在"))
		return
	}

	buf = GetJson(w, r, model)
	if buf == nil {
		return
	}
	w.Write(buf)
}

func Delete(model any, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(ResponseBuf("019", "Http请求读取出错"))
		return
	}

	db := GetDB(w, r)
	if db == nil {
		return
	}

	result := db.Delete(model, string(buf))

	if !DisposeQueryResult(w, r, result) {
		return
	}
	w.Write(SuccessBuf("删除成功"))
}

func Change(model any, w http.ResponseWriter, r *http.Request, omi ...string) {
	defer r.Body.Close()
	if !UnMarshal(w, r, model) {
		return
	}

	db := GetDB(w, r)
	if db == nil {
		return
	}

	for _, str := range omi {
		db = db.Omit(str)
	}
	result := db.Updates(model)
	if !DisposeAddResult(w, r, result) {
		return
	}
	w.Write(SuccessBuf("修改成功"))
}

func Base(model any, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf := GetJson(w, r, model)
	if buf == nil {
		return
	}
	w.Write(buf)
}

func Count(model any, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	odds := make(map[string]interface{})

	if !UnMarshal(w, r, &odds) {
		return
	}

	db := sqltool.GetDB()
	if db == nil {
		buf, _ := json.Marshal(&Response{
			Statu: "020",
			Info:  "数据库链接失效",
		})
		w.Write(buf)
		return
	}

	db = db.Model(model)
	oddCnt := 0
	for _, odd := range odds {
		if oddCnt == 0 {
			db = db.Where(SetWhere(odd, db))
		} else {
			db = db.Or(SetWhere(odd, db))
		}
		oddCnt++
		if db.Error != nil {
			w.Write(ResponseBuf("022", db.Error.Error()))
			return
		}
	}

	var cnt int64

	result := db.Count(&cnt)
	if result.Error != nil {
		buf, _ := json.Marshal(&Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}
	w.Write(SuccessBuf(fmt.Sprint(cnt)))
}

func SetWhere(odd interface{}, db *gorm.DB) *gorm.DB {
	switch odd.(type) {
	case map[string]interface{}:

	default:
		return db
	}
	for k, v := range odd.(map[string]interface{}) {
		key := strings.TrimSpace(k)
		if k == "order" {
			db = db.Order(v)
		} else if strings.HasSuffix(key, "?") {
			db = db.Where(k, v)
		} else {
			db = db.Where(k+" = ?", v)
		}
	}
	return db
}

func List(model any, w http.ResponseWriter, r *http.Request, omi ...string) {
	defer r.Body.Close()

	var qu interface{}

	odds := make(map[string]interface{})

	if !UnMarshal(w, r, &odds) {
		return
	}

	db := sqltool.GetDB()
	if db == nil {
		buf, _ := json.Marshal(&Response{
			Statu: "020",
			Info:  "数据库链接失效",
		})
		w.Write(buf)
		return
	}

	oddCnt := 0
	for key, odd := range odds {
		if key == "pagequery" {
			continue
		}
		if oddCnt == 0 {
			db = db.Where(SetWhere(odd, db))
		} else {
			db = db.Or(SetWhere(odd, db))
		}
		if db.Error != nil {
			w.Write(ResponseBuf("022", db.Error.Error()))
			return
		}
		oddCnt++
	}

	for _, str := range omi {
		db = db.Omit(str)
	}

	qu, ok := odds["pagequery"]
	if ok {
		qump := qu.(map[string]interface{})
		offset, ok := qump["offset"]
		if !ok {
			w.Write(ResponseBuf("019", "pagequery缺少offset字段"))
			return
		}
		pagesize, ok := qump["pagesize"]
		if !ok {
			w.Write(ResponseBuf("019", "pagequery缺少pagesize字段"))
			return
		}
		off, err := strconv.Atoi(fmt.Sprint(offset))
		if err != nil {
			w.Write(ResponseBuf("019", err.Error()))
			return
		}
		siz, err := strconv.Atoi(fmt.Sprint(pagesize))
		if err != nil {
			w.Write(ResponseBuf("019", err.Error()))
			return
		}
		db = db.Offset(off * siz).Limit(siz)
	}

	result := db.Find(model)
	if result.Error != nil {
		buf, _ := json.Marshal(&Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}
	buf, err := json.Marshal(model)
	if err != nil {
		buf, _ = json.Marshal(&Response{
			Statu: "005",
			Info:  err.Error(),
		})
		w.Write(buf)
		return
	}
	w.Write(SuccessBuf(string(buf)))
}

func GetIDFromTokenNotWrite(r *http.Request) uint {
	token, err := r.Cookie("cugtoken")
	if err != nil {
		return 0
	}
	cugtoken := token.String()
	if !strings.HasPrefix(cugtoken, "cugtoken=") {
		return 0
	}
	cugtoken = cugtoken[9:]
	ip, err := properties.Get("LoginIP")
	if err != nil {
		return 0
	}
	port, err := properties.Get("LoginPort")
	if err != nil {
		return 0
	}

	res, err := http.Post("http://"+ip+":"+port+"/ojLogin/userInfo/getLoginIdByToken", "application/json", strings.NewReader("{\"cugtoken\":\""+cugtoken+"\"}"))
	if err != nil {
		return 0
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0
	}
	resModel := make(map[string]interface{})

	err = json.Unmarshal(buf, &resModel)
	if err != nil {
		return 0
	}

	IDStr, ok := resModel["data"]
	if !ok || IDStr == nil {
		return 0
	}
	ID, err := strconv.Atoi(IDStr.(string))
	if err != nil {
		return 0
	}
	// if fmt.Sprint(ID) != readID {
	// 	w.Write(ResponseBuf("032", "非法用户"))
	// 	return 0
	// }
	return uint(ID)
}

func GetIDFromToken(w http.ResponseWriter, r *http.Request) uint {
	// buf, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	w.Write(ResponseBuf("019", "Http请求出错"))
	// 	return 0
	// }
	// readID := string(buf)
	token, err := r.Cookie("cugtoken")
	if err != nil {
		w.Write(ResponseBuf("032", "非法用户"))
		return 0
	}
	cugtoken := token.String()
	if !strings.HasPrefix(cugtoken, "cugtoken=") {
		w.Write(ResponseBuf("032", "非法用户"))
		return 0
	}
	cugtoken = cugtoken[9:]
	ip, err := properties.Get("LoginIP")
	if err != nil {
		w.Write(ResponseBuf("030", "配置文件中不存在LoginIP"))
		return 0
	}
	port, err := properties.Get("LoginPort")
	if err != nil {
		w.Write(ResponseBuf("030", "配置文件中不存在LoginPort"))
		return 0
	}

	res, err := http.Post("http://"+ip+":"+port+"/ojLogin/userInfo/getLoginIdByToken", "application/json", strings.NewReader("{\"cugtoken\":\""+cugtoken+"\"}"))
	if err != nil {
		w.Write(ResponseBuf("031", "远程调用失败"))
		return 0
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		w.Write(ResponseBuf("031", "远程调用失败"))
		return 0
	}
	resModel := make(map[string]interface{})

	err = json.Unmarshal(buf, &resModel)
	if err != nil {
		w.Write(ResponseBuf("018", "Json反序列化出错"))
		return 0
	}

	IDStr, ok := resModel["data"]
	if !ok || IDStr == nil {
		w.Write(ResponseBuf("031", "远程调用失败"))
		return 0
	}
	ID, err := strconv.Atoi(IDStr.(string))
	if err != nil {
		w.Write(ResponseBuf("031", "远程调用失败"))
		return 0
	}
	// if fmt.Sprint(ID) != readID {
	// 	w.Write(ResponseBuf("032", "非法用户"))
	// 	return 0
	// }
	return uint(ID)
}

func SetPageQuery(w http.ResponseWriter, r *http.Request, odds map[string]interface{}, db *gorm.DB) (*gorm.DB, bool) {
	qu, ok := odds["pagequery"]
	if ok {
		qump := qu.(map[string]interface{})
		offset, ok := qump["offset"]
		if !ok {
			w.Write(ResponseBuf("019", "pagequery缺少offset字段"))
			return nil, true
		}
		pagesize, ok := qump["pagesize"]
		if !ok {
			w.Write(ResponseBuf("019", "pagequery缺少pagesize字段"))
			return nil, true
		}
		off, err := strconv.Atoi(fmt.Sprint(offset))
		if err != nil {
			w.Write(ResponseBuf("019", err.Error()))
			return nil, true
		}
		siz, err := strconv.Atoi(fmt.Sprint(pagesize))
		if err != nil {
			w.Write(ResponseBuf("019", err.Error()))
			return nil, true
		}
		db = db.Offset(off * siz).Limit(siz)
		return db, true
	}
	return db, false
}
func SetOrder(w http.ResponseWriter, r *http.Request, odds map[string]interface{}, db *gorm.DB) (*gorm.DB, bool) {
	qu, ok := odds["order"]
	if ok {
		switch qu.(type) {
		case string:
			db = db.Order(qu)
		default:
			w.Write(ResponseBuf("019", "order字段不是合法字符串"))
			return nil, true
		}
		return db, true
	}
	return db, false
}

func GetParam[V any](w http.ResponseWriter, r *http.Request, params map[string]interface{}, key string, param *V) bool {
	res, ok := params[key]
	if !ok {
		w.Write(ResponseBuf("019", "请求缺少："+key+"字段"))
		return false
	}
	val, ok := res.(V)
	if !ok {
		w.Write(ResponseBuf("019", key+"字段格式错误"))
		return false
	}
	(*param) = val
	return true
}
