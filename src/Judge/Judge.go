package judge

import (
	httptool "CUGOj-Data/src/HttpTool"
	queuetool "CUGOj-Data/src/QueueTool"
	sqltool "CUGOj-Data/src/SqlTool"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func AddJudge(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	res := sqltool.Judge{}
	buf, err := httptool.UnMarshal(r, &res)
	if err != nil {
		w.Write(buf)
		return
	}
	res.SubmitTime = time.Now()
	res.Status = "Pending"
	db, err := httptool.GetDB()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	problem := sqltool.Problem{}
	result := db.Find(&problem, res.PID)
	buf = httptool.DisposeQueryResult(result)
	if buf != nil {
		w.Write(buf)
		return
	}
	res.PShowID = problem.ShowID
	res.PTitle = problem.Title
	res.Length = len(res.Code)

	result = db.Create(&res)
	buf = httptool.DisposeAddResult(result)
	if buf != nil {
		w.Write(buf)
		return
	}

	err = queuetool.NewWork("judge", fmt.Sprint(res.ID))
	if err != nil {
		buf, _ = json.Marshal(httptool.Response{
			Statu: "023",
			Info:  "评测队列发布出错",
		})
		w.Write(buf)
		db.Delete(&res)
		return
	}
	w.Write(httptool.SuccessBuf("提交成功"))
}

func GetBaseJudge(w http.ResponseWriter, _ *http.Request) {
	res := sqltool.Judge{
		Status: "Pending",
	}
	buf, _ := httptool.GetJson(&res, "Problem", "UserInfo")
	w.Write(buf)
}

func GetJudge(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "Http请求读取出错"))
		return
	}
	db, err := httptool.GetDB()
	if err != nil {
		w.Write(httptool.DbError())
		return
	}
	res := sqltool.Judge{}

	result := db.Preload("UserInfo").Omit("Problem", "Code").Find(&res, string(buf))
	if result.Error != nil {
		w.Write(httptool.ResponseBuf("022", "数据库查询失败:"+result.Error.Error()))
		return
	}
	if result.RowsAffected == 0 {
		w.Write(httptool.ResponseBuf("022", "查询的数据不存在"))
		return
	}
	buf, err = httptool.GetJson(&res, "Problem")

	if err != nil {
		w.Write(httptool.ResponseBuf("005", "Json序列化错误"))
		return
	}
	w.Write(httptool.SuccessBuf(string(buf)))
}

func GetJudgeFull(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "Http请求读取出错"))
		return
	}
	res, err := sqltool.QueryJudge(string(buf))
	if err != nil {
		w.Write(httptool.ResponseBuf("022", "数据库查询失败:"+err.Error()))
		return
	}
	buf, err = httptool.GetJson(&res, "Problem")

	if err != nil {
		w.Write(httptool.ResponseBuf("005", "Json序列化错误"))
		return
	}
	w.Write(httptool.SuccessBuf(string(buf)))
}

func GetJudgeCases(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "Http请求读取出错"))
		return
	}
	res, err := sqltool.QueryJudgeCases(string(buf))
	if err != nil {
		w.Write(httptool.ResponseBuf("022", "数据库查询出错"))
		return
	}
	buf, err = httptool.GetJson(&res, "Judge", "JID")
	if err != nil {
		w.Write(httptool.ResponseBuf("005", "Json序列化发生问题"))
		return
	}
	w.Write(httptool.SuccessBuf(string(buf)))
}

func GetJudgeList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	model := []sqltool.Judge{}

	var qu interface{}

	odds := make(map[string]interface{})

	buf, err := httptool.UnMarshal(r, &odds)
	if err != nil {
		w.Write(buf)
		return
	}

	db := sqltool.GetDB()
	if db == nil {
		buf, _ = json.Marshal(&httptool.Response{
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
			w.Write(httptool.ResponseBuf("022", db.Error.Error()))
			return
		}
		oddCnt++
	}

	qu, ok := odds["pagequery"]
	if ok {
		qump := qu.(map[string]interface{})
		offset, ok := qump["offset"]
		if !ok {
			w.Write(httptool.ResponseBuf("019", "pagequery缺少offset字段"))
			return
		}
		pagesize, ok := qump["pagesize"]
		if !ok {
			w.Write(httptool.ResponseBuf("019", "pagequery缺少pagesize字段"))
			return
		}
		off, err := strconv.Atoi(fmt.Sprint(offset))
		if err != nil {
			w.Write(httptool.ResponseBuf("019", err.Error()))
			return
		}
		siz, err := strconv.Atoi(fmt.Sprint(pagesize))
		if err != nil {
			w.Write(httptool.ResponseBuf("019", err.Error()))
			return
		}
		db = db.Offset(off * siz).Limit(siz)
	}
	db = db.Preload("Problem", func(_db *gorm.DB) *gorm.DB {
		return _db.Select("ID", "Title", "ShowID")
	})
	db = db.Preload("UserInfo", func(_db *gorm.DB) *gorm.DB {
		return _db.Select("ID", "Nickname", "Username", "Realname")
	})
	result := db.Omit("Code").Find(&model)
	if result.Error != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}
	buf, err = json.Marshal(&model)
	if err != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "005",
			Info:  err.Error(),
		})
		w.Write(buf)
		return
	}
	w.Write(httptool.SuccessBuf(string(buf)))
}

func GetJudgeCount(w http.ResponseWriter, r *http.Request) {
	httptool.Count(&sqltool.Judge{}, w, r)
}
func DeleteJudge(w http.ResponseWriter, r *http.Request) {
	httptool.Delete(&sqltool.Judge{}, w, r)
}
