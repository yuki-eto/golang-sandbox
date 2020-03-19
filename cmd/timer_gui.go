package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/martinlindhe/notify"
	"github.com/zserge/webview"
)

const html = `
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
    <style>
      html, body {
        margin: 0;
        height: 100%%;
        overflow: hidden;
        -webkit-user-select: none;
      }
    </style>
  </head>
  <body>
    <div id="root">
      <div class="jumbotron text-center text-monospace" v-show="isEnabledTimer">
        <h1 class="display-4 text-monospace">{{ timer.remain }}</h1>
        <p class="lead">{{ timer.now }}</p>
      </div>

      <form v-show="!isEnabledTimer">
        <div class="form-group">
          <label for="formControlHours">Hours</label>
          <input type="range" class="custom-range" min="0" max="23" step="1" id="formControlHours" v-model="hours" @input="setDuration" />
          {{ hours }}
        </div>
        <div class="form-group">
          <label for="formControlMinutes">Minutes</label>
          <input type="range" class="custom-range" min="0" max="59" step="1" id="formControlMinutes" v-model="minutes" @input="setDuration" />
          {{ minutes }}
        </div>
        <div class="form-group">
          <label for="formControlSeconds">Seconds</label>
          <input type="range" class="custom-range" min="0" max="59" step="1" id="formControlSeconds" v-model="seconds" @input="setDuration" />
          {{ seconds }}
        </div>
      </form>

      <div class="text-center">
        <button type="button" class="btn btn-warning" v-show="isEnabledTimer && !timer.isStopped" @click="stopTimer">Stop</button>
        <button type="button" class="btn btn-danger" v-show="timer.isStarted && timer.isStopped" @click="resetTimer">Reset</button>
        <button type="button" class="btn btn-primary" v-show="!timer.isStarted || timer.isStopped" @click="startTimer">Start</button>
      </div>
    </div>


<script src="https://code.jquery.com/jquery-3.4.1.slim.min.js" integrity="sha384-J6qa4849blE2+poT4WnyKhv5vZF5SrPo0iEjwBvKU7imGFAV0wwj1yYfoRSJoZ+n" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/popper.js@1.16.0/dist/umd/popper.min.js" integrity="sha384-Q6E9RHvbIyZFJoft+2mJbHaEWldlvI9IOYy5n3zV9zzTtmI3UksdQRVvoxMfooAo" crossorigin="anonymous"></script>
<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/js/bootstrap.min.js" integrity="sha384-wfSDF2E50Y2D1uUdj0O3uMBJnjuUD4Ih7YwaYd1iqfktj0Uod8GCExl3Og8ifwB6" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/vue@2.6.11"></script>
<script>
const timer = Vue.observable(Timer.data);
const vm = new Vue({
  el: '#root',
  data: {
    timer: timer,
    hours: 0,
    minutes: 5,
    seconds: 0,
    intervalId: null,
  },
  mounted: function() {
    const self = this;
  },
  methods: {
    setDuration() {
      Timer.setDuration(parseInt(this.hours, 10), parseInt(this.minutes, 10), parseInt(this.seconds, 10));
    },
    startTimer() {
      Timer.start();
      const id = setInterval(() => {
        Timer.update();
        this.timer = Timer.data;
      }, 500);
    },
    stopTimer() {
      Timer.stop();
      if (this.intervalId !== null) {
        clearInterval(this.intervalId);
        this.intervalId = null;
      }
    },
    resetTimer() {
      Timer.reset(parseInt(this.hours, 10), parseInt(this.minutes, 10), parseInt(this.seconds, 10));
    },
  },
  computed: {
    isEnabledTimer: function() {
      return this.timer.isStarted && this.timer.remain !== '';
    },
  },
});
</script>
  </body>
</html>
`

type Timer struct {
	duration time.Duration
	h, m, s  uint64
	target   *time.Time
	timer    *time.Timer

	IsStarted bool   `json:"isStarted"`
	IsStopped bool   `json:"isStopped"`
	Now       string `json:"now"`
	Remain    string `json:"remain"`
}

func (t *Timer) SetDuration(h, m, s uint64) {
	t.h = h
	t.m = m
	t.s = s
}

func (t *Timer) Reset(h, m, s uint64) {
	t.SetDuration(h, m, s)
	t.timer = nil
	t.target = nil
	t.IsStarted = false
	t.IsStopped = false
	t.Now = ""
	t.Remain = ""
}

func (t *Timer) Start() {
	t.IsStarted = true
	t.IsStopped = false
	if t.duration == 0 {
		d := time.Duration(t.h) * time.Hour
		d += time.Duration(t.m) * time.Minute
		d += time.Duration(t.s) * time.Second
		t.duration = d
	}

	target := time.Now().Add(t.duration)
	t.target = &target
	if t.timer != nil {
		t.timer.Reset(t.duration)
		return
	}

	timer := time.NewTimer(t.duration)
	t.timer = timer
	go func() {
		<-timer.C
		notify.Alert("Timer", fmt.Sprint("duration passed"), "", "")
		t.target = nil
		t.timer = nil
		t.IsStarted = false
		t.duration = 0
	}()
}

func (t *Timer) Stop() {
	t.timer.Stop()
	now := time.Now()
	diff := t.target.Sub(now)
	t.duration = diff
	t.IsStopped = true
	t.target = nil
}

func (t *Timer) Update() {

	now := time.Now()
	t.Now = now.Format("2006/01/02 15:04:05 MST")
	if t.target == nil || now.After(*t.target) {
		return
	}

	diff := t.target.Sub(now)
	h := uint64(diff.Hours()) % 24
	m := uint64(diff.Minutes()) % 60
	s := uint64(diff.Seconds()) % 60
	t.Remain = fmt.Sprintf("Remain %02d:%02d:%02d", h, m, s)
	t.duration = diff
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run() error {
	timer := &Timer{}
	timer.SetDuration(0, 5, 0)

	w := webview.New(webview.Settings{
		Title:                  "Timer",
		URL:                    "data:text/html," + url.PathEscape(html),
		Width:                  640,
		Height:                 340,
		Resizable:              false,
		Debug:                  true,
		ExternalInvokeCallback: nil,
	})
	defer w.Exit()
	w.Dispatch(func() {
		w.Bind("Timer", timer)
	})

	w.Run()
	return nil
}
