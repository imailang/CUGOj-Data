package practice

import (
	httptool "CUGOj-Data/src/HttpTool"
	sqltool "CUGOj-Data/src/SqlTool"
	"net/http"
	"strings"
)

func CheckPermission(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ID := httptool.GetIDFromToken(w, r)
	if ID == 0 {
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}

	model := sqltool.UserInfo{}
	result := db.Model(&model).Select("id", "title").Where("id=?", ID).Find(&model)

	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}

	titles := strings.Split(model.Title, ";")
	for _, str := range titles {
		if strings.HasSuffix(str, "队员") {
			w.Write(httptool.SuccessBuf(model.GetID()))
			return
		}
	}
	w.Write(httptool.ResponseBuf("032", "无权限"))
}
