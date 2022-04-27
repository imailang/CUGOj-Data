package problem

import (
	filetool "CUGOj-Data/src/FileTool"
	httptool "CUGOj-Data/src/HttpTool"
	sqltool "CUGOj-Data/src/SqlTool"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func ChangeProblem(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	problem := sqltool.Problem{}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "019",
			Info:  err.Error(),
		})
		w.Write(buf)
		return
	}
	err = json.Unmarshal(buf, &problem)
	if err != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "018",
			Info:  err.Error(),
		})
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
	result := db.Omit("ID").Updates(&problem.Description)
	if result.Error != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}
	problem.Description = sqltool.ProblemDescription{}
	result = db.Omit("description").Omit("d_id").Omit("id").Updates(&problem)
	if result.Error != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}

	result = db.Model(&sqltool.Judge{}).Where("p_id=?", problem.ID).Updates(map[string]interface{}{"PTitle": problem.Title, "PShowID": problem.ShowID})
	if result.Error != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}

	buf, _ = json.Marshal(&httptool.Response{
		Statu: "000",
		Info:  "修改成功",
	})
	w.Write(buf)
}

func AddProblem(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	problem := sqltool.Problem{}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "019",
			Info:  err.Error(),
		})
		w.Write(buf)
		return
	}
	err = json.Unmarshal(buf, &problem)
	if err != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "018",
			Info:  err.Error(),
		})
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
	if problem.Description == (sqltool.ProblemDescription{}) {
		problem.Description = sqltool.ProblemDescription{
			Examples: "[]",
		}
	}
	result := db.Create(&problem)
	if result.Error != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "021",
			Info:  result.Error.Error(),
		})
		w.Write(buf)
		return
	}
	os.Mkdir(filetool.Home()+"data/problems/"+fmt.Sprint(problem.ID), 0777)
	os.Mkdir(filetool.Home()+"data/problems/"+fmt.Sprint(problem.ID)+"/files", 0777)
	os.Mkdir(filetool.Home()+"data/problems/"+fmt.Sprint(problem.ID)+"/cases", 0777)
	buf, _ = json.Marshal(&httptool.Response{
		Statu: "000",
		Info:  fmt.Sprint(problem.ID),
	})
	w.Write(buf)
}

func GetProblem(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "019",
			Info:  err.Error(),
		})
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
	ID, err := strconv.Atoi(string(buf))
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "主键参数应为数字"))
		return
	}
	res := sqltool.Problem{}
	result := db.Preload("Description").Omit("CaseFiles").Find(&res, ID)
	if err = result.Error; err != nil {
		buf, _ = json.Marshal(&httptool.Response{
			Statu: "021",
			Info:  err.Error(),
		})
		w.Write(buf)
		return
	}
	if result.RowsAffected == 0 {
		w.Write(httptool.ResponseBuf("022", "查询的数据不存在"))
		return
	}
	buf, err = json.Marshal(&res)
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

func GetBaseProblem(w http.ResponseWriter, _ *http.Request) {
	res := sqltool.Problem{
		JudgeMode:   0,
		Title:       "请输入标题",
		TimeLimit:   1000,
		MemoryLimit: 256,
		StackLimit:  128,
		Description: sqltool.ProblemDescription{
			Markdown:    false,
			BackGround:  "题目背景",
			Description: "题目描述",
			Input:       "输入描述",
			Output:      "输出描述",
			Examples:    "[]",
			Hint:        "提示",
		},
		Source:         "CUG-ACM队",
		Owner:          0,
		CodeShare:      false,
		SpjLanguage:    "",
		CaseVersion:    1,
		OpenCaseResult: false,
		CaseFiles:      "[]",
	}
	buf, _ := json.Marshal(&res)
	w.Write(buf)
}

func GetProblemList(w http.ResponseWriter, r *http.Request) {
	httptool.List(&[]sqltool.Problem{}, w, r)
}

func GetProblemCount(w http.ResponseWriter, r *http.Request) {
	httptool.Count(&sqltool.Problem{}, w, r)
}

func DeleteProblem(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	model := sqltool.Problem{}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "Http请求读取出错"))
		return
	}

	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}

	path := filetool.Home() + "data/problems/" + string(buf)

	result := db.Delete(model, string(buf))

	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}

	os.RemoveAll(path)

	w.Write(httptool.SuccessBuf("删除成功"))
}
