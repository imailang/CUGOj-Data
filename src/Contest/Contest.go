package contest

import (
	httptool "CUGOj-Data/src/HttpTool"
	sqltool "CUGOj-Data/src/SqlTool"
	"net/http"
	"time"
)

func GetContest(w http.ResponseWriter, r *http.Request) {
	httptool.Query(&[]sqltool.Contest{}, w, r)
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
	httptool.List(&[]sqltool.Contest{}, w, r)
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
