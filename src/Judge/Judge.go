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
	if !httptool.UnMarshal(w, r, &res) {
		return
	}
	res.SubmitTime = time.Now()
	res.Status = "Pending"
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	problem := sqltool.Problem{}
	result := db.Find(&problem, res.PID)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	res.PShowID = problem.ShowID
	res.PTitle = problem.Title
	res.Length = len(res.Code)

	result = db.Create(&res)
	if !httptool.DisposeAddResult(w, r, result) {
		return
	}

	err := queuetool.NewWork("judge", fmt.Sprint(res.ID))
	if err != nil {
		buf, _ := json.Marshal(httptool.Response{
			Statu: "023",
			Info:  "评测队列发布出错",
		})
		w.Write(buf)
		db.Delete(&res)
		return
	}
	w.Write(httptool.SuccessBuf(res.GetID()))
}

func GetBaseJudge(w http.ResponseWriter, _ *http.Request) {
	res := sqltool.Judge{
		Status: "Pending",
	}
	buf := httptool.GetJson(w, nil, &res, "Problem", "UserInfo")
	if buf == nil {
		return
	}
	w.Write(buf)
}

func GetJudge(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "Http请求读取出错"))
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
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
	buf = httptool.GetJson(w, r, &res, "Problem")

	if buf == nil {
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
	buf = httptool.GetJson(w, r, &res, "Problem")

	if buf == nil {
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
	buf = httptool.GetJson(w, r, &res, "Judge", "JID")
	if buf == nil {
		return
	}
	w.Write(httptool.SuccessBuf(string(buf)))
}

func GetJudgeList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	model := []sqltool.Judge{}

	var qu interface{}

	odds := make(map[string]interface{})

	if !httptool.UnMarshal(w, r, &odds) {
		return
	}

	db := sqltool.GetDB()
	if db == nil {
		buf, _ := json.Marshal(&httptool.Response{
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
	result := db.Order("submit_time desc").Omit("Code").Find(&model)
	if result.Error != nil {
		buf, _ := json.Marshal(&httptool.Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}
	buf, err := json.Marshal(&model)
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
