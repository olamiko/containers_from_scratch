package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// docker         run image <cmd> <params>
// go run main.go run       <cmd> <params>

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command")
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	err := os.Mkdir(filepath.Join(pids, "ola"), 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	checkError(ioutil.WriteFile(filepath.Join(pids, "ola/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	checkError(ioutil.WriteFile(filepath.Join(pids, "ola/notify_on_release"), []byte("1"), 0700))
	checkError(ioutil.WriteFile(filepath.Join(pids, "ola/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func run() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	cmd.Run()
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cg()
	syscall.Sethostname([]byte("container"))

	err := syscall.Chroot("linux-fs")
	checkError(err)

	err = syscall.Chdir("linux-fs")
	checkError(err)

	err = syscall.Mount("proc", "proc", "proc", 0, "")
	checkError(err)

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	syscall.Unmount("/proc", 0)

}
