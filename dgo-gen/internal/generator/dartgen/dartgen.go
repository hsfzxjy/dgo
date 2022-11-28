package dartgen

import (
	"os"
	"os/exec"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/config"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/generator/irgen"
)

type Generator struct{}

func (d *Generator) AddType(*exported.Type) {}

func (d *Generator) Save() {
	exception.Die(os.Chdir(config.Struct.DartProject.Path))
	cmd := exec.Command("dart", "run", "dgo_gen_dart")
	stdinPipe, err := cmd.StdinPipe()
	exception.Die(err)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Start()
	stdinPipe.Write(irgen.MarshaledPayload)
	stdinPipe.Close()
	cmd.Wait()
	if code := cmd.ProcessState.ExitCode(); code != 0 {
		os.Exit(code)
	}
}
