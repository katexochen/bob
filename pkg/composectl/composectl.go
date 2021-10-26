package composectl

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
)

var (
	ErrInvalidProject = fmt.Errorf("invalid project")
	ErrComposeError   = fmt.Errorf("compose error")
)

type ComposeController struct {
	project *types.Project
	service api.Service

	stdout pipe
	stderr pipe
	stdin  pipe

	logger *logger
}

type pipe struct {
	r *os.File
	w *os.File
}

func New(project *types.Project, conflicts, mappings string) (*ComposeController, error) {
	if project == nil || project.Name == "" {
		return nil, ErrInvalidProject
	}

	c := &ComposeController{
		project: project,
	}

	// create pipes for stdout, stderr and stdin
	var err error
	c.stdout.r, c.stdout.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	c.stderr.r, c.stderr.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	c.stdin.r, c.stdin.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	if conflicts != "" {
		_, err = c.stdout.w.Write([]byte(conflicts))
		if err != nil {
			return nil, err
		}
	}

	if mappings != "" {
		_, err = c.stdout.w.Write([]byte(mappings))
		if err != nil {
			return nil, err
		}
	}

	logger, err := NewLogger(c.stdout.w)
	if err != nil {
		return nil, err
	}
	c.logger = logger

	dockerCli, err := command.NewDockerCli(
		command.WithCombinedStreams(nil),
		command.WithOutputStream(nil),
		command.WithErrorStream(nil),
		command.WithInputStream(nil),
	)
	if err != nil {
		return nil, err
	}

	err = dockerCli.Initialize(flags.NewClientOptions())
	if err != nil {
		return nil, err
	}

	c.service = compose.NewComposeService(dockerCli.Client(), &configfile.ConfigFile{})

	return c, nil
}

func (ctl *ComposeController) Up(ctx context.Context) error {

	err := ctl.service.Up(ctx, ctl.project, api.UpOptions{
		//Start: api.StartOptions{
		//	Attach :      ctl.logger,
		//},
		//Create: api.CreateOptions{
		//	QuietPull:            true,
		//},
	})
	if err != nil {
		return err
	}

	go func() {
		err := ctl.service.Logs(ctx, ctl.project.Name, ctl.logger, api.LogOptions{
			Services:   nil,
			Tail:       "",
			Since:      "",
			Until:      "",
			Follow:     true,
			Timestamps: true,
		})
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func (ctl *ComposeController) Down(ctx context.Context) error {
	if ctl.project == nil {
		return ErrInvalidProject
	}

	err := ctl.service.Down(ctx, ctl.project.Name, api.DownOptions{
		Project: ctl.project,
	})
	if err != nil {
		return err
	}

	//if ctl.stderr.String() != "" {
	//	return ErrComposeError
	//}

	return nil
}

func (ctl *ComposeController) Stdout() io.Reader {
	return ctl.stdout.r
}

func (ctl *ComposeController) Stderr() io.Reader {
	return ctl.stderr.r
}

func (ctl *ComposeController) Stdin() io.Writer {
	return ctl.stdin.w
}
