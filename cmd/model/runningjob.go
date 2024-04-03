package model

type RunningJob struct {
	Job
	Url string
}

func CreateRunningJob(job *Job, url string) *RunningJob {
	runningJob := RunningJob {
		Job: *job,
		Url: url,
	}

	return &runningJob
}