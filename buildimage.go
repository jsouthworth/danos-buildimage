package buildimage

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"
)

type Builder struct {
	cli     *client.Client
	ctx     context.Context
	version string
	destDir string
	srcDir  string
	pkgDir  string

	containerID string
}

type MakeBuilderOption func(*Builder)

func WithClient(cli *client.Client) MakeBuilderOption {
	return func(b *Builder) {
		b.cli = cli
	}
}

func WithContext(ctx context.Context) MakeBuilderOption {
	return func(b *Builder) {
		b.ctx = ctx
	}
}

func Version(version string) MakeBuilderOption {
	return func(b *Builder) {
		b.version = version
	}
}

func SourceDirectory(srcDir string) MakeBuilderOption {
	return func(b *Builder) {
		b.srcDir = srcDir
	}
}

func DestinationDirectory(destDir string) MakeBuilderOption {
	return func(b *Builder) {
		b.destDir = destDir
	}
}

func PreferredPackageDirectory(pkgDir string) MakeBuilderOption {
	return func(b *Builder) {
		b.pkgDir = pkgDir
	}
}

func MakeBuilder(opts ...MakeBuilderOption) (*Builder, error) {
	b := new(Builder)
	b.version = "latest"
	for _, opt := range opts {
		opt(b)
	}
	if b.cli == nil {
		cli, err := client.NewEnvClient()
		if err != nil {
			return nil, err
		}
		b.cli = cli
	}
	if b.ctx == nil {
		b.ctx = context.Background()
	}
	if b.destDir == "" {
		return nil, errors.New("must supply Destination Directory")
	}
	return b, nil
}

func (b *Builder) canonicalImageName() string {
	return "registry.hub.docker.com/" + b.imageName()
}

func (b *Builder) imageName() string {
	return "jsouthworth/danos-buildimage:" + b.version
}

func (b *Builder) pullEnvironment() error {
	log.Println("pulling environment", b.canonicalImageName())
	r, err := b.cli.ImagePull(
		b.ctx,
		b.canonicalImageName(),
		types.ImagePullOptions{},
	)
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, r)
	return nil
}

func (b *Builder) getBindMounts() []string {
	binds := []string{b.destDir + ":/mnt/output"}
	if b.pkgDir != "" {
		binds = append(binds, b.pkgDir+":/mnt/pkgs")
	}
	if b.srcDir != "" {
		binds = append(binds, b.srcDir+":/mnt/src")
	}
	return binds
}

func (b *Builder) genEnvironment() []string {
	out := []string{}
	if b.srcDir != "" {
		out = append(out, "DANOS_SRC_MOUNTED=1")
	}
	return out
}

func (b *Builder) createEnvironment() error {
	log.Println("creating environment", b.canonicalImageName())
	createResp, err := b.cli.ContainerCreate(
		b.ctx,
		&container.Config{
			Image:        b.canonicalImageName(),
			Env:          b.genEnvironment(),
			AttachStdout: true,
			AttachStderr: true,
		},
		&container.HostConfig{
			Binds:      b.getBindMounts(),
			Privileged: true,
		},
		nil,
		"",
	)
	if err != nil {
		return err
	}
	b.containerID = createResp.ID
	log.Println("containerID", b.containerID)
	return nil
}

func (b *Builder) buildImage() error {
	log.Println("building image", b.containerID)
	out, err := b.cli.ContainerAttach(
		b.ctx,
		b.containerID,
		types.ContainerAttachOptions{
			Stream: true,
			Stdout: true,
			Stderr: true,
		},
	)
	if err != nil {
		return err
	}
	defer out.Close()
	go stdcopy.StdCopy(os.Stdout, os.Stderr, out.Reader)

	err = b.cli.ContainerStart(
		b.ctx,
		b.containerID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		return err
	}

	_, err = b.cli.ContainerWait(b.ctx, b.containerID)
	return err
}

func (b *Builder) deleteEnvironment() error {
	if b.containerID == "" {
		return nil
	}
	log.Println("deleting environment", b.containerID)
	return b.cli.ContainerRemove(
		b.ctx,
		b.containerID,
		types.ContainerRemoveOptions{},
	)
}

func (b *Builder) Build() (err error) {
	type buildStep func() error

	defer func() {
		e := b.deleteEnvironment()
		if err == nil {
			err = e
		}
	}()

	steps := []buildStep{
		b.pullEnvironment,
		b.createEnvironment,
		b.buildImage,
	}
	for _, step := range steps {
		err := step()
		if err != nil {
			return err
		}
	}
	return nil
}
