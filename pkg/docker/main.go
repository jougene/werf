package docker

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"

	"github.com/flant/logboek"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	cli       *command.DockerCli
	apiClient *client.Client
)

func Init(dockerConfigDir string) error {
	if dockerConfigDir != "" {
		cliconfig.SetDir(dockerConfigDir)
	}

	if err := os.Setenv("DOCKER_CONFIG", dockerConfigDir); err != nil {
		return fmt.Errorf("cannot set DOCKER_CONFIG to %s: %s", dockerConfigDir, err)
	}

	if err := setDockerClient(); err != nil {
		return err
	}

	if err := setDockerApiClient(); err != nil {
		return err
	}

	return nil
}

func ServerVersion() (*types.Version, error) {
	ctx := context.Background()
	version, err := apiClient.ServerVersion(ctx)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

type customReadCloser struct {
	*os.File
}

// ignore closing stdin after interactive mode, e.g. docker run -ti
func (r customReadCloser) Close() error {
	return nil
}

func setDockerClient() error {
	cliOpts := []command.DockerCliOption{
		command.WithInputStream(customReadCloser{os.Stdin}),
		command.WithOutputStream(logboek.GetOutStream()),
		command.WithErrorStream(logboek.GetErrStream()),
		command.WithContentTrust(false),
	}

	newCli, err := command.NewDockerCli(cliOpts...)
	if err != nil {
		return err
	}

	newCli.Out().SetIsTerminal(terminal.IsTerminal(int(os.Stdout.Fd())))
	newCli.In().SetIsTerminal(terminal.IsTerminal(int(os.Stdin.Fd())))

	opts := flags.NewClientOptions()
	if err := newCli.Initialize(opts); err != nil {
		return err
	}

	cli = newCli

	return nil
}

func setDockerApiClient() error {
	ctx := context.Background()
	serverVersion, err := cli.Client().ServerVersion(ctx)
	apiClient, err = client.NewClientWithOpts(client.WithVersion(serverVersion.APIVersion))
	if err != nil {
		return err
	}

	return nil
}

func Debug() bool {
	return os.Getenv("WERF_DEBUG_DOCKER") == "1"
}
