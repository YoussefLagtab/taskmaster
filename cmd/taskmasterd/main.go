package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type UnixHTTPServer struct {
	SocketFile string
	Chmod string
}

type TaskmasterConfig struct {
	Logfile string
	ChildLogDir string
}

type Program struct {
	// config
	Name string
	Command string
	Numprocs int
	Autostart bool
	Startsecs int
	Restart string
	Exitcodes []int
	Stopsignal string
	Stopwaitsecs int
	Stdout_logfile string
	Stderr_logfile string
	Env []string
	Workingdir string
	Umask string

	// status
	processes []*exec.Cmd
}

func main()  {
	// unixServer := &UnixHTTPServer{
	// 	SocketFile: "./var/run/taskmaster.sock",
	// 	Chmod: "0700",
	// }

	taskmasterConfig := &TaskmasterConfig{
		Logfile: "./var/log/taskmaster/taskmaster.log",
		ChildLogDir: "./var/log/taskmaster",
	}

	programs := []*Program{
		{
			Name: "yes",
			Command: "./bin/yes.sh",
			Numprocs: 1,
			Autostart: true,
			Startsecs: 3,
			Restart: "unexpected",
			Exitcodes: []int{1, 2},
			Stopsignal: "TERM",
			Stopwaitsecs: 10,
			Stdout_logfile: "AUTO",
			Stderr_logfile: "AUTO",
			Env: []string{ "TEST=HELLO_WORLD", },
			Workingdir: "",
			Umask: "",
		},
		// &Program{
		// 	Name: "yes",
		// 	Command: "/usr/bin/yes",
		// 	Numprocs: 1,
		// 	Autostart: true,
		// 	Startsecs: 3,
		// 	Restart: "unexpected",
		// 	Exitcodes: []int{1, 2},
		// 	Stopsignal: "TERM",
		// 	Stopwaitsecs: 10,
		// 	Stdout_logfile: "AUTO",
		// 	Stderr_logfile: "AUTO",
		// 	Env: map[string]string{ "TEST": "HELLO_WORLD" },
		// 	Workingdir: "",
		// 	Umask: "",
		// },
	}


	startPrograms(taskmasterConfig, programs, os.Environ())

}
func startPrograms(taskmasterConfig *TaskmasterConfig, programs []*Program, env []string)  {
	for _, p := range(programs) {
		if (!p.Autostart) {
			continue
		}
		childsEnv := append(env, p.Env...)
		for i := 0; i < p.Numprocs; i++ {
			// get cmd
			cmd := exec.CommandContext(context.Background(), p.Command);

			// set env
			cmd.Env = childsEnv

			// attach stdout
			logfile := p.Stdout_logfile
			if p.Stdout_logfile == "AUTO" {
				logfile = path.Join(
					taskmasterConfig.ChildLogDir,
					strings.Join(
						[]string{p.Name, strconv.Itoa(i), ".log"}, "_",
					),
				)
			}
			file,err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				log.Fatalf("Can't open %s\nerror: %v", logfile, err)
			}
			cmd.Stdout = file

			// attach stderr
			logfile = p.Stderr_logfile
			if p.Stderr_logfile == "AUTO" {
				logfile = path.Join(
					taskmasterConfig.ChildLogDir,
					strings.Join(
						[]string{p.Name, strconv.Itoa(i), ".error"}, "_",
					),
				)
			}
			file, err = os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				log.Fatalf("Can't open %s\nerror: %v", logfile, err)
			}
			cmd.Stderr = file

			p.processes =  append(p.processes, cmd)
		}
	}

	for _, p := range(programs) {
		for _, cmd := range(p.processes) {
			cmd.Start()
		}
	}
}