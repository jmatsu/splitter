package exec

import (
	"context"
	"errors"
	"fmt"
	"github.com/magiconair/properties/assert"
	"k8s.io/utils/exec"
	testingexec "k8s.io/utils/exec/testing"
	"strings"
	"testing"
)

func Test_NewCommandLine(t *testing.T) {
	cases := map[string]struct {
		cmd  string
		args []string
		err  error
	}{
		"no arg": {
			cmd:  "echo",
			args: []string{},
		},
		"arg 1": {
			cmd: "echo",
			args: []string{
				"one",
			},
		},
		"arg N": {
			cmd: "echo",
			args: []string{
				"one",
				"two",
			},
		},
		"no arg err": {
			cmd:  "err",
			args: []string{},
			err:  errors.New("err is expected"),
		},
		"arg 1 err": {
			cmd: "err",
			args: []string{
				"one",
			},
			err: errors.New("err is expected"),
		},
		"arg N err": {
			cmd: "err",
			args: []string{
				"one",
				"two",
			},
			err: errors.New("err is expected"),
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			sh := &testingexec.FakeExec{
				CommandScript: []testingexec.FakeCommandAction{
					func(cmd string, args ...string) exec.Cmd {
						fakeCmd := &testingexec.FakeCmd{}
						testingexec.InitFakeCmd(fakeCmd, cmd, args...)
						fakeCmd.RunScript = []testingexec.FakeAction{
							func() ([]byte, []byte, error) {
								argstr := strings.Join(args, " ")
								return []byte(fmt.Sprintf("stdout:%s", argstr)), []byte(fmt.Sprintf("stderr:%s", argstr)), c.err
							},
						}
						return fakeCmd
					},
				},
				ExactOrder: true,
			}
			cli := NewCommandLine(context.TODO(), sh)
			stdout, stderr, err := cli.Exec(c.cmd, c.args...)

			argstr := strings.Join(c.args, " ")

			assert.Matches(t, string(stdout), fmt.Sprintf("stdout:%s", argstr))
			assert.Matches(t, string(stderr), fmt.Sprintf("stderr:%s", argstr))

			if c.err != nil {
				if err != nil {
					unwrapped := errors.Unwrap(err)
					assert.Equal(t, true, errors.Is(unwrapped, c.err))
				} else {
					t.Fatalf("err is expected but none")
				}
			} else {
				if err != nil {
					t.Fatalf("no err is expected: %v", err)
				}
			}
		})
	}
}
