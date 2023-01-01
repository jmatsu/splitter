package task

import (
	"context"
	"errors"
	"github.com/jmatsu/splitter/internal/config"
	"k8s.io/utils/exec"
	testingexec "k8s.io/utils/exec/testing"
	"strings"
	"testing"
)

func Test_Executor_Execute(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		conf          *config.ExecutionConfig
		contentResult error

		expect      error
		expectCalls int
	}{
		"pre-step only": {
			conf: &config.ExecutionConfig{
				PreSteps: [][]string{
					{
						"hello", "pre-step1",
					},
					{
						"hello", "pre-step2",
					},
				},
				PostSteps: [][]string{},
			},
			contentResult: nil,

			expect:      nil,
			expectCalls: 2,
		},
		"post-step only": {
			conf: &config.ExecutionConfig{
				PreSteps: [][]string{},
				PostSteps: [][]string{
					{
						"hello", "post-step1",
					},
					{
						"hello", "post-step2",
					},
				},
			},
			contentResult: nil,

			expect:      nil,
			expectCalls: 2,
		},
		"both": {
			conf: &config.ExecutionConfig{
				PreSteps: [][]string{
					{
						"hello", "pre-step1",
					},
					{
						"hello", "pre-step2",
					},
				},
				PostSteps: [][]string{
					{
						"hello", "post-step1",
					},
					{
						"hello", "post-step2",
					},
				},
			},
			contentResult: nil,

			expect:      nil,
			expectCalls: 4,
		},
		"pre-step fails": {
			conf: &config.ExecutionConfig{
				PreSteps: [][]string{
					{
						"err", "pre-step1",
					},
				},
				PostSteps: [][]string{
					{
						"err", "post-step1",
					},
				},
			},
			contentResult: errors.New("content fails"),

			expect:      errors.New("pre-step1"),
			expectCalls: 1,
		},
		"content fails": {
			conf: &config.ExecutionConfig{
				PreSteps: [][]string{
					{
						"hello", "pre-step1",
					},
				},
				PostSteps: [][]string{
					{
						"err", "post-step1",
					},
				},
			},
			contentResult: errors.New("content fails"),

			expect:      errors.New("content fails"),
			expectCalls: 1,
		},
		"post-step fails": {
			conf: &config.ExecutionConfig{
				PreSteps: [][]string{
					{
						"hello", "pre-step1",
					},
				},
				PostSteps: [][]string{
					{
						"err", "post-step1",
					},
				},
			},
			contentResult: nil,

			expect:      errors.New("post-step1"),
			expectCalls: 2,
		},
		"empty steps": {
			conf: &config.ExecutionConfig{
				PreSteps:  [][]string{},
				PostSteps: [][]string{},
			},

			expect:      nil,
			expectCalls: 0,
		},
		"empty conf": {
			conf: &config.ExecutionConfig{},

			expect:      nil,
			expectCalls: 0,
		},
		"zero": {
			expect:      nil,
			expectCalls: 0,
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var stubCommandActions []testingexec.FakeCommandAction

			if c.conf != nil {
				if c.conf.PreSteps != nil {
					for _, s := range c.conf.PreSteps {
						s := s
						fakeCmd := &testingexec.FakeCmd{}
						testingexec.InitFakeCmd(fakeCmd, s[0], s[1:]...)
						fakeCmd.RunScript = []testingexec.FakeAction{
							func() ([]byte, []byte, error) {
								switch s[0] {
								case "err":
									return nil, nil, errors.New(strings.Join(s, " "))
								default:
									return nil, nil, nil
								}
							},
						}
						stubCommandActions = append(stubCommandActions, func(cmd string, args ...string) exec.Cmd {
							return fakeCmd
						})
					}
				}

				if c.conf.PostSteps != nil {
					for _, s := range c.conf.PostSteps {
						s := s
						fakeCmd := &testingexec.FakeCmd{}
						testingexec.InitFakeCmd(fakeCmd, s[0], s[1:]...)
						fakeCmd.RunScript = []testingexec.FakeAction{
							func() ([]byte, []byte, error) {
								switch s[0] {
								case "err":
									return nil, nil, errors.New(strings.Join(s, " "))
								default:
									return nil, nil, nil
								}
							},
						}
						stubCommandActions = append(stubCommandActions, func(cmd string, args ...string) exec.Cmd {
							return fakeCmd
						})
					}
				}
			}

			sh := &testingexec.FakeExec{
				CommandScript: stubCommandActions,
				ExactOrder:    true,
			}

			executor := NewExecutor(sh, context.TODO(), c.conf)

			result := executor.Execute(func() error {
				return c.contentResult
			})

			if c.expect != nil && result != nil {
				if !strings.Contains(result.Error(), c.expect.Error()) {
					t.Fatalf("%v is expected but not: %v", c.expect, result)
				}
			} else if c.expect != nil {
				t.Fatalf("failure was expected but not: %v", c.expect)
			} else if result != nil {
				t.Fatalf("no failure was expected but: %v", result)
			}

			if sh.CommandCalls != c.expectCalls {
				t.Errorf("%d calls are expected but actual calls are %d", sh.CommandCalls, c.expectCalls)
			}
		})
	}
}
