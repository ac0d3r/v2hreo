package vmess

import (
	"time"
	"v2rayss/logs"
)

var (
	//NoPing no-ping
	NoPing time.Duration = -1
)

// Ping check vmess host ping response speed
func Ping(host *Host, round int, dst string) (time.Duration, error) {
	out, err := Vmess2Outbound(host, false)
	if err != nil {
		return NoPing, err
	}
	server, err := StartV2Ray(false, nil, out)
	if err != nil {
		return NoPing, err
	}
	durationList := []time.Duration{}

	defer func() {
		if err := server.Close(); err != nil {
			logs.Info(err)
		}
	}()

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(3 * time.Duration(round) * time.Second)
		timeout <- true
	}()

LOOP:
	for count := 0; count < round; count++ {
		chDelay := make(chan time.Duration)
		go func() {
			delay, err := measureDelay(server, 5*time.Second, dst)
			if err != nil {
				logs.Info(err)
			}
			chDelay <- delay
		}()

		select {
		case delay := <-chDelay:
			if delay > 0 {
				durationList = append(durationList, time.Duration(delay))
			}
		case <-timeout:
			break LOOP
		}
	}
	if len(durationList) == 0 {
		return NoPing, nil
	}
	//take the average
	return delayAverage(durationList), nil
}

func delayAverage(list []time.Duration) time.Duration {
	delay := time.Duration(0)
	for _, d := range list {
		delay += d
	}
	return time.Duration(int(delay) / len(list))
}
