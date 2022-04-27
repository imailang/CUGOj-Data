package problem

import (
	filetool "CUGOj-Data/src/FileTool"
	httptool "CUGOj-Data/src/HttpTool"
	sqltool "CUGOj-Data/src/SqlTool"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
)

func DfsGetFiles(folder string, paths map[string]string) {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if file.IsDir() {
			DfsGetFiles(folder+"/"+file.Name(), paths)
		} else {
			ext := path.Ext(file.Name())
			name := folder + "/" + file.Name()[:len(file.Name())-len(ext)]

			if ext == ".in" || ext == ".out" {
				if _, ok := paths[name]; ok {
					paths[name] = "legal"
				} else {
					paths[name] = ext[1:]
				}
			} else if filetool.CheckExt(ext) {
				paths[name+ext] = "comp"
			} else {
				paths[name+ext] = "unk"
			}
		}
	}
}

func RemovePre(paths map[string]string, pre string) map[string]string {
	res := make(map[string]string)
	length := len(pre)
	for k, v := range paths {
		res[k[length:]] = v
	}
	return res
}

type FileData struct {
	Path string
	Info string
}

func Files(Id string) []FileData {
	paths := make(map[string]string)
	DfsGetFiles(filetool.Home()+"data/problems/"+Id+"/files", paths)
	paths = RemovePre(paths, filetool.Home()+"data/problmes/"+Id+"/files/")
	res := make([]FileData, len(paths))
	pos := 0
	for k, v := range paths {
		res[pos] = FileData{
			Path: k,
			Info: v,
		}
		pos++
	}
	return res
}

func GetFiles(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", err.Error()))
		return
	}
	paths := Files(string(buf))
	buf = httptool.GetJson(w, r, &paths)
	if buf == nil {
		return
	}
	w.Write(httptool.SuccessBuf(string(buf)))

}

func UnComp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	args := make([]string, 2)
	if !httptool.UnMarshal(w, r, &args) {
		return
	}
	if len(args) != 2 {
		w.Write(httptool.ResponseBuf("019", "解压只接收两个参数(ID,path)"))
		return
	}
	err := filetool.UnComp(args[0], args[1], path.Ext(args[1]))
	if err != nil {
		w.Write(httptool.ResponseBuf("026", err.Error()))
		return
	}
	w.Write(httptool.SuccessBuf("解压成功"))
}

func GetCasesUsed(w http.ResponseWriter, r *http.Request) {
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
	res := sqltool.Problem{}
	result := db.Select("CaseFiles").Find(&res, string(buf))
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
	w.Write(httptool.SuccessBuf(res.CaseFiles))
}

func UploadFiles(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method == "OPTIONS" {
		return
	}
	err := r.ParseMultipartForm(16 * 1024 * 1024)
	if err != nil {
		w.Write(httptool.ResponseBuf("019", "请求格式错误"))
		return
	}
	id := r.PostForm.Get("ID")
	if id == "" {
		w.Write(httptool.ResponseBuf("019", "请求缺少ID字段"))
		return
	}
	files := r.MultipartForm.File["file"]
	successCnt := 0
	errorInfo := ""
	for _, file := range files {
		fs, err := file.Open()
		if err != nil {
			errorInfo += err.Error() + "\n"
			continue
		}
		defer fs.Close()
		local, err := os.Create(filetool.Home() + "data/problems/" + fmt.Sprint(id) + "/files/" + file.Filename)
		if err != nil {
			errorInfo += err.Error() + "\n"
			continue
		}
		defer local.Close()

		io.Copy(local, fs)
		successCnt++
	}
	if successCnt == len(files) {
		w.Write(httptool.SuccessBuf("上传成功"))
	} else {
		w.Write(httptool.ResponseBuf("027", "上传成功："+fmt.Sprint(successCnt)+"\n部分文件上传失败，错误信息：\n"+errorInfo))
	}
}

func RemoveFiles(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	model := make([]string, 1)
	if !httptool.UnMarshal(w, r, &model) {
		return
	}
	if len(model) == 0 {
		w.Write(httptool.ResponseBuf("019", "至少传递一个参数(ID)"))
	}
	pre := filetool.Home() + "data/problems/" + model[0] + "/files/"
	errorInfo := ""
	successSize := 0
	for _, str := range model[1:] {
		path := pre + str
		err := os.Remove(path)
		if err != nil {
			errorInfo += err.Error() + "\n"
			continue
		}
		successSize++
	}
	if successSize == len(model)-1 {
		w.Write(httptool.SuccessBuf("删除成功"))
	} else {
		w.Write(httptool.ResponseBuf("028", "部分文件删除失败，成功数："+fmt.Sprint(successSize)+"\n错误信息："+errorInfo))
	}
}

type caseData struct {
	ID   string
	Path string
}

func SetCases(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	model := make([]string, 1)
	if !httptool.UnMarshal(w, r, &model) {
		return
	}
	if len(model) == 0 {
		w.Write(httptool.ResponseBuf("019", "至少传递一个参数(ID)"))
	}
	files := model[1:]
	sort.Strings(files)
	res := make([]caseData, len(files))
	for i, str := range files {
		res[i] = caseData{
			ID:   fmt.Sprint(i + 1),
			Path: str,
		}
	}
	buf, err := json.Marshal(&res)
	if err != nil {
		w.Write(httptool.ResponseBuf("005", err.Error()))
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	result := db.Model(&sqltool.Problem{}).Where("id=?", model[0]).Update("case_files", string(buf))
	if !httptool.DisposeAddResult(w, r, result) {
		return
	}
	w.Write(httptool.SuccessBuf("修改测试用例成功"))
}
