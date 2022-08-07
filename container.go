package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const numProcessesLimit = "20"

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
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |syscall.CLONE_NEWUSER,
		Unshareflags: syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID: 1000,
			Size: 1,
		}},
	}

	assert(cmd.Run())
}

// run the container with random host name.
// the root of the container is set to ubuntu-fs.
func child() {
	fmt.Printf("[%v] %v start running\n", os.Getpid(), os.Args[2:])
	
	cg()
	
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	assert(syscall.Sethostname([]byte(randomNameGenerator())))
	assert(syscall.Chroot("./ubuntu-fs"))
	assert(os.Chdir("/"))
	assert(syscall.Mount("proc", "proc", "proc", 0, ""))

	assert(cmd.Run())

	assert(syscall.Unmount("/proc", 0))
}

// create control group, set max number of processes to numProcessesLimit,
// and add container process to this control group.
func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := path.Join(cgroups, "pids")
	os.Mkdir(pids, 0755)
	assert(os.WriteFile(path.Join(pids, "pids.max"), []byte(numProcessesLimit), 0700))
	assert(os.WriteFile(path.Join(pids, "notify_on_release"), []byte("1"), 0700))
	assert(os.WriteFile(path.Join(pids, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}


func assert(err error) {
	if err != nil {
		panic(err)
	}
}