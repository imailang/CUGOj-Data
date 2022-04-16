package main

import (
	contest "CUGOj-Data/src/Contest"
	judge "CUGOj-Data/src/Judge"
	problem "CUGOj-Data/src/Problem"
	properties "CUGOj-Data/src/Properties"
	sqltool "CUGOj-Data/src/SqlTool"
	user "CUGOj-Data/src/User"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

var routers = map[string]func(http.ResponseWriter, *http.Request){
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

	"/user/adduser":       user.AddUser,
	"/user/deleteuser":    user.DeleteUser,
	"/user/getuser":       user.GetUser,
	"/user/changeuser":    user.ChangeUser,
	"/user/getbaseuser":   user.GetBaseUser,
	"/user/checkusername": user.CheckUsername,
	"/user/checknickname": user.CheckNickname,
	"/user/checkemail":    user.CheckEmail,
	"/user/getuserlist":   user.GetUserList,
	"/user/getusercount":  user.GetUserCount,

	"/contest/addcontest":      contest.AddContest,
	"/contest/deletecontest":   contest.DeleteContest,
	"/contest/changecontest":   contest.ChangeContest,
	"/contest/getcontest":      contest.GetContest,
	"/contest/getbasecontest":  contest.GetBaseContest,
	"/contest/getcontestlist":  contest.GetContestList,
	"/contest/getcontestcount": contest.GetContestCount,
}

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
	for k, v := range routers {
		r.Post(k, v)
	}
	http.ListenAndServe("0.0.0.0:13000", r)
}
