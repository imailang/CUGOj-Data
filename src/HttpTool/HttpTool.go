package httptool

import (
	sqltool "CUGOj-Data/src/SqlTool"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

type Response struct {
	Statu string
	Info  string
}

func UnMarshal(r *http.Request, model any) ([]byte, error) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		buf, _ = json.Marshal(&Response{
			Statu: "019",
			Info:  err.Error(),
		})
		return buf, err
	}
	err = json.Unmarshal(buf, model)
	if err != nil {
		buf, _ = json.Marshal(&Response{
			Statu: "018",
			Info:  err.Error(),
		})
	}
	return buf, err
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

func GetDB() (*gorm.DB, error) {
	db := sqltool.GetDB()
	if db == nil {
		buf, _ := json.Marshal(&Response{
			Statu: "020",
			Info:  "数据库链接失效",
		})
		return nil, Error{Info: string(buf)}
	}
	return db, nil
}

func DisposeAddResult(result *gorm.DB) []byte {
	if result.Error != nil {
		buf, _ := json.Marshal(&Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		return buf
	}
	return nil
}

func DisposeQueryResult(result *gorm.DB) []byte {
	if result.Error != nil {
		buf, _ := json.Marshal(&Response{
			Statu: "022",
			Info:  result.Error.Error(),
		})
		return buf
	}
	return nil
}

func SuccessBuf(info string) []byte {
	buf, _ := json.Marshal(&Response{
		Statu: "000",
		Info:  info,
	})
	return buf
}

func GetJson(model any, ignores ...string) ([]byte, error) {
	buf, err := json.Marshal(model)
	if err != nil {
		return buf, err
	}
	tmp := make(map[string]interface{})
	err = json.Unmarshal(buf, &tmp)
	if err != nil {
		return buf, nil
	}
	for _, key := range ignores {
		delete(tmp, key)
	}
	buf, err = json.Marshal(&tmp)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func Add(model sqltool.SqlModel, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := UnMarshal(r, model)
	if err != nil {
		w.Write(buf)
		return
	}

	db, err := GetDB()
	if err != nil {
		w.Write(DbError())
		return
	}

	result := db.Create(model)
	buf = DisposeAddResult(result)
	if buf != nil {
		w.Write(buf)
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

	db, err := GetDB()
	if err != nil {
		w.Write(DbError())
		return
	}
	for _, str := range omi {
		db = db.Omit(str)
	}

	result := db.Find(model, string(buf))

	buf = DisposeQueryResult(result)
	if buf != nil {
		w.Write(buf)
		return
	}
	if result.RowsAffected == 0 {
		w.Write(ResponseBuf("022", "查询的数据不存在"))
		return
	}

	buf, err = GetJson(model)
	if err != nil {
		w.Write(ResponseBuf("005", "Json序列化出错"))
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

	db, err := GetDB()
	if err != nil {
		w.Write(DbError())
		return
	}

	result := db.Delete(model, string(buf))

	buf = DisposeQueryResult(result)
	if buf != nil {
		w.Write(buf)
		return
	}
	w.Write(SuccessBuf("删除成功"))
}

func Change(model any, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := UnMarshal(r, model)
	if err != nil {
		w.Write(buf)
		return
	}

	db, err := GetDB()
	if err != nil {
		w.Write(DbError())
		return
	}

	result := db.Updates(model)
	buf = DisposeAddResult(result)
	if buf != nil {
		w.Write(buf)
		return
	}
	w.Write(SuccessBuf("修改成功"))
}

func Base(model any, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := GetJson(model)
	if err != nil {
		w.Write(ResponseBuf("005", "序列化出现错误"))
		return
	}
	w.Write(buf)
}

func Count(model any, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	odds := make(map[string]interface{})

	buf, err := UnMarshal(r, &odds)
	if err != nil {
		w.Write(buf)
		return
	}

	db := sqltool.GetDB()
	if db == nil {
		buf, _ = json.Marshal(&Response{
			Statu: "020",
			Info:  "数据库链接失效",
		})
		w.Write(buf)
		return
	}

	db = db.Model(model)

	for _, odd := range odds {
		db = db.Where(odd)
		if db.Error != nil {
			w.Write(ResponseBuf("022", db.Error.Error()))
			return
		}
	}

	var cnt int64

	result := db.Count(&cnt)
	if result.Error != nil {
		buf, _ = json.Marshal(&Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}
	w.Write(SuccessBuf(fmt.Sprint(cnt)))
}

func List(model any, w http.ResponseWriter, r *http.Request, omi ...string) {
	defer r.Body.Close()

	var qu interface{}

	odds := make(map[string]interface{})

	buf, err := UnMarshal(r, &odds)
	if err != nil {
		w.Write(buf)
		return
	}

	db := sqltool.GetDB()
	if db == nil {
		buf, _ = json.Marshal(&Response{
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
			db = db.Where(odd)
		} else {
			db = db.Or(odd)
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
		buf, _ = json.Marshal(&Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}
	buf, err = json.Marshal(model)
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
