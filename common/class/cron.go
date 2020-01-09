package class

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type result struct {
	output []byte
	err    error
}

func main() {
	var (
		cmd        *exec.Cmd
		ctx        context.Context
		cancelFunc context.CancelFunc
		resChan    chan *result
		res        *result
	)
	resChan = make(chan *result, 1000)
	go func() {
		var (
			output []byte
			err    error
			res    *result
		)
		ctx, cancelFunc = context.WithCancel(context.TODO())
		cmd = exec.CommandContext(ctx, "/bin/bash", "-c", "sleep 5;ls -al")
		output, err = cmd.CombinedOutput()
		res = &result{
			err:    err,
			output: output,
		}
		resChan <- res
	}()
	time.Sleep(2 * time.Second)
	cancelFunc()
	if res = <-resChan; res.err != nil {
		fmt.Println(res.err.Error())
	}
	fmt.Println(string(res.output))
}
