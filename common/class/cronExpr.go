package class

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

type job struct {
	expr string
	nextTime time.Time
}

func initCronExpr() {
	var (
		expr *cronexpr.Expression
		jobTable map[string]*job
		now time.Time
		err error
	)
	jobTable = make(map[string]*job)
	jobTable["job1"] = &job{
		expr: "* * * * * *",
	}
	//弄一个循环检查这些
	for {
		now = time.Now()
		for jobName, job := range jobTable {
			job := job
			if job.nextTime.Before(now) || job.nextTime.Equal(now) {
				go func(jobName string) {
					fmt.Printf("job: %s, time: %s\n", jobName, now.String())
				}(jobName)
				if expr, err  = cronexpr.Parse(job.expr); err != nil {
					fmt.Println(err.Error())
				}
				job.nextTime = expr.Next(now)
				fmt.Printf("nextm time : %s\n", job.nextTime.String())
			}
		}
		select {
		case <-time.NewTimer(time.Microsecond * 100).C:
		}
	}
}