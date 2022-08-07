package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// go run container.go run <cmd> <params>

func main () {
	if len(os.Args) < 2 {
        panic("few arguments")
    }

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command")
	}
}

// create new namespaces for the container,
// and re-envoke the process with the child command in the new namespaces.
func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	assert(cmd.Run())
}

func child() {
	fmt.Printf("[%v] %v start running\n", os.Getpid(), os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	assert(syscall.Sethostname([]byte("container")))
	assert(syscall.Chroot("./ubuntu-fs"))
	assert(os.Chdir("/"))
	assert(syscall.Mount("proc", "proc", "proc", 0, ""))

	assert(cmd.Run())

	assert(syscall.Unmount("/proc", 0))
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}