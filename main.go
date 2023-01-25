package main

import (
	"crontask/tasker"
	"fmt"
	"log"
	"os"
	"strconv"
)

func mustFindEnv(name string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		panic(fmt.Errorf("env:%s not found", name))
	}
	return v
}

func mustGetCronExpression() string {
	v := mustFindEnv(keyCronTaskExpression)
	if len(v) == 0 {
		panic(fmt.Errorf("nil env:%s", keyCronTaskExpression))
	}
	return v
}

func mustReadUserGroup() (uint64, uint64) {
	suid := mustFindEnv(keyUserID)
	sgid := mustFindEnv(keyGroupID)
	uid, uidErr := strconv.ParseUint(suid, 10, 64)
	if uidErr != nil {
		panic(uidErr)
	}
	gid, gidErr := strconv.ParseUint(sgid, 10, 64)
	if gidErr != nil {
		panic(gidErr)
	}
	return uid, gid
}

func mustGetCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}

func isEnableUserGroupSpec() bool {
	_, ok := os.LookupEnv(keyEnableUserGroupSpec)
	return ok
}

func main() {
	if len(os.Args) <= 1 {
		panic(fmt.Errorf("no exec program found, exp: crontask ${program} [${arg1}, ${arg2}, ...]"))
	}
	expression := mustGetCronExpression()
	var uid, gid uint64 = 0, 0
	enableUserGroupSpec := false
	if isEnableUserGroupSpec() {
		enableUserGroupSpec = true
		uid, gid = mustReadUserGroup()
	}
	cwd := mustGetCWD()
	program := os.Args[1]
	args := os.Args[2:]

	log.Printf("tasker init, expression:%s, cred:[enable:%t, uid:%d, gid:%d], cwd:%s, program:%s, args:[%+v]", expression, enableUserGroupSpec, uid, gid, cwd, program, args)

	opts := []tasker.Option{
		tasker.WithCronExpression(expression),
		tasker.WithWorkDir(cwd),
		tasker.WithProgram(program, args),
	}
	if enableUserGroupSpec {
		opts = append(opts, tasker.WithUserGroup(uid, gid))
	}
	tk, err := tasker.NewTasker(opts...)
	if err != nil {
		panic(err)
	}
	if err := tk.Run(); err != nil {
		panic(err)
	}
}
