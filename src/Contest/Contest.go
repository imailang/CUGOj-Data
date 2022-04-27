package contest

import (
	httptool "CUGOj-Data/src/HttpTool"
	sqltool "CUGOj-Data/src/SqlTool"
	team "CUGOj-Data/src/Team"
	user "CUGOj-Data/src/User"
	"fmt"
	"net/http"
	"time"
)

func GetContest(w http.ResponseWriter, r *http.Request) {
	httptool.Query(&[]sqltool.Contest{}, w, r, "password")
}
func AddContest(w http.ResponseWriter, r *http.Request) {
	httptool.Add(&sqltool.Contest{}, w, r)
}
func ChangeContest(w http.ResponseWriter, r *http.Request) {
	httptool.Change(&[]sqltool.Contest{}, w, r)
}
func DeleteContest(w http.ResponseWriter, r *http.Request) {
	httptool.Delete(&[]sqltool.Contest{}, w, r)
}
func GetContestList(w http.ResponseWriter, r *http.Request) {
	httptool.List(&[]sqltool.Contest{}, w, r, "password")
}
func GetContestCount(w http.ResponseWriter, r *http.Request) {
	httptool.Count(&[]sqltool.Contest{}, w, r)
}
func GetBaseContest(w http.ResponseWriter, r *http.Request) {
	httptool.Base(&sqltool.Contest{
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(5 * time.Hour),
		SealRankTime: time.Now().Add(4 * time.Hour),
	}, w, r)
}

type UserContestData struct {
	ID            uint
	Title         string
	StartTime     time.Time
	EndTime       time.Time
	Owner         uint
	RegisterCount int
	Type          byte
	Permission    byte
	RegisteMode   byte
	Visible       byte
}

func GetUserContestList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ID := httptool.GetIDFromTokenNotWrite(r)

	odds := make(map[string]interface{})
	if !httptool.UnMarshal(w, r, &odds) {
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	model := sqltool.Contest{}
	db = db.Model(&model)
	whereString := ""
	length := len(odds)
	db, ok := httptool.SetPageQuery(w, r, odds, db)
	if db == nil {
		return
	}
	if ok {
		length--
	}
	db, ok = httptool.SetOrder(w, r, odds, db)
	if db == nil {
		return
	}
	if ok {
		length--
	}
	args := make([]interface{}, length)
	pos := 0
	for k, v := range odds {
		if k == "pagequery" || k == "order" {
			continue
		}
		if whereString != "" {
			whereString += " and "
		}
		whereString += k + " ?"
		args[pos] = v
		pos++
	}
	db = db.Where(whereString, args...)
	if ID != 0 {
		if user.CheckAdminByID(ID) {
			db = db.Where("owner in ?", []uint{0, 1, ID})
		} else {
			db = db.Where("owner in ?", []uint{0, ID})
		}
	} else {
		db = db.Where("owner = ?", 0)
	}
	res := []sqltool.Contest{}
	result := db.Find(&res)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	CID := make([]uint, result.RowsAffected)
	UserContestDatas := make([]UserContestData, result.RowsAffected)
	for i := 0; i < int(result.RowsAffected); i++ {
		CID[i] = res[i].ID
		UserContestDatas[i].ID = res[i].ID
		UserContestDatas[i].StartTime = res[i].StartTime
		UserContestDatas[i].EndTime = res[i].EndTime
		UserContestDatas[i].Owner = res[i].Owner
		UserContestDatas[i].Title = res[i].Title
		UserContestDatas[i].Type = res[i].Type
		UserContestDatas[i].Permission = 0
		UserContestDatas[i].RegisteMode = res[i].RegisteMode
		UserContestDatas[i].Visible = res[i].Visible
	}
	resmodel := sqltool.ContestRegister{}

	db = httptool.GetDB(w, r)
	if db == nil {
		return
	}
	// subQuery1 := db.Model(&resmodel).Select("c_id", "count(*) as cnt").Where("c_id in ?", CID).Where("t_id = ?", 0).Group("c_id")
	// subQuery2 := db.Model(&resmodel).Where("c_id in ?", CID).Where("t_id <> ?", 0).Distinct("t_id", "c_id")
	// subQuery3 := db.Model(&resmodel).Select("c_id", "count(*) as cnt").Table("(?) as tc", subQuery2).Group("c_id")
	// // rows, err := db.Model(&resmodel).Select("c_id", "count(*)").Where("c_id in ?", CID).Where(db.Where(db.Where("t_id = ?", 0)).Or(db.Where("t_id <> ?", 0).Distinct("t_id", "c_id"))).Group("c_id").Rows()
	// rows, err := db.Select("u.c_id", "u.cnt+t.cnt").Table("(?) as u,(?) as t", subQuery1, subQuery3).Where("u.c_id = t.c_id").Rows()
	rows, err := db.Raw("SELECT `c_id`,count(*) as cnt FROM `contest_registers` WHERE c_id in ? AND t_id = 0 GROUP BY `c_id` UNION ALL SELECT `c_id`,count(*) as cnt FROM (SELECT DISTINCT `t_id`,`c_id` FROM `contest_registers` WHERE c_id in ? AND t_id <> 0) as tc GROUP BY `c_id` ORDER BY `c_id`", CID, CID).Rows()
	if err != nil {
		w.Write(httptool.ResponseBuf("022", "数据库查询出错"))
		return
	}
	for rows.Next() {
		key := ""
		count := 0
		err := rows.Scan(&key, &count)

		if err != nil {
			w.Write(httptool.ResponseBuf("022", "数据库查询出错"))
			return
		}
		for i, data := range UserContestDatas {
			if fmt.Sprint(data.ID) == key {
				UserContestDatas[i].RegisterCount = count
				break
			}
		}
	}

	if ID != 0 {
		db = httptool.GetDB(w, r)
		if db == nil {
			return
		}
		for i := range UserContestDatas {
			UserContestDatas[i].Permission = 2
		}

		rows, err = db.Model(&resmodel).Select("c_id", "count(*)").Where("c_id in ?", CID).Where("uid = ?", ID).Group("c_id").Rows()
		if err != nil {
			w.Write(httptool.ResponseBuf("022", "数据库查询出错"))
			return
		}
		for rows.Next() {
			key := ""
			count := 0
			err := rows.Scan(&key, &count)
			if err != nil {
				w.Write(httptool.ResponseBuf("022", "数据库查询出错"))
				return
			}
			for i, data := range UserContestDatas {
				if fmt.Sprint(data.ID) == key {
					if count != 0 {
						UserContestDatas[i].Permission = 1
					} else {
						UserContestDatas[i].Permission = 2
					}
					break

				}
			}

		}

	}

	buf := httptool.GetJson(w, r, &UserContestDatas)
	if buf == nil {
		return
	}
	w.Write(httptool.SuccessBuf(string(buf)))
}

func RegisterContest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ID := httptool.GetIDFromToken(w, r)
	if ID == 0 {
		return
	}
	params := make(map[string]interface{})
	if !httptool.UnMarshal(w, r, &params) {
		return
	}
	var cid uint
	var tmp float64
	if !httptool.GetParam(w, r, params, "cid", &tmp) {
		return
	}
	cid = uint(tmp)
	var tid uint
	if !httptool.GetParam(w, r, params, "tid", &tmp) {
		return
	}
	tid = uint(tmp)
	var mode uint
	if !httptool.GetParam(w, r, params, "mode", &tmp) {
		return
	}
	mode = uint(tmp)
	var password string
	if !httptool.GetParam(w, r, params, "password", &password) {
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	contest := sqltool.Contest{}
	result := db.Find(&contest, cid)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	if contest.Owner > 1 && contest.Owner != ID {
		w.Write(httptool.ResponseBuf("032", "无权限参赛"))
		return
	} else if contest.Owner == 1 && !user.CheckAdminByID(ID) {
		w.Write(httptool.ResponseBuf("032", "无权限参赛"))
		return
	}
	if contest.Visible == 1 {
		w.Write(httptool.ResponseBuf("032", "无权限参赛"))
		return
	} else if contest.Visible >= 2 {
		if contest.Password != password {
			w.Write(httptool.ResponseBuf("033", "密码错误"))
			return
		}
	}
	register := sqltool.ContestRegister{}
	db = httptool.GetDB(w, r)
	if db == nil {
		return
	}
	result = db.Where("c_id = ? and uid = ?", cid, ID).Find(&register)
	if result.Error != nil {
		w.Write(httptool.ResponseBuf("022", "数据库查询出错"))
		return
	} else if result.RowsAffected != 0 {
		w.Write(httptool.ResponseBuf("025", "已报名"))
		return
	}

	db = httptool.GetDB(w, r)
	if db == nil {
		return
	}
	if contest.RegisteMode == 0 {
		result = db.Create(&sqltool.ContestRegister{
			CID: cid,
			UID: ID,
			TID: 0,
		})
		if !httptool.DisposeAddResult(w, r, result) {
			return
		}
	} else if contest.RegisteMode == 1 {
		members := team.GetMembers(tid)
		if members == nil {
			w.Write(httptool.ResponseBuf("032", "请使用本人所在队伍报名"))
			return
		}
		registers := make([]sqltool.ContestRegister, len(members))
		for i, uid := range members {
			registers[i] = sqltool.ContestRegister{
				CID: cid,
				UID: uid,
				TID: tid,
			}
		}
		result = db.Create(&registers)
		if !httptool.DisposeAddResult(w, r, result) {
			return
		}
	} else {
		if mode == 1 {
			result = db.Create(&sqltool.ContestRegister{
				CID: cid,
				UID: ID,
				TID: 0,
			})
			if !httptool.DisposeAddResult(w, r, result) {
				return
			}
		} else if mode == 2 {
			members := team.GetMembers(tid)
			if members == nil {
				w.Write(httptool.ResponseBuf("032", "请使用本人所在队伍报名"))
				return
			}
			registers := make([]sqltool.ContestRegister, len(members))
			for i, uid := range members {
				registers[i] = sqltool.ContestRegister{
					CID: cid,
					UID: uid,
					TID: tid,
				}
			}
			result = db.Create(&registers)
			if !httptool.DisposeAddResult(w, r, result) {
				return
			}
		} else {
			w.Write(httptool.ResponseBuf("019", "mode字段值存在问题，仅允许'1'或'2'"))
		}
	}
	w.Write(httptool.SuccessBuf("注册成功"))
}

func GetUserContestCount(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ID := httptool.GetIDFromTokenNotWrite(r)

	odds := make(map[string]interface{})
	if !httptool.UnMarshal(w, r, &odds) {
		return
	}
	db := httptool.GetDB(w, r)
	if db == nil {
		return
	}
	model := sqltool.Contest{}
	db = db.Model(&model)
	whereString := ""
	length := len(odds)
	args := make([]interface{}, length)
	pos := 0
	for k, v := range odds {
		if k == "pagequery" || k == "order" {
			length--
			continue
		}
		if whereString != "" {
			whereString += " and "
		}
		whereString += k + " ?"
		args[pos] = v
		pos++
	}
	db = db.Where(whereString, args[:length]...)
	if ID != 0 {
		if user.CheckAdminByID(ID) {
			db = db.Where("owner in ?", []uint{0, 1, ID})
		} else {
			db = db.Where("owner in ?", []uint{0, ID})
		}
	} else {
		db = db.Where("owner = ?", 0)
	}
	var count int64
	result := db.Count(&count)
	if !httptool.DisposeQueryResult(w, r, result) {
		return
	}
	w.Write(httptool.SuccessBuf(fmt.Sprint(count)))
}
