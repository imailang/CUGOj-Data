package main

import (
	contest "CUGOj-Data/src/Contest"
	judge "CUGOj-Data/src/Judge"
	practice "CUGOj-Data/src/Practice"
	problem "CUGOj-Data/src/Problem"
	properties "CUGOj-Data/src/Properties"
	sqltool "CUGOj-Data/src/SqlTool"
	team "CUGOj-Data/src/Team"
	user "CUGOj-Data/src/User"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

var posts = map[string]func(http.ResponseWriter, *http.Request){
	"/problem/addproblem":      problem.AddProblem,
	"/problem/getproblem":      problem.GetProblem,
	"/problem/getbaseproblem":  problem.GetBaseProblem,
	"/problem/getproblemlist":  problem.GetProblemList,
	"/problem/changeproblem":   problem.ChangeProblem,
	"/problem/getproblemcount": problem.GetProblemCount,
	"/problem/deleteproblem":   problem.DeleteProblem,

	"/judge/addjudge":      judge.AddJudge,
	"/judge/getbasejudge":  judge.GetBaseJudge,
	"/judge/getjudge":      judge.GetJudge,
	"/judge/getjudgefull":  judge.GetJudgeFull,
	"/judge/getjudgecases": judge.GetJudgeCases,
	"/judge/getjudgelist":  judge.GetJudgeList,
	"/judge/getjudgecount": judge.GetJudgeCount,
	"/judge/deletejudge":   judge.DeleteJudge,

	"/user/adduser":           user.AddUser,
	"/user/deleteuser":        user.DeleteUser,
	"/user/getuser":           user.GetUser,
	"/user/changeuser":        user.ChangeUser,
	"/user/getbaseuser":       user.GetBaseUser,
	"/user/checkusername":     user.CheckUsername,
	"/user/checknickname":     user.CheckNickname,
	"/user/checkemail":        user.CheckEmail,
	"/user/getuserlist":       user.GetUserList,
	"/user/getusercount":      user.GetUserCount,
	"/user/getuserbyusername": user.GetUserByUsername,
	"/user/checkpassword":     user.CheckPassword,
	"/user/getsubmitcount":    user.GetSubmmitCount,

	"/contest/addcontest":          contest.AddContest,
	"/contest/deletecontest":       contest.DeleteContest,
	"/contest/changecontest":       contest.ChangeContest,
	"/contest/getcontest":          contest.GetContest,
	"/contest/getbasecontest":      contest.GetBaseContest,
	"/contest/getcontestlist":      contest.GetContestList,
	"/contest/getcontestcount":     contest.GetContestCount,
	"/contest/getusercontestlist":  contest.GetUserContestList,
	"/contest/getusercontestcount": contest.GetUserContestCount,
	"/contest/registercontest":     contest.RegisterContest,

	"/cases/getfiles":    problem.GetFiles,
	"/cases/getcaseused": problem.GetCasesUsed,
	"/cases/uncompress":  problem.UnComp,
	"/cases/uploadfiles": problem.UploadFiles,
	"/cases/removefiles": problem.RemoveFiles,
	"/cases/setcases":    problem.SetCases,

	"/practice/checkpermission": practice.CheckPermission,

	"/team/getteamlist":     team.GetTeamList,
	"/team/getselfteamlist": team.GetSelfTeamList,
}

// var options = map[string]func(http.ResponseWriter, *http.Request){
// 	"/cases/uploadfiles": problem.UploadFiles,
// }

func main() {
	err := properties.LoadProperties()
	if err != nil {
		fmt.Println(err)
	}
	err = sqltool.InitialSql()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("初始化结束")

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	for k, v := range posts {
		r.Post(k, v)
	}
	http.ListenAndServe("0.0.0.0:13000", r)
}
