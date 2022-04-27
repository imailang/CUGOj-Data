package team

import (
	httptool "CUGOj-Data/src/HttpTool"
	sqltool "CUGOj-Data/src/SqlTool"
	"net/http"
)

func GetTeamList(w http.ResponseWriter, r *http.Request) {
	httptool.List(&sqltool.Team{}, w, r)
}

func GetSelfTeamList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ID := httptool.GetIDFromToken(w, r)
	if ID == 0 {
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	res := []sqltool.Team{}
	result := db.Model(&sqltool.Team{}).Where("id in (?)", db.Model(&sqltool.TeamMember{}).Select("t_id").Where("uid = ?", ID)).Find(&res)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	buf := httptool.GetJson(w, r, &res)
	if buf == nil {
		return
	}
	w.Write(httptool.SuccessBuf(string(buf)))
}

func GetMembers(TID uint) []uint {
	db := sqltool.GetDB()
	if db == nil {
		return nil
	}
	model := []sqltool.TeamMember{}
	result := db.Where("t_id = ?", TID).Find(&model)
	if result.Error != nil {
		return nil
	}
	res := make([]uint, len(model))
	for i, item := range model {
		res[i] = item.UID
	}
	return res
}
